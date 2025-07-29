<script lang="ts">
	import { untrack } from 'svelte';
	import { X } from 'lucide-svelte';
	import { twMerge } from 'tailwind-merge';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import Select from '$lib/components/Select.svelte';
	import type { AuditLog, AuditLogFilters, OrgUser } from '$lib/services/admin/types';

	type FilterKeys =
		| 'user_id'
		| 'mcp_id'
		| 'mcp_server_display_name'
		| 'mcp_server_catalog_entry_name'
		| 'call_type'
		| 'session_id'
		| 'client_name'
		| 'client_version'
		| 'response_status'
		| 'client_ip';

	interface Props {
		auditLogs: (AuditLog & { user: string })[];
		onClose: () => void;
		filters?: AuditLogFilters;
		fetchUserById: (userId: string) => Promise<OrgUser | undefined>;
	}

	type FilterSet = {
		label: string;
		property: string;
		values: Record<string, FilterValue>;
		selected: string;
	};

	type FilterValue = {
		label: string;
		id: string;
	};

	let { auditLogs, onClose, filters, fetchUserById }: Props = $props();

	let filterInputs = $state<Record<FilterKeys, FilterSet>>({});
	const filterInputsAsArray = $derived(Object.values(filterInputs));

	$effect(() => {
		if (filters || auditLogs) {
			generateFilters(auditLogs, filters).then((res) => {
				untrack(() => (filterInputs = { ...res }));
			});
		}
	});

	async function generateFilters(logs: typeof auditLogs, filters?: AuditLogFilters) {
		const filterSets: Record<FilterKeys, FilterSet> = {
			user_id: {
				label: 'User',
				property: 'user_id',
				values: {},
				selected: filters?.['user_id'] ?? ''
			},
			mcp_server_display_name: {
				label: 'MCP Server',
				property: 'mcp_server_display_name',
				values: {},
				selected: filters?.['mcp_server_display_name'] ?? ''
			},
			mcp_server_catalog_entry_name: {
				label: 'MCP Server Catalog Entry Name',
				property: 'mcp_server_catalog_entry_name',
				values: {},
				selected: filters?.['mcp_server_catalog_entry_name'] ?? ''
			},
			call_type: {
				label: 'Call Type',
				property: 'call_type',
				values: {},
				selected: filters?.['call_type'] ?? ''
			},
			client_name: {
				label: 'Client Name',
				property: 'client_name',
				values: {},
				selected: filters?.['client_name'] ?? ''
			},
			client_version: {
				label: 'Client Version',
				property: 'client_version',
				values: {},
				selected: filters?.['client_version'] ?? ''
			},
			// response_status: {
			// 	label: 'Response Status',
			// 	property: 'response_status',
			// 	values: {},
			// 	selected: filters?.client ?? ''
			// },
			session_id: {
				label: 'Session ID',
				property: 'session_id',
				values: {},
				selected: filters?.['session_id'] ?? ''
			},
			client_ip: {
				label: 'Client IP',
				property: 'client_ip',
				values: {},
				selected: filters?.['client_ip'] ?? ''
			}
		};

		const processLog = async (filters: typeof filterSets, log) => {
			const { userID, mcpServerDisplayName, client, callType, sessionID } = log;

			if (userID) {
				const user = await fetchUserById(userID);
				if (user) {
					filters['user_id'].values[userID] = {
						label: user?.displayName ?? 'Unknown',
						id: userID
					};
				}
			}

			if (mcpServerDisplayName) {
				filters['mcp_server_display_name'].values[mcpServerDisplayName] = {
					label: mcpServerDisplayName,
					id: mcpServerDisplayName
				};
			}

			if (client) {
				filters['client_name'].values[client.name] = {
					label: client.name,
					id: client.name
				};
			}

			if (callType) {
				filters['call_type'].values[callType] = {
					label: callType,
					id: callType
				};
			}

			if (sessionID) {
				filters['session_id'].values[sessionID] = {
					label: sessionID,
					id: sessionID
				};
			}
		};

		await Promise.all(logs.map((log) => processLog(filterSets, log)));

		return filterSets;
	}

	function handleApplyFilters() {
		const url = page.url;

		for (const filterInput of filterInputsAsArray) {
			if (filterInput.selected) {
				url.searchParams.set(
					filterInput.property,
					encodeURIComponent(filterInput.selected.toString())
				);
			} else {
				page.url.searchParams.delete(filterInput.property);
			}
		}

		goto(url, { noScroll: true });
	}
</script>

<div class="dark:border-surface3 md:w-sm h-full w-screen border-l border-transparent">
	<div class="relative w-full text-center">
		<h4 class="p-4 text-xl font-semibold">Filters</h4>
		<button class="icon-button absolute right-4 top-1/2 -translate-y-1/2" onclick={onClose}>
			<X class="size-5" />
		</button>
	</div>
	<div
		class="default-scrollbar-thin flex h-[calc(100%-60px)] flex-col gap-4 overflow-y-auto p-4 pt-0"
	>
		{#each filterInputsAsArray as filterInput, index (filterInput.property)}
			{@const options = Object.values(filterInput.values)}

			<div class={twMerge('mb-2 flex flex-col gap-1', !options.length && 'opacity-50')}>
				<label for={filterInput.property} class="text-md font-light">
					By {filterInput.label}
				</label>

				<Select
					class="dark:border-surface3 bg-surface1 border border-transparent shadow-inner dark:bg-black"
					classes={{
						root: 'w-full',
						clear: 'hover:bg-surface3 bg-transparent'
					}}
					{options}
					selected={filterInput.selected}
					multiple={true}
					onSelect={(_, value) => {
						// filterInputsAsArray[index].selected = option.id.toString();
						// const key = filterInputsAsArray[index].property;
						// const values = new Set(
						// 	filterInputs[key]?.selected
						// 		?.split(',')
						// 		.map((d) => d.trim())
						// 		.filter(Boolean) ?? []
						// );
						// values.add(option.id.toString());
						// filterInputs[key].selected = values.values().toArray().join(',');
						// console.log(filterInputs[key].selected);
						filterInput.selected = value ?? '';
					}}
					onClear={(_, value) => {
						const key = filterInputsAsArray[index].property;
						filterInputs[key].selected = value;
					}}
					position="top"
				/>
			</div>
		{/each}
		<div class="mt-auto">
			<button
				class="button-primary text-md w-full rounded-lg px-4 py-2"
				onclick={handleApplyFilters}>Apply Filters</button
			>
		</div>
	</div>
</div>
