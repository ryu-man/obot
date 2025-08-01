<script lang="ts">
	import { clickOutside } from '$lib/actions/clickoutside';
	import { tooltip } from '$lib/actions/tooltip.svelte';
	import McpServerEntryForm from '$lib/components/admin/McpServerEntryForm.svelte';
	import Confirm from '$lib/components/Confirm.svelte';
	import DotDotDot from '$lib/components/DotDotDot.svelte';
	import Layout from '$lib/components/Layout.svelte';
	import ResponsiveDialog from '$lib/components/ResponsiveDialog.svelte';
	import Table from '$lib/components/Table.svelte';
	import { DEFAULT_MCP_CATALOG_ID, PAGE_TRANSITION_DURATION } from '$lib/constants';
	import {
		fetchMcpServerAndEntries,
		getAdminMcpServerAndEntries,
		initMcpServerAndEntries
	} from '$lib/context/admin/mcpServerAndEntries.svelte';
	import { AdminService, type MCPCatalogServer } from '$lib/services';
	import type { MCPCatalog, MCPCatalogEntry } from '$lib/services/admin/types';
	import {
		AlertTriangle,
		Container,
		Eye,
		LoaderCircle,
		Plus,
		RefreshCcw,
		Server,
		Trash2,
		TriangleAlert,
		User,
		Users,
		X
	} from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { fade, fly, slide } from 'svelte/transition';
	import { goto } from '$app/navigation';
	import { afterNavigate } from '$app/navigation';
	import { browser } from '$app/environment';
	import BackLink from '$lib/components/admin/BackLink.svelte';
	import Search from '$lib/components/Search.svelte';

	const defaultCatalogId = DEFAULT_MCP_CATALOG_ID;
	let search = $state('');

	initMcpServerAndEntries();
	const mcpServerAndEntries = getAdminMcpServerAndEntries();

	onMount(async () => {
		await fetchMcpServerAndEntries(defaultCatalogId, mcpServerAndEntries, (entries, servers) => {
			const serverId = new URL(window.location.href).searchParams.get('id');
			if (serverId) {
				const foundEntry = entries.find((e) => e.id === serverId);
				const foundServer = servers.find((s) => s.id === serverId);
				const found = foundEntry || foundServer;

				if (found && selectedEntryServer?.id !== found.id) {
					selectedEntryServer = found;
					showServerForm = true;
				} else if (!found && selectedEntryServer) {
					selectedEntryServer = undefined;
					showServerForm = false;
				}
			} else {
				selectedEntryServer = undefined;
				showServerForm = false;
			}
		});
		defaultCatalog = await AdminService.getMCPCatalog(defaultCatalogId);
	});

	afterNavigate(({ to }) => {
		if (browser && to?.url) {
			const serverId = to.url.searchParams.get('id');
			const createNewType = to.url.searchParams.get('new') as 'single' | 'multi' | 'remote';
			if (createNewType) {
				selectServerType(createNewType, false);
			} else if (!serverId && (selectedEntryServer || showServerForm)) {
				selectedEntryServer = undefined;
				showServerForm = false;
			}
		}
	});

	function convertEntriesToTableData(entries: MCPCatalogEntry[] | undefined) {
		if (!entries) {
			return [];
		}

		return entries
			.filter((entry) => !entry.deleted)
			.map((entry) => {
				return {
					id: entry.id,
					name: entry.manifest?.name ?? '',
					icon: entry.manifest?.icon,
					source: entry.sourceURL || 'manual',
					data: entry,
					users: entry.userCount ?? 0,
					editable: !entry.sourceURL,
					type: entry.manifest.runtime === 'remote' ? 'remote' : 'single'
				};
			});
	}

	function convertServersToTableData(servers: MCPCatalogServer[] | undefined) {
		if (!servers) {
			return [];
		}

		return servers
			.filter((server) => !server.catalogEntryID && !server.deleted)
			.map((server) => {
				return {
					id: server.id,
					name: server.manifest.name ?? '',
					icon: server.manifest.icon,
					source: 'manual',
					type: 'multi',
					data: server,
					users: server.mcpServerInstanceUserCount ?? 0,
					editable: true
				};
			});
	}

	function convertEntriesAndServersToTableData(
		entries: MCPCatalogEntry[],
		servers: MCPCatalogServer[]
	) {
		const entriesTableData = convertEntriesToTableData(entries);
		const serversTableData = convertServersToTableData(servers);
		return [...entriesTableData, ...serversTableData];
	}

	let totalCount = $derived(
		mcpServerAndEntries.entries.length + mcpServerAndEntries.servers.length
	);
	let tableData = $derived(
		convertEntriesAndServersToTableData(mcpServerAndEntries.entries, mcpServerAndEntries.servers)
	);

	let filteredTableData = $derived(
		tableData.filter((d) => d.name.toLowerCase().includes(search.toLowerCase()))
	);

	let defaultCatalog = $state<MCPCatalog>();
	let editingSource = $state<{ index: number; value: string }>();
	let sourceDialog = $state<HTMLDialogElement>();
	let selectServerTypeDialog = $state<ReturnType<typeof ResponsiveDialog>>();
	let selectedServerType = $state<'single' | 'multi' | 'remote'>();
	let selectedEntryServer = $state<MCPCatalogEntry | MCPCatalogServer>();

	let syncError = $state<{ url: string; error: string }>();
	let syncErrorDialog = $state<ReturnType<typeof ResponsiveDialog>>();

	let showServerForm = $state(false);
	let deletingEntry = $state<MCPCatalogEntry>();
	let deletingServer = $state<MCPCatalogServer>();
	let deletingSource = $state<string>();
	let saving = $state(false);
	let refreshing = $state(false);
	let sourceError = $state<string>();
	function selectServerType(type: 'single' | 'multi' | 'remote', updateUrl = true) {
		selectedServerType = type;
		selectServerTypeDialog?.close();
		showServerForm = true;
		if (updateUrl) {
			goto(`/admin/mcp-servers?new=${type}`, { replaceState: false });
		}
	}

	function closeSourceDialog() {
		editingSource = undefined;
		sourceError = undefined;
		sourceDialog?.close();
	}

	async function refresh() {
		refreshing = true;
		await AdminService.refreshMCPCatalog(defaultCatalogId);
		await fetchMcpServerAndEntries(defaultCatalogId, mcpServerAndEntries);
		refreshing = false;
	}
	const duration = PAGE_TRANSITION_DURATION;
