<script lang="ts">
	import { page } from '$app/state';
	import {
		AdminService,
		ChatService,
		type MCPCatalogEntry,
		type MCPCatalogServer,
		type OrgUser
	} from '$lib/services';
	import { profile } from '$lib/stores';
	import { twMerge } from 'tailwind-merge';
	import Table from '../table/Table.svelte';
	import { onMount } from 'svelte';
	import { AlertCircle, ChevronRight, Server, Trash2 } from 'lucide-svelte';
	import { ADMIN_SESSION_STORAGE, DEFAULT_MCP_CATALOG_ID } from '$lib/constants';
	import { delay, openUrl } from '$lib/utils';
	import { resolve } from '$app/paths';
	import Confirm from '../Confirm.svelte';
	import { goto } from '$lib/url';

	interface Props {
		entity?: 'workspace' | 'catalog';
		entityId?: string;
		catalogEntry?: MCPCatalogEntry;
		mcpServerId?: string;
		mcpServerInstanceId?: string;
		classes?: {
			title?: string;
		};
		name: string;
		connectedUsers: OrgUser[];
	}

	let { name, connectedUsers, classes, entityId, catalogEntry, mcpServerId }: Props = $props();
	let isAdminUrl = $derived(page.url.pathname.includes('/admin'));
	let servers = $state<MCPCatalogServer[]>([]);
	let serversMap = $derived(new Map(servers.map((s) => [s.catalogEntryID || s.id, s])));
	let deleting = $state(false);
	let showDeleteInstanceConfirm = $state(false);

	const createdByUser = $derived(connectedUsers[0]);
	const isServerCreatedByCurrentUser = $derived(createdByUser?.id === profile.current.id);

	onMount(async () => {
		if (!mcpServerId || !catalogEntry?.id || !entityId) return;

		const deployedCatalogEntryServers =
			await AdminService.listAllCatalogDeployedSingleRemoteServers(DEFAULT_MCP_CATALOG_ID);
		const deployedWorkspaceCatalogEntryServers =
			await AdminService.listAllWorkspaceDeployedSingleRemoteServers();

		servers = [
			...deployedCatalogEntryServers.filter((s) => s.compositeName === mcpServerId),
			...deployedWorkspaceCatalogEntryServers.filter((s) => s.compositeName === mcpServerId)
		];
	});

	function getAuditLogUrl(d: OrgUser) {
		if (!catalogEntry?.id) return null;
		if (!isAdminUrl) return null;
		if (!profile.current?.hasAdminAccess?.()) return null;
		return `/admin/mcp-servers/c/${catalogEntry.id}?view=audit-logs&user_id=${d.id}`;
	}

	async function handleDeleteInstance(serverId: typeof mcpServerId) {
		if (!serverId) return;

		deleting = true;
		try {
			await ChatService.deleteSingleOrRemoteMcpServer(serverId);

			// Small delay to allow backend to process the deletion
			await delay(500);

			//
			goto('/admin/mcp-servers');
		} catch (err) {
			console.error('Failed to delete instance:', err);
			// Remove from deleted set if deletion failed
		} finally {
			deleting = false;
			showDeleteInstanceConfirm = false;
		}
	}
</script>

<div class="flex items-center gap-3">
	<div class="flex w-full items-center justify-between">
		<h1 class={twMerge('text-2xl font-semibold', classes?.title)}>
			{name}
		</h1>

		<div class="flex gap-2">
			<button
				onclick={() => {
					if (deleting) return;
					if (!mcpServerId) return;

					showDeleteInstanceConfirm = true;
				}}
				class="button-destructive flex items-center gap-2 rounded-md px-3 py-1.5 text-xs font-medium disabled:opacity-50"
				disabled={deleting}
			>
				<Trash2 class="size-3" />
				Delete Server
			</button>
		</div>
	</div>
</div>

