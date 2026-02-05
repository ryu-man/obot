<script lang="ts">
	import {
		AdminService,
		ChatService,
		Group,
		type LaunchServerType,
		type MCPCatalogEntry,
		type MCPCatalogServer,
		type MCPServerInstance,
		type OrgUser
	} from '$lib/services';

	import {
		Captions,
		CircleFadingArrowUp,
		Ellipsis,
		GitCompare,
		LoaderCircle,
		Power,
		Router,
		Square,
		SquareCheck,
		Trash2
	} from 'lucide-svelte';
	import { formatTimeAgo } from '$lib/time';
	import { profile } from '$lib/stores';
	import DotDotDot from '../DotDotDot.svelte';
	import { onMount } from 'svelte';
	import Table from '../table/Table.svelte';
	import { ADMIN_SESSION_STORAGE } from '$lib/constants';
	import { tooltip } from '$lib/actions/tooltip.svelte';
	import Confirm from '../Confirm.svelte';
	import McpServerK8sInfo from './McpServerK8sInfo.svelte';
	import { delay, openUrl } from '$lib/utils';
	import DiffDialog from './DiffDialog.svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';

	interface Props {
		id?: string;
		entity?: 'workspace' | 'catalog';
		entry?: MCPCatalogEntry | MCPCatalogServer;
		users?: OrgUser[];
		type?: LaunchServerType;
		usedAs?: 'page' | 'tab';
	}

	let { id, entity = 'catalog', entry, users = [], type, usedAs = 'page' }: Props = $props();

	let listServerInstances = $state<Promise<MCPServerInstance[]>>();
	let listEntryServers = $state<Promise<MCPCatalogServer[]>>();

	let showConfirm = $state<
		{ type: 'multi' } | { type: 'single'; server: MCPCatalogServer } | undefined
	>();
	let showRestartConfirm = $state<
		{ server: MCPCatalogServer; onsuccess: () => void } | undefined
	>();
	let showDeleteConfirm = $state<{ server: MCPCatalogServer; onsuccess: () => void } | undefined>();
	let diffDialog = $state<ReturnType<typeof DiffDialog>>();
	let diffServer = $state<MCPCatalogServer>();
	let selected = $state<Record<string, MCPCatalogServer>>({});
	let restarting = $state(false);
	let deleting = $state(false);
	let updating = $state<Record<string, { inProgress: boolean; error: string }>>({});

	let hasSelected = $derived(Object.values(selected).some((v) => v));
	let usersMap = $derived(new Map(users.map((u) => [u.id, u])));
	let isAdminUrl = $derived(page.url.pathname.includes('/admin'));

	onMount(load);

	function load() {
		if (entry && !('isCatalogEntry' in entry) && id) {
			if (entry.catalogEntryID) {
				listServerInstances = Promise.resolve([
					{ id: entry.id, userID: entry.userID, created: entry.created }
				]);
			} else if (entity === 'workspace') {
				listServerInstances = ChatService.listWorkspaceMcpCatalogServerInstances(id, entry.id);
			} else {
				listServerInstances = AdminService.listMcpCatalogServerInstances(id, entry.id);
			}
		} else if (entry && 'isCatalogEntry' in entry && id) {
			loadEntryServers(id, entry);
		}
	}

	function loadEntryServers(id: string, entry: MCPCatalogEntry) {
		listEntryServers =
			entity === 'workspace'
				? ChatService.listWorkspaceMCPServersForEntry(id, entry.id)
				: AdminService.listMCPServersForEntry(id, entry.id);
	}

	async function handleMultiUpdate() {
		if (!id || !entry) return;
		for (const serverId of Object.keys(selected)) {
			updating[serverId] = { inProgress: true, error: '' };
			try {
				await (entity === 'workspace' && id && entry
					? ChatService.triggerWorkspaceMcpServerUpdate(id, entry.id, serverId)
					: ChatService.triggerMcpServerUpdate(serverId));
				updating[serverId] = { inProgress: false, error: '' };
			} catch (error) {
				updating[serverId] = {
					inProgress: false,
					error: error instanceof Error ? error.message : 'An unknown error occurred'
				};
			} finally {
				delete updating[serverId];
			}
		}

		load();
		selected = {};
	}

	async function updateServer(server?: MCPCatalogServer) {
		if (!id || !entry || !server) return;

		updating[server.id] = { inProgress: true, error: '' };
		try {
			await (entity === 'workspace' && id && entry
				? ChatService.triggerWorkspaceMcpServerUpdate(id, entry.id, server.id)
				: ChatService.triggerMcpServerUpdate(server.id));

			load();
		} catch (err) {
			updating[server.id] = {
				inProgress: false,
				error: err instanceof Error ? err.message : 'An unknown error occurred'
			};
		}

		delete updating[server.id];
	}

	function setLastVisitedMcpServer() {
		if (!entry) return;
		const name = entry.manifest?.name;
		sessionStorage.setItem(
			ADMIN_SESSION_STORAGE.LAST_VISITED_MCP_SERVER,
			JSON.stringify({ id: entry.id, name, type, entity, entityId: id })
		);
	}

	function getAuditLogUrl(d: MCPCatalogServer) {
		if (isAdminUrl) {
			if (!profile.current?.hasAdminAccess?.()) return null;
			return entity === 'workspace'
				? `/admin/mcp-servers/w/${id}/c/${entry?.id}?view=audit-logs&mcp_id=${d.id}&user_id=${d.userID}`
				: `/admin/mcp-servers/c/${entry?.id}?view=audit-logs&mcp_id=${d.id}&user_id=${d.userID}`;
		}

		return profile.current?.groups.includes(Group.POWERUSER)
			? `/mcp-servers/c/${entry?.id}?view=audit-logs&mcp_id=${d.id}&user_id=${d.userID}`
			: null;
	}

	async function handleRestartServer(server: MCPCatalogServer) {
		if (!id || !entry) return;

		restarting = true;
		try {
			if (entity === 'workspace') {
				// Check if we're in a catalog entry context
				if ('isCatalogEntry' in entry) {
					// Restart server from a catalog entry in a workspace
					await ChatService.restartWorkspaceCatalogEntryServerDeployment(
						id, // workspace ID
						entry.id, // catalog entry ID
						server.id // server deployment ID
					);
				} else {
					// Restart a direct multi-user server in a workspace
					await ChatService.restartWorkspaceK8sServerDeployment(
						id, // workspace ID
						server.id // server deployment ID
					);
				}
			} else {
				// For global catalog, restart the deployed server
				await AdminService.restartK8sDeployment(server.id);
			}
		} catch (err) {
			console.error('Failed to restart server:', err);
		} finally {
			restarting = false;
			showRestartConfirm = undefined;
		}
	}

	async function handleDeleteServer(server: MCPCatalogServer) {
		if (!id || !entry || !('isCatalogEntry' in entry)) return;

		deleting = true;
		try {
			if (entity === 'workspace') {
				// Delete server from workspace
				await ChatService.deleteWorkspaceMCPCatalogServer(
					id, // workspace ID
					server.id // server ID
				);
			} else {
				// Delete server from global catalog
				// Note: We're deleting a deployed server instance, not the catalog entry
				await ChatService.deleteSingleOrRemoteMcpServer(server.id);
			}

			// Small delay to allow backend to process the deletion
			await delay(500);

			// Refresh the list - create a new promise to trigger reactivity
			loadEntryServers(id, entry);
		} catch (err) {
			console.error('Failed to delete server:', err);
		} finally {
			deleting = false;
			showDeleteConfirm = undefined;
		}
	}