</script>

<Layout>
	<div class="flex flex-col gap-8 pt-4 pb-8" in:fade>
		{#if showServerForm}
			{@render configureEntryScreen()}
		{:else}
			{@render mainContent()}
		{/if}
	</div>
</Layout>

{#snippet mainContent()}
	<div
		class="flex flex-col gap-8"
		in:fly={{ x: 100, delay: duration, duration }}
		out:fly={{ x: -100, duration }}
	>
		<div class="flex items-center justify-between">
			<h1 class="flex items-center gap-2 text-2xl font-semibold">
				MCP Servers
				<button class="button-small flex items-center gap-1 text-xs font-normal" onclick={refresh}>
					{#if refreshing}
						<LoaderCircle class="size-4 animate-spin" /> Refreshing...
					{:else}
						<RefreshCcw class="size-4" />
						Refresh
					{/if}
				</button>
			</h1>
			{#if totalCount > 0}
				{@render addServerButton()}
			{/if}
		</div>

		<Search
			class="dark:bg-surface1 dark:border-surface3 border border-transparent bg-white shadow-sm"
			onChange={(val) => (search = val)}
			placeholder="Search servers..."
		/>

		{#if mcpServerAndEntries.loading}
			<div class="my-2 flex items-center justify-center">
				<LoaderCircle class="size-6 animate-spin" />
			</div>
		{:else if totalCount === 0}
			<div class="mt-12 flex w-md flex-col items-center gap-4 self-center text-center">
				<Server class="size-24 text-gray-200 dark:text-gray-900" />
				<h4 class="text-lg font-semibold text-gray-400 dark:text-gray-600">
					No created MCP servers
				</h4>
				<p class="text-sm font-light text-gray-400 dark:text-gray-600">
					Looks like you don't have any servers created yet. <br />
					Click the button below to get started.
				</p>

				{@render addServerButton()}
			</div>
		{:else}
			<Table
				data={filteredTableData}
				fields={['name', 'type', 'users', 'source']}
				onSelectRow={(d) => {
					if (d.type === 'single' || d.type === 'remote') {
						goto(`/admin/mcp-servers/c/${d.id}`);
					} else {
						goto(`/admin/mcp-servers/s/${d.id}`);
					}
				}}
				sortable={['name', 'type', 'users', 'source']}
				noDataMessage="No catalog servers added."
			>
				{#snippet onRenderColumn(property, d)}
					{#if property === 'name'}
						<div class="flex flex-shrink-0 items-center gap-2">
							<div
								class="bg-surface1 flex items-center justify-center rounded-sm p-0.5 dark:bg-gray-600"
							>
								{#if d.icon}
									<img src={d.icon} alt={d.name} class="size-6" />
								{:else}
									<Server class="size-6" />
								{/if}
							</div>
							<p class="flex items-center gap-1">
								{d.name}
								{#if d.source !== 'manual'}
									<span class="text-xs text-gray-500">({d.source.split('/').pop()})</span>{/if}
							</p>
						</div>
					{:else if property === 'type'}
						{d.type === 'single' ? 'Single User' : d.type === 'multi' ? 'Multi-User' : 'Remote'}
					{:else if property === 'source'}
						{d.source === 'manual' ? 'Web Console' : d.source}
					{:else}
						{d[property as keyof typeof d]}
					{/if}
				{/snippet}
				{#snippet actions(d)}
					{#if d.editable}
						<button
							class="icon-button hover:text-red-500"
							onclick={(e) => {
								e.stopPropagation();
								if (d.data.type === 'mcpserver') {
									deletingServer = d.data as MCPCatalogServer;
								} else {
									deletingEntry = d.data as MCPCatalogEntry;
								}
							}}
							use:tooltip={'Delete Entry'}
						>
							<Trash2 class="size-4" />
						</button>
					{/if}
					<button class="icon-button hover:text-blue-500" use:tooltip={'View Entry'}>
						<Eye class="size-4" />
					</button>
				{/snippet}
			</Table>
		{/if}

		{#if defaultCatalog?.sourceURLs && defaultCatalog.sourceURLs.length > 0 && defaultCatalog.id}
			<div class="flex flex-col gap-2" in:slide={{ axis: 'y', duration }}>
				<h2 class="mb-2 text-lg font-semibold">Git Source URLs</h2>

				<Table
					data={defaultCatalog?.sourceURLs?.map((url, index) => ({ id: index, url })) ?? []}
					fields={['url']}
					headers={[
						{
							property: 'url',
							title: 'URL'
						}
					]}
					noDataMessage="No Git Source URLs added."
					setRowClasses={(d) => {
						if (defaultCatalog?.syncErrors?.[d.url]) {
							return 'bg-yellow-500/10';
						}
						return '';
					}}
				>
					{#snippet actions(d)}
						<button
							class="icon-button hover:text-red-500"
							onclick={() => {
								deletingSource = d.url;
							}}
						>
							<Trash2 class="size-4" />
						</button>
					{/snippet}
					{#snippet onRenderColumn(property, d)}
						{#if property === 'url'}
							<div class="flex items-center gap-2">
								<p>{d.url}</p>
								{#if defaultCatalog?.syncErrors?.[d.url]}
									<button
										onclick={() => {
											syncError = {
												url: d.url,
												error: defaultCatalog?.syncErrors?.[d.url] ?? ''
											};
											syncErrorDialog?.open();
										}}
										use:tooltip={{
											text: 'An issue occurred. Click to see more details.',
											classes: ['break-words']
										}}
									>
										<TriangleAlert class="size-4 text-yellow-500" />
									</button>
								{/if}
							</div>
						{/if}
					{/snippet}
				</Table>
			</div>
		{/if}
	</div>
{/snippet}

{#snippet configureEntryScreen()}
	{@const currentLabelType =
		selectedServerType === 'single'
			? 'Single User'
			: selectedServerType === 'multi'
				? 'Multi-User'
				: 'Remote'}
	<div class="flex flex-col gap-6" in:fly={{ x: 100, delay: duration, duration }}>
		<BackLink fromURL="mcp-servers" currentLabel={`Create ${currentLabelType} Server`} />
		<McpServerEntryForm
			type={selectedServerType}
			catalogId={defaultCatalogId}
			onCancel={() => {
				selectedEntryServer = undefined;
				showServerForm = false;
			}}
			onSubmit={async (id, type) => {
				if (type === 'single' || type === 'remote') {
					goto(`/admin/mcp-servers/c/${id}`);
				} else {
					goto(`/admin/mcp-servers/s/${id}`);
				}
			}}
		/>
	</div>
{/snippet}

{#snippet addServerButton()}
	<DotDotDot class="button-primary text-sm" placement="bottom">
		{#snippet icon()}
			<span class="flex items-center gap-1">
				<Plus class="size-4" /> Add MCP Server
			</span>
		{/snippet}
		<div class="default-dialog flex min-w-max flex-col p-2">
			<button
				class="menu-button"
				onclick={() => {
					selectServerTypeDialog?.open();
				}}
			>
				Add server
			</button>
			<button
				class="menu-button"
				onclick={() => {
					editingSource = {
						index: -1,
						value: ''
					};
					sourceDialog?.showModal();
				}}
			>
				Add server(s) from Git
			</button>
		</div>
	</DotDotDot>
{/snippet}

<dialog
	bind:this={sourceDialog}
	use:clickOutside={() => closeSourceDialog()}
	class="w-full max-w-md p-4"
>
	{#if editingSource}
		<h3 class="default-dialog-title">
			{editingSource.index === -1 ? 'Add Source URL' : 'Edit Source URL'}
			<button onclick={() => closeSourceDialog()} class="icon-button">
				<X class="size-5" />
			</button>
		</h3>

		<div class="my-4 flex flex-col gap-1">
			<label for="catalog-source-name" class="flex-1 text-sm font-light capitalize"
				>Source URL
			</label>
			<input id="catalog-source-name" bind:value={editingSource.value} class="text-input-filled" />
		</div>

		{#if sourceError}
			<div class="mb-4 flex flex-col gap-2 text-red-500 dark:text-red-400">
				<div class="flex items-center gap-2">
					<AlertTriangle class="size-6 flex-shrink-0 self-start" />
					<p class="my-0.5 flex flex-col text-sm font-semibold">Error adding source URL:</p>
				</div>
				<span class="font-sm font-light break-all">{sourceError}</span>
			</div>
		{/if}

		<div class="flex w-full justify-end gap-2">
			<button class="button" disabled={saving} onclick={() => closeSourceDialog()}>Cancel</button>
			<button
				class="button-primary"
				disabled={saving}
				onclick={async () => {
					if (!editingSource) {
						return;
					}

					saving = true;
					sourceError = undefined;

					try {
						const catalog = await AdminService.getMCPCatalog(defaultCatalogId);

						if (editingSource.index === -1) {
							catalog.sourceURLs = [...(catalog.sourceURLs ?? []), editingSource.value];
						} else {
							catalog.sourceURLs[editingSource.index] = editingSource.value;
						}

						const response = await AdminService.updateMCPCatalog(defaultCatalogId, catalog, {
							dontLogErrors: true
						});
						defaultCatalog = response;
						await refresh();
						closeSourceDialog();
					} catch (error) {
						sourceError = error instanceof Error ? error.message : 'An unexpected error occurred';
					} finally {
						saving = false;
					}
				}}
			>
				Add
			</button>
		</div>
	{/if}
</dialog>

<Confirm
	msg="Are you sure you want to delete this server?"
	show={Boolean(deletingEntry)}
	onsuccess={async () => {
		if (!deletingEntry) {
			return;
		}

		await AdminService.deleteMCPCatalogEntry(defaultCatalogId, deletingEntry.id);
		await fetchMcpServerAndEntries(defaultCatalogId, mcpServerAndEntries);
		deletingEntry = undefined;
	}}
	oncancel={() => (deletingEntry = undefined)}
/>

<Confirm
	msg="Are you sure you want to delete this server?"
	show={Boolean(deletingServer)}
	onsuccess={async () => {
		if (!deletingServer) {
			return;
		}

		await AdminService.deleteMCPCatalogServer(defaultCatalogId, deletingServer.id);
		await fetchMcpServerAndEntries(defaultCatalogId, mcpServerAndEntries);
		deletingServer = undefined;
	}}
	oncancel={() => (deletingServer = undefined)}
/>

<Confirm
	msg="Are you sure you want to delete this Git Source URL?"
	show={Boolean(deletingSource)}
	onsuccess={async () => {
		if (!deletingSource || !defaultCatalog) {
			return;
		}

		const response = await AdminService.updateMCPCatalog(defaultCatalogId, {
			...defaultCatalog,
			sourceURLs: defaultCatalog.sourceURLs?.filter((url) => url !== deletingSource)
		});
		await refresh();
		defaultCatalog = response;
		deletingSource = undefined;
	}}
	oncancel={() => (deletingSource = undefined)}
/>

<ResponsiveDialog title="Select Server Type" class="md:w-lg" bind:this={selectServerTypeDialog}>
	<div class="my-4 flex flex-col gap-4">
		<button
			class="group dark:bg-surface2 hover:bg-surface1 dark:hover:bg-surface3 dark:border-surface3 border-surface2 flex cursor-pointer items-center gap-4 rounded-md border bg-white px-2 py-4 text-left transition-colors duration-300"
			onclick={() => selectServerType('single')}
		>
			<User
				class="size-12 flex-shrink-0 pl-1 text-gray-500 transition-colors group-hover:text-inherit"
			/>
			<div>
				<p class="mb-1 text-sm font-semibold">Single User Server</p>
				<span class="block text-xs leading-4 text-gray-400 dark:text-gray-600">
					This option is appropriate for servers that require individualized configuration or were
					not designed for multi-user access, such as most studio MCP servers. When a user selects
					this server, a private instance will be created for them.
				</span>
			</div>
		</button>
		<button
			class="group dark:bg-surface2 hover:bg-surface1 dark:hover:bg-surface3 dark:border-surface3 border-surface2 flex cursor-pointer items-center gap-4 rounded-md border bg-white px-2 py-4 text-left transition-colors duration-300"
			onclick={() => selectServerType('multi')}
		>
			<Users
				class="size-12 flex-shrink-0 pl-1 text-gray-500 transition-colors group-hover:text-inherit"
			/>
			<div>
				<p class="mb-1 text-sm font-semibold">Multi-User Server</p>
				<span class="block text-xs leading-4 text-gray-400 dark:text-gray-600">
					This option is appropriate for servers designed to handle multiple user connections, such
					as most Streamable HTTP servers. When you create this server, a running instance will be
					deployed and any user with access to this catlog will be able to connect to it.
				</span>
			</div>
		</button>
		<button
			class="group dark:bg-surface2 hover:bg-surface1 dark:hover:bg-surface3 dark:border-surface3 border-surface2 flex cursor-pointer items-center gap-4 rounded-md border bg-white px-2 py-4 text-left transition-colors duration-300"
			onclick={() => selectServerType('remote')}
		>
			<Container
				class="size-12 flex-shrink-0 pl-1 text-gray-500 transition-colors group-hover:text-inherit"
			/>
			<div>
				<p class="mb-1 text-sm font-semibold">Remote Server</p>
				<span class="block text-xs leading-4 text-gray-400 dark:text-gray-600">
					This option is appropriate for allowing users to connect to MCP servers that are already
					elsewhere. When a user selects this server, their connection to the remote MCP server will
					go through the Obot gateway.
				</span>
			</div>
		</button>
	</div>
</ResponsiveDialog>

<ResponsiveDialog title="Git Source URL Sync" bind:this={syncErrorDialog} class="md:w-2xl">
	<div class="mb-4 flex flex-col gap-4">
		<div class="notification-alert flex flex-col gap-2">
			<div class="flex items-center gap-2">
				<AlertTriangle class="size-6 flex-shrink-0 self-start text-yellow-500" />
				<p class="my-0.5 flex flex-col text-sm font-semibold">
					An issue occurred fetching this source URL:
				</p>
			</div>
			<span class="font-sm font-light break-all">{syncError?.error}</span>
		</div>
	</div>
</ResponsiveDialog>

<svelte:head>
	<title>Obot | MCP Servers</title>
</svelte:head>