{#if catalogEntry?.manifest.compositeConfig?.componentServers}
	<div>
		<h2 class="mb-2 text-lg font-semibold">MCP Servers</h2>
		<div class="flex flex-col gap-2">
			{#each catalogEntry.manifest.compositeConfig.componentServers as componentServer (componentServer.catalogEntryID || componentServer.mcpServerID)}
				{@const catalogEntryServerId =
					componentServer.catalogEntryID && serversMap.get(componentServer.catalogEntryID)?.id}
				{@const mcpServerId =
					componentServer.mcpServerID && serversMap.get(componentServer.mcpServerID)?.id}
				{@const componentExists = !!(catalogEntryServerId || mcpServerId)}

				{#if componentExists}
					<button
						onclick={(e) => {
							const isCtrlClick = e.metaKey || e.ctrlKey;
							const url = componentServer.catalogEntryID
								? `/admin/mcp-servers/c/${componentServer.catalogEntryID}/instance/${catalogEntryServerId}`
								: `/admin/mcp-servers/s/${componentServer.mcpServerID}/details`;

							sessionStorage.setItem(
								ADMIN_SESSION_STORAGE.LAST_VISITED_MCP_SERVER,
								JSON.stringify({
									id: catalogEntry?.id,
									name,
									type: 'composite',
									entity: 'catalog',
									entityId: DEFAULT_MCP_CATALOG_ID,
									serverId: mcpServerId
								})
							);

							openUrl(url, isCtrlClick);
						}}
						class="dark:bg-surface1 dark:border-surface3 dark:hover:bg-surface2 bg-background flex items-center justify-between gap-2 rounded-lg border border-transparent p-2 pl-4 shadow-sm hover:bg-gray-50"
					>
						<div class="flex items-center gap-2">
							<div class="icon">
								{#if componentServer.manifest?.icon}
									<img
										src={componentServer.manifest?.icon}
										alt={componentServer.manifest?.name}
										class="size-6"
									/>
								{:else}
									<Server class="size-6" />
								{/if}
							</div>
							<p class="text-sm">{componentServer.manifest?.name}</p>
							{#if catalogEntryServerId}
								<span class="text-on-surface1 text-sm">({catalogEntryServerId})</span>
							{/if}
						</div>
						<div class="icon-button">
							<ChevronRight class="size-6" />
						</div>
					</button>
				{:else}
					<div
						class="dark:bg-surface1 dark:border-surface3 bg-background flex items-center justify-between gap-2 rounded-lg border border-transparent p-2 pl-4 opacity-60 shadow-sm"
					>
						<div class="flex items-center gap-2">
							<div class="icon">
								{#if componentServer.manifest?.icon}
									<img
										src={componentServer.manifest?.icon}
										alt={componentServer.manifest?.name}
										class="size-6"
									/>
								{:else}
									<Server class="size-6" />
								{/if}
							</div>
							<p class="text-sm">{componentServer.manifest?.name}</p>
							<span
								class="text-on-surface1 flex items-center gap-1 text-xs"
								title="This component server no longer exists"
							>
								<AlertCircle class="size-4" />
								<span>Deleted</span>
							</span>
						</div>
					</div>
				{/if}
			{/each}
		</div>
	</div>
{/if}

<div>
	<h2 class="mb-2 text-lg font-semibold">Connected Users</h2>

	<!-- show connected URL, configuration settings -->
	<Table data={connectedUsers} fields={['name']}>
		{#snippet onRenderColumn(property: string, d: OrgUser)}
			{#if property === 'name'}
				{d.email || d.username || 'Unknown'}
			{:else}
				{d[property as keyof typeof d]}
			{/if}
		{/snippet}

		{#snippet actions(d)}
			{@const auditLogsUrl = getAuditLogUrl(d)}
			{#if auditLogsUrl}
				<a href={resolve(auditLogsUrl as `/${string}`)} class="button-text"> View Audit Logs </a>
			{/if}
		{/snippet}
	</Table>
</div>

{#if mcpServerId}
	<!-- Single-user server delete confirmation -->
	{@const isDeletingOwnServer = isServerCreatedByCurrentUser}
	{@const serverOwner = createdByUser?.email || createdByUser?.username || 'Unknown User'}
	<Confirm
		show={!!showDeleteInstanceConfirm}
		onsuccess={async () => {
			if (!showDeleteInstanceConfirm) return;
			await handleDeleteInstance(mcpServerId);
		}}
		oncancel={() => {
			showDeleteInstanceConfirm = false;
			deleting = false;
		}}
		loading={deleting}
		msg={isDeletingOwnServer ? 'Delete Your MCP Server?' : `Delete User MCP Server?`}
		type="delete"
		title={isDeletingOwnServer ? 'Delete My Server' : 'Delete User Server'}
	>
		{#snippet note()}
			{#if isDeletingOwnServer}
				<p class="text-sm">
					You are about to delete <span class="font-semibold">your own MCP server</span>. All
					configurations and connections will be permanently removed.
				</p>
				<p class="mt-4 text-sm font-semibold text-red-500">This action cannot be undone.</p>
			{:else}
				<p class="text-sm">
					You are about to delete an MCP server owned by <span class="font-semibold"
						>{serverOwner}</span
					>. All configurations and connections will be permanently removed.
				</p>
				<p class="mt-4 text-sm font-semibold text-red-500">This action cannot be undone.</p>
			{/if}
		{/snippet}
	</Confirm>
{/if}