</script>

{#if listServerInstances}
	{#await listServerInstances}
		<div class="flex w-full justify-center">
			<LoaderCircle class="size-6 animate-spin" />
		</div>
	{:then instances}
		{#if entry && (type === 'multi' || instances.length > 0)}
			<div class="flex flex-col gap-6">
				<McpServerK8sInfo
					{id}
					{entity}
					mcpServerId={entry.id}
					name={'manifest' in entry ? entry.manifest.name || '' : ''}
					connectedUsers={instances.map((instance) => {
						const user = usersMap.get(instance.userID)!;
						return {
							...user,
							mcpInstanceId: instance.id
						};
					})}
					title="Details"
					classes={{
						title: 'text-lg font-semibold'
					}}
					readonly={profile.current.isAdminReadonly?.()}
					{usedAs}
				/>
			</div>
		{:else}
			{@render emptyInstancesContent()}
		{/if}
	{/await}
{:else if listEntryServers}
	{#await listEntryServers}
		<div class="flex w-full justify-center">
			<LoaderCircle class="size-6 animate-spin" />
		</div>
	{:then servers}
		{@const numServerUpdatesNeeded = servers.filter((s) => s.needsUpdate).length}
		{#if servers.length > 0}
			{#if numServerUpdatesNeeded}
				<button
					class="bg-background group mb-2 w-fit rounded-md"
					onclick={() => {
						// TODO: show all servers with upgrade & update all option
					}}
				>
					<div
						class="border-primary bg-primary/10 group-hover:bg-primary/20 dark:bg-primary/30 dark:group-hover:bg-primary/40 flex items-center gap-1 rounded-md border px-4 py-2 transition-colors duration-300"
					>
						<CircleFadingArrowUp class="text-primary size-4" />
						<p class="text-primary text-sm font-light">
							{#if numServerUpdatesNeeded === 1}
								1 instance has an update available.
							{:else}
								{numServerUpdatesNeeded} instances have updates available.
							{/if}
						</p>
					</div>
				</button>
			{/if}
			<Table
				data={servers}
				fields={type === 'single' || type === 'composite'
					? ['userID', 'created']
					: ['url', 'userID', 'created']}
				headers={[
					{ title: 'User', property: 'userID' },
					{ title: 'URL', property: 'url' }
				]}
				onClickRow={type === 'single' || type === 'composite' || type === 'remote'
					? (d, isCtrlClick) => {
							setLastVisitedMcpServer();

							const url =
								entity === 'workspace'
									? isAdminUrl
										? `/admin/mcp-servers/w/${id}/c/${entry?.id}/instance/${d.id}/details`
										: `/mcp-servers/c/${entry?.id}/instance/${d.id}/details`
									: `/admin/mcp-servers/c/${entry?.id}/instance/${d.id}/details`;
							openUrl(url, isCtrlClick);
						}
					: undefined}
			>
				{#snippet onRenderColumn(property, d)}
					{#if property === 'url'}
						<span class="flex items-center gap-1">
							{d.manifest.remoteConfig?.url}
							{#if d.needsUpdate}
								<div
									use:tooltip={{
										text: 'This server needs an update. View Diff to see the changes.',
										classes: ['break-words', 'w-58']
									}}
								>
									<CircleFadingArrowUp class="text-primary size-4" />
								</div>
							{/if}
						</span>
					{:else if property === 'userID'}
						{@const user = usersMap.get(d[property] as string)}
						<span class="flex items-center gap-1">
							{#if users.length === 0}
								<!--This covers the case where a Power User is listing their own servers.-->
								{profile.current.email || 'Unknown'}
							{:else}
								{user?.email || user?.username || 'Unknown'}
							{/if}
							{#if type === 'single' || type === 'composite'}
								{#if d.needsUpdate}
									<div
										use:tooltip={{
											text: 'This server needs an update. View Diff to see the changes.',
											classes: ['break-words', 'w-58']
										}}
									>
										<CircleFadingArrowUp class="text-primary size-4" />
									</div>
								{/if}
							{/if}
						</span>
					{:else if property === 'created'}
						{formatTimeAgo(d[property] as unknown as string).fullDate}
					{:else}
						{d[property as keyof typeof d]}
					{/if}
				{/snippet}

				{#snippet actions(d)}
					{@const auditLogsUrl = getAuditLogUrl(d)}

					<!-- Check for permissions -->
					{@const isInstanceOwner = d.userID === profile.current.id}
					{@const hasAdminAccess = profile.current.hasAdminAccess?.()}
					{@const canRestart =
						hasAdminAccess ||
						(entity === 'workspace' && isInstanceOwner) ||
						(type === 'single' && isInstanceOwner)}
					{@const canDelete = hasAdminAccess || isInstanceOwner}

					<div class="flex items-center gap-1">
						<DotDotDot class="icon-button hover:dark:bg-background/50" classes={{ menu: 'gap-1' }}>
							{#snippet icon()}
								<Ellipsis class="size-4" />
							{/snippet}

							{#snippet children({ toggle })}
								{#if auditLogsUrl}
									<a class="menu-button" href={resolve(auditLogsUrl as `/${string}`)}>
										<Captions class="size-4" /> View Audit Logs
									</a>
								{/if}

								{#if canRestart}
									<button
										class="menu-button"
										disabled={restarting}
										onclick={(e) => {
											e.stopPropagation();
											showRestartConfirm = {
												server: d,
												onsuccess: () => {
													toggle(false);
												}
											};
										}}
									>
										{#if restarting}
											<LoaderCircle class="size-4 animate-spin" />
										{:else}
											<Power class="size-4" />
										{/if}
										Restart Server
									</button>
								{/if}

								{#if canDelete}
									<button
										class="menu-button"
										disabled={deleting}
										onclick={(e) => {
											e.stopPropagation();
											showDeleteConfirm = {
												server: d,
												onsuccess: () => {
													toggle(false);
												}
											};
										}}
									>
										{#if deleting}
											<LoaderCircle class="size-4 animate-spin" />
										{:else}
											<Trash2 class="size-4" />
										{/if}
										Delete Server
									</button>
								{/if}

								{#if d.needsUpdate}
									<button
										class="menu-button bg-primary/10 text-primary hover:bg-primary/20"
										disabled={updating[d.id]?.inProgress || !!d.compositeName}
										onclick={async (e) => {
											e.stopPropagation();
											showConfirm = {
												type: 'single',
												server: d
											};
										}}
										use:tooltip={d.compositeName
											? {
													text: 'Cannot directly update a descendant of a composite server; update the composite MCP server instead.',
													classes: ['w-md'],
													disablePortal: true
												}
											: undefined}
									>
										{#if updating[d.id]?.inProgress}
											<LoaderCircle class="size-4 animate-spin" />
										{:else}
											<CircleFadingArrowUp class="size-4" />
										{/if}
										Update Server
									</button>

									<button
										class="menu-button"
										onclick={(e) => {
											e.stopPropagation();
											diffServer = d;
											diffDialog?.open();
										}}
									>
										<GitCompare class="size-4" /> View Diff
									</button>
								{/if}
							{/snippet}
						</DotDotDot>

						{#if d.needsUpdate}
							<button
								class="icon-button hover:bg-black/50"
								onclick={(e) => {
									e.stopPropagation();
									if (selected[d.id]) {
										delete selected[d.id];
									} else {
										selected[d.id] = d;
									}
								}}
							>
								{#if selected[d.id]}
									<SquareCheck class="size-5" />
								{:else}
									<Square class="size-5" />
								{/if}
							</button>
						{:else if numServerUpdatesNeeded > 0}
							<div class="size-10"></div>
							<div class="size-10"></div>
						{/if}
					</div>
				{/snippet}
			</Table>

			{#if hasSelected}
				{@const numSelected = Object.keys(selected).length}
				{@const updatingInProgress = Object.values(updating).some((u) => u.inProgress)}
				<div
					class="bg-surface1 dark:bg-background sticky bottom-0 left-0 mt-auto flex w-[calc(100%+2em)] -translate-x-4 justify-end gap-4 p-4 md:w-[calc(100%+4em)] md:-translate-x-8 md:px-8"
				>
					<div class="flex w-full items-center justify-between">
						<p class="text-sm font-medium">
							{numSelected} server instance{numSelected === 1 ? '' : 's'} selected
						</p>
						<div class="flex items-center gap-4">
							<button
								class="button flex items-center gap-1"
								onclick={() => {
									selected = {};
									updating = {};
								}}
							>
								Cancel
							</button>
							<button
								class="button-primary flex items-center gap-1"
								onclick={() => {
									showConfirm = {
										type: 'multi'
									};
								}}
								disabled={updatingInProgress}
							>
								{#if updatingInProgress}
									<LoaderCircle class="size-5" />
								{:else}
									Update Servers
								{/if}
							</button>
						</div>
					</div>
				</div>
			{/if}
		{:else}
			{@render emptyInstancesContent()}
		{/if}
	{/await}
{:else}
	{@render emptyInstancesContent()}
{/if}

<DiffDialog bind:this={diffDialog} fromServer={diffServer} toServer={entry} />

{#snippet emptyInstancesContent()}
	<div class="mt-12 flex w-md flex-col items-center gap-4 self-center text-center">
		<Router class="text-on-surface1 size-24 opacity-50" />
		<h4 class="text-on-surface1 text-lg font-semibold">No server details</h4>
		<p class="text-on-surface1 text-sm font-light">No details available yet for this server.</p>
	</div>
{/snippet}

<Confirm
	show={!!showRestartConfirm}
	onsuccess={async () => {
		if (!showRestartConfirm) return;
		await handleRestartServer(showRestartConfirm.server);
		showRestartConfirm.onsuccess();

		showRestartConfirm = undefined;
	}}
	oncancel={() => (showRestartConfirm = undefined)}
	loading={restarting}
	msg={`Restart server instance created by ${usersMap.get(showRestartConfirm?.server.userID || '')?.email || usersMap.get(showRestartConfirm?.server.userID || '')?.username || profile.current.email || 'Unknown'}?`}
	type="info"
	title="Confirm Restart"
>
	{#snippet note()}
		The server deployment will be restarted. This may cause temporary interruption for the user.
	{/snippet}
</Confirm>

<Confirm
	show={!!showDeleteConfirm}
	onsuccess={async () => {
		if (!showDeleteConfirm) return;
		await handleDeleteServer(showDeleteConfirm.server);
		showDeleteConfirm.onsuccess();

		showDeleteConfirm = undefined;
	}}
	oncancel={() => (showDeleteConfirm = undefined)}
	loading={deleting}
	msg={`Delete server instance created by ${usersMap.get(showDeleteConfirm?.server.userID || '')?.email || usersMap.get(showDeleteConfirm?.server.userID || '')?.username || profile.current.email || 'Unknown'}?`}
	type="delete"
	title="Confirm Delete"
>
	{#snippet note()}
		This will permanently delete the server instance. This action cannot be undone.
	{/snippet}
</Confirm>

<Confirm
	show={!!showConfirm}
	onsuccess={async () => {
		if (!showConfirm) return;
		if (showConfirm.type === 'single') {
			await updateServer(showConfirm.server);
		} else {
			await handleMultiUpdate();
		}
		showConfirm = undefined;
	}}
	oncancel={() => (showConfirm = undefined)}
	classes={{
		confirm: 'bg-primary hover:bg-primary/50 transition-colors duration-200'
	}}
	msg={`Update ${showConfirm?.type === 'single' ? showConfirm.server.id : 'selected server(s)'}?`}
	type="info"
	title="Confirm Update"
>
	{#snippet note()}
		If this update introduces new required configuration parameters, users will have to supply them
		before they can use {showConfirm?.type === 'multi' ? 'these servers' : 'this server'} again.
	{/snippet}
</Confirm>
