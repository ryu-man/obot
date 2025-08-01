package mcpcatalog

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/obot-platform/nah/pkg/apply"
	"github.com/obot-platform/nah/pkg/name"
	"github.com/obot-platform/nah/pkg/router"
	"github.com/obot-platform/obot/apiclient/types"
	"github.com/obot-platform/obot/logger"
	"github.com/obot-platform/obot/pkg/accesscontrolrule"
	gclient "github.com/obot-platform/obot/pkg/gateway/client"
	v1 "github.com/obot-platform/obot/pkg/storage/apis/obot.obot.ai/v1"
	"github.com/obot-platform/obot/pkg/system"
	"github.com/obot-platform/obot/pkg/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

var log = logger.Package()

const (
	// These are used to force catalog sync on startup, used for times when changes are made to
	// catalogs, and they must be synced on the next start.
	forceSyncStartupAnnotation = "obot.ai/force-sync-startup"
	// Bump this any time this functionality is needed.
	startupSyncGeneration = "1"
)

type Handler struct {
	defaultCatalogPath      string
	gatewayClient           *gclient.Client
	accessControlRuleHelper *accesscontrolrule.Helper
}

func New(defaultCatalogPath string, gatewayClient *gclient.Client, accessControlRuleHelper *accesscontrolrule.Helper) *Handler {
	return &Handler{
		defaultCatalogPath:      defaultCatalogPath,
		gatewayClient:           gatewayClient,
		accessControlRuleHelper: accessControlRuleHelper,
	}
}

func (h *Handler) Sync(req router.Request, resp router.Response) error {
	mcpCatalog := req.Object.(*v1.MCPCatalog)

	forceSync := mcpCatalog.Annotations[v1.MCPCatalogSyncAnnotation] == "true" || mcpCatalog.Annotations[forceSyncStartupAnnotation] != startupSyncGeneration
	if !forceSync && !mcpCatalog.Status.LastSyncTime.IsZero() {
		timeSinceLastSync := time.Since(mcpCatalog.Status.LastSyncTime.Time)
		if timeSinceLastSync < time.Hour {
			resp.RetryAfter(time.Hour - timeSinceLastSync)
			return nil
		}
	}

	toAdd := make([]client.Object, 0)
	mcpCatalog.Status.SyncErrors = make(map[string]string)

	for _, sourceURL := range mcpCatalog.Spec.SourceURLs {
		objs, err := h.readMCPCatalog(mcpCatalog.Name, sourceURL)
		if err != nil {
			log.Errorf("failed to read catalog %s: %v", sourceURL, err)
			mcpCatalog.Status.SyncErrors[sourceURL] = err.Error()
			continue
		}
		delete(mcpCatalog.Status.SyncErrors, sourceURL)

		toAdd = append(toAdd, objs...)
	}

	mcpCatalog.Status.LastSyncTime = metav1.Now()
	if err := req.Client.Status().Update(req.Ctx, mcpCatalog); err != nil {
		return fmt.Errorf("failed to update catalog status: %w", err)
	}
	if forceSync {
		delete(mcpCatalog.Annotations, v1.MCPCatalogSyncAnnotation)
		if mcpCatalog.Annotations == nil {
			mcpCatalog.Annotations = make(map[string]string, 1)
		}
		mcpCatalog.Annotations[forceSyncStartupAnnotation] = startupSyncGeneration
		if err := req.Client.Update(req.Ctx, mcpCatalog); err != nil {
			return fmt.Errorf("failed to update catalog: %w", err)
		}
	}

	// We want to refresh this every hour.
	// TODO(g-linville): make this configurable.
	resp.RetryAfter(time.Hour)

	// Don't run apply if there are sync errors
	if len(mcpCatalog.Status.SyncErrors) > 0 {
		return nil
	}

	// I know we don't want to do apply anymore. But we were doing it before in a different place.
	// Now we're doing it here. It's not important enough to change right now.
	return apply.New(req.Client).WithOwnerSubContext(fmt.Sprintf("catalog-%s", mcpCatalog.Name)).
		WithPruneTypes(&v1.MCPServerCatalogEntry{}).Apply(req.Ctx, mcpCatalog, toAdd...)
}

func (h *Handler) readMCPCatalog(catalogName, sourceURL string) ([]client.Object, error) {
	var entries []types.MCPServerCatalogEntryManifest

	if strings.HasPrefix(sourceURL, "http://") || strings.HasPrefix(sourceURL, "https://") {
		if isGitHubURL(sourceURL) {
			var err error
			entries, err = readGitHubCatalog(sourceURL)
			if err != nil {
				return nil, fmt.Errorf("failed to read GitHub catalog %s: %w", sourceURL, err)
			}
		} else {
			// If it wasn't a GitHub repo, treat it as a raw file.
			resp, err := http.Get(sourceURL)
			if err != nil {
				return nil, fmt.Errorf("failed to read catalog %s: %w", sourceURL, err)
			}
			defer resp.Body.Close()

			contents, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read catalog %s: %w", sourceURL, err)
			}

			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("unexpected status when reading catalog %s: %s", sourceURL, string(contents))
			}

			if err = yaml.Unmarshal(contents, &entries); err != nil {
				return nil, fmt.Errorf("failed to decode catalog %s: %w", sourceURL, err)
			}
		}
	} else {
		fileInfo, err := os.Stat(sourceURL)
		if err != nil {
			return nil, fmt.Errorf("failed to stat catalog %s: %w", sourceURL, err)
		}

		if fileInfo.IsDir() {
			entries, err = h.readMCPCatalogDirectory(sourceURL)
			if err != nil {
				return nil, fmt.Errorf("failed to read catalog %s: %w", sourceURL, err)
			}
		} else {
			contents, err := os.ReadFile(sourceURL)
			if err != nil {
				return nil, fmt.Errorf("failed to read catalog %s: %w", sourceURL, err)
			}

			if err = yaml.Unmarshal(contents, &entries); err != nil {
				return nil, fmt.Errorf("failed to decode catalog %s: %w", sourceURL, err)
			}
		}
	}

	objs := make([]client.Object, 0, len(entries))

	for _, entry := range entries {
		if entry.Metadata["categories"] == "Official" {
			delete(entry.Metadata, "categories") // This shouldn't happen, but do this just in case.
			// We don't want to mark random MCP servers from the catalog as official.
		}

		cleanName := strings.ToLower(strings.ReplaceAll(entry.Name, " ", "-"))

		catalogEntry := v1.MCPServerCatalogEntry{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name.SafeHashConcatName(catalogName, cleanName),
				Namespace: system.DefaultNamespace,
			},
			Spec: v1.MCPServerCatalogEntrySpec{
				MCPCatalogName: catalogName,
				SourceURL:      sourceURL,
				Editable:       false, // entries from source URLs are not editable
			},
		}

		// Check the metadata for default disabled tools.
		if entry.Metadata["unsupportedTools"] != "" {
			catalogEntry.Spec.UnsupportedTools = strings.Split(entry.Metadata["unsupportedTools"], ",")
		}

		// Sanitize the environment variables
		for i, env := range entry.Env {
			if env.Key == "" {
				env.Key = env.Name
			}

			if filepath.Ext(env.Key) != "" {
				env.Key = strings.ReplaceAll(env.Key, ".", "_")
				env.File = true
			}

			env.Key = strings.ReplaceAll(strings.ToUpper(env.Key), "-", "_")

			entry.Env[i] = env
		}

		// Sanitize the headers
		if entry.Runtime == types.RuntimeRemote && entry.RemoteConfig != nil {
			for i, header := range entry.RemoteConfig.Headers {
				if header.Key == "" {
					header.Key = header.Name
				}

				header.Key = strings.ReplaceAll(strings.ToUpper(header.Key), "_", "-")
				entry.RemoteConfig.Headers[i] = header
			}
		}

		if err := validation.ValidateCatalogEntryManifest(entry); err != nil {
			return nil, fmt.Errorf("failed to validate catalog entry %s: %w", entry.Name, err)
		}
		catalogEntry.Spec.Manifest = entry

		objs = append(objs, &catalogEntry)
	}

	return objs, nil
}

func (h *Handler) readMCPCatalogDirectory(catalog string) ([]types.MCPServerCatalogEntryManifest, error) {
	files, err := os.ReadDir(catalog)
	if err != nil {
		return nil, fmt.Errorf("failed to read catalog directory %s: %w", catalog, err)
	}

	var entries []types.MCPServerCatalogEntryManifest
	for _, file := range files {
		if file.IsDir() {
			nestedEntries, err := h.readMCPCatalogDirectory(filepath.Join(catalog, file.Name()))
			if err != nil {
				return nil, fmt.Errorf("failed to read nested catalog directory %s: %w", file.Name(), err)
			}
			entries = append(entries, nestedEntries...)
		} else if strings.HasSuffix(file.Name(), ".json") || strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml") {
			contents, err := os.ReadFile(filepath.Join(catalog, file.Name()))
			if err != nil {
				return nil, fmt.Errorf("failed to read catalog file %s: %w", file.Name(), err)
			}

			var entry types.MCPServerCatalogEntryManifest
			if err = yaml.Unmarshal(contents, &entry); err != nil {
				return nil, fmt.Errorf("failed to decode catalog file %s: %w", file.Name(), err)
			}
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

func (h *Handler) SetUpDefaultMCPCatalog(ctx context.Context, c client.Client) error {
	var existing v1.MCPCatalog
	if err := c.Get(ctx, router.Key(system.DefaultNamespace, system.DefaultCatalog), &existing); err == nil {
		// TODO: Remove this migration logic once we've migrated all Obot deployments to the new catalog path.
		if i := slices.IndexFunc(existing.Spec.SourceURLs, func(url string) bool {
			matched, _ := regexp.MatchString(`^(\./)?/?catalog$`, url)
			return matched
		}); i >= 0 {
			existing.Spec.SourceURLs[i] = h.defaultCatalogPath
			if err := c.Update(ctx, &existing); err != nil {
				return fmt.Errorf("failed to migrate default catalog: %w", err)
			}
		}

		return nil
	}

	sourceURLs := []string{}
	if h.defaultCatalogPath != "" {
		sourceURLs = append(sourceURLs, h.defaultCatalogPath)
	}

	if err := c.Create(ctx, &v1.MCPCatalog{
		ObjectMeta: metav1.ObjectMeta{
			Name:      system.DefaultCatalog,
			Namespace: system.DefaultNamespace,
		},
		Spec: v1.MCPCatalogSpec{
			DisplayName: "Default",
			SourceURLs:  sourceURLs,
		},
	}); err != nil {
		return fmt.Errorf("failed to create default catalog: %w", err)
	}

	return nil
}

// DeleteUnauthorizedMCPServers is a handler that deletes MCP servers that are no longer authorized to exist.
// This can happen whenever AccessControlRules change.
// It does not delete MCPServerInstances, since those have a delete ref to their MCPServer, and will be deleted automatically.
func (h *Handler) DeleteUnauthorizedMCPServers(req router.Request, _ router.Response) error {
	// List AccessControlRules so that this handler gets triggered any time one of them changes.
	if err := req.List(&v1.AccessControlRuleList{}, &client.ListOptions{
		Namespace: system.DefaultNamespace,
	}); err != nil {
		return fmt.Errorf("failed to list access control rules: %w", err)
	}

	var mcpServers v1.MCPServerList
	if err := req.List(&mcpServers, &client.ListOptions{
		Namespace: system.DefaultNamespace,
	}); err != nil {
		return fmt.Errorf("failed to list MCP servers: %w", err)
	}

	// Iterate through each MCPServer and make sure it is still allowed to exist.
	for _, server := range mcpServers.Items {
		if server.Spec.ThreadName != "" || server.Spec.SharedWithinMCPCatalogName != "" {
			// For project-scoped servers and multi-user servers created by the admin, we don't need to check them.
			continue
		}

		user, err := h.gatewayClient.UserByID(req.Ctx, server.Spec.UserID)
		if err != nil {
			return fmt.Errorf("failed to get user %s: %w", server.Spec.UserID, err)
		}

		if user.Role.HasRole(types.RoleAdmin) {
			// Don't delete servers created by admins.
			continue
		}

		if server.Spec.MCPServerCatalogEntryName == "" {
			// If the server doesn't have a catalog entry name, that's bad, because it should. Delete it.
			log.Infof("Deleting MCP server %q because it does not correspond to a catalog entry", server.Name)
			if err := req.Delete(&server); err != nil {
				return fmt.Errorf("failed to delete MCP server %s: %w", server.Name, err)
			}
			continue
		}

		hasAccess, err := h.accessControlRuleHelper.UserHasAccessToMCPServerCatalogEntry(server.Spec.UserID, server.Spec.MCPServerCatalogEntryName)
		if err != nil {
			return fmt.Errorf("failed to check if user %s has access to catalog entry %s: %w", server.Spec.UserID, server.Spec.MCPServerCatalogEntryName, err)
		}

		if !hasAccess {
			log.Infof("Deleting MCP server %q because it is no longer authorized to exist", server.Name)
			if err := req.Delete(&server); err != nil {
				return fmt.Errorf("failed to delete MCP server %s: %w", server.Name, err)
			}
		}
	}

	return nil
}

// DeleteUnauthorizedMCPServerInstances is a handler that deletes MCPServerInstances that point to multi-user MCPServers created by the admin,
// where the user who owns the MCPServerInstance is no longer authorized to use the MCPServer.
// This can happen whenever AccessControlRules change.
func (h *Handler) DeleteUnauthorizedMCPServerInstances(req router.Request, _ router.Response) error {
	// List AccessControlRules so that this handler gets triggered any time one of them changes.
	if err := req.List(&v1.AccessControlRuleList{}, &client.ListOptions{
		Namespace: system.DefaultNamespace,
	}); err != nil {
		return fmt.Errorf("failed to list access control rules: %w", err)
	}

	var mcpServerInstances v1.MCPServerInstanceList
	if err := req.List(&mcpServerInstances, &client.ListOptions{
		Namespace: system.DefaultNamespace,
	}); err != nil {
		return fmt.Errorf("failed to list MCP server instances: %w", err)
	}

	// Iterate through each MCPServerInstance and make sure it is still allowed to exist.
	for _, instance := range mcpServerInstances.Items {
		if !instance.DeletionTimestamp.IsZero() {
			continue
		}

		user, err := h.gatewayClient.UserByID(req.Ctx, instance.Spec.UserID)
		if err != nil {
			return fmt.Errorf("failed to get user %s: %w", instance.Spec.UserID, err)
		}

		if user.Role.HasRole(types.RoleAdmin) {
			// Don't delete instances created by admins.
			continue
		}

		hasAccess, err := h.accessControlRuleHelper.UserHasAccessToMCPServer(instance.Spec.UserID, instance.Spec.MCPServerName)
		if err != nil {
			return fmt.Errorf("failed to check if user %s has access to MCP server %s: %w", instance.Spec.UserID, instance.Spec.MCPServerName, err)
		}

		if !hasAccess {
			log.Infof("Deleting MCPServerInstance %q because it is no longer authorized to exist", instance.Name)
			if err := req.Delete(&instance); err != nil {
				return fmt.Errorf("failed to delete MCPServerInstance %s: %w", instance.Name, err)
			}
		}
	}

	return nil
}
