<script module>
	export type FilterKey = Exclude<
		keyof AuditLogFilters,
		'query' | 'offset' | 'limit' | 'start_time' | 'end_time'
	>;

	export type FilterInput = {
		label: string;
		property: string;
		selected: string | number;
		options: { id: string; label: string }[];
	};

	export type FilterOption = {
		label: string;
		id: string;
	};
</script>

<script lang="ts">
	import AuditFilter from './AuditFilter.svelte';
	import { X } from 'lucide-svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import type { AuditLogFilters, OrgUser } from '$lib/services/admin/types';
	import { AdminService } from '$lib/services';
	import { untrack } from 'svelte';

	interface Props {
		filters: AuditLogFilters;
		onClose: () => void;
		fetchUserById: (userId: string) => Promise<OrgUser | undefined>;
		getFilterDisplayLabel?: (key: keyof AuditLogFilters) => string;
	}

	let { onClose, filters, fetchUserById, getFilterDisplayLabel }: Props = $props();

	type FilterOptions = Record<FilterKey, FilterOption[]>;
	let filtersOptions: FilterOptions = $state({} as FilterOptions);

	type FilterInputs = Record<FilterKey, FilterInput>;
	let filterInputs = (
		[
			'user_id',
			'mcp_id',
			'mcp_server_display_name',
			'mcp_server_catalog_entry_name',
			'call_type',
			'client_name',
			'client_version',
			'client_ip',
			'response_status',
			'session_id',
		] as FilterKey[]
	).reduce((acc, filterId) => {
		acc[filterId] = {
			property: filterId,
			label: getFilterDisplayLabel?.(filterId) ?? filterId.replace(/_(\w)/, ' $1'),
			get selected() {
				return filters?.[filterId] ?? '';
			},
			set selected(v) {
				filters[filterId] = v;
			},
			get options() {
				return filtersOptions[filterId];
			}
		};
		return acc;
	}, {} as FilterInputs);

	// {
	// 	user_id: {
	// 		label: 'User',
	// 		property: 'user_id',
	// 		get selected() {
	// 			return filters?.['user_id'] ?? '';
	// 		},
	// 		set selected(v) {
	// 			filters['user_id'] = v;
	// 		},
	// 		get options() {
	// 			return filtersOptions['user_id'];
	// 		}
	// 	},
	// 	mcp_server_display_name: {
	// 		label: 'MCP Server',
	// 		property: 'mcp_server_display_name',
	// 		get selected() {
	// 			return filters?.['mcp_server_display_name'] ?? '';
	// 		},
	// 		set selected(v) {
	// 			filters['mcp_server_display_name'] = v;
	// 		},
	// 		get options() {
	// 			return filtersOptions['mcp_server_display_name'];
	// 		}
	// 	},
	// 	mcp_server_catalog_entry_name: {
	// 		label: 'MCP Server Catalog Entry Name',
	// 		property: 'mcp_server_catalog_entry_name',
	// 		get selected() {
	// 			return filters?.['mcp_server_catalog_entry_name'] ?? '';
	// 		},
	// 		set selected(v) {
	// 			filters['mcp_server_catalog_entry_name'] = v;
	// 		},
	// 		get options() {
	// 			return filtersOptions['mcp_server_catalog_entry_name'];
	// 		}
	// 	},
	// 	call_type: {
	// 		label: 'Call Type',
	// 		property: 'call_type',
	// 		get selected() {
	// 			return filters?.['call_type'] ?? '';
	// 		},
	// 		set selected(v) {
	// 			filters['call_type'] = v;
	// 		},
	// 		get options() {
	// 			return filtersOptions['call_type'];
	// 		}
	// 	},
	// 	client_name: {
	// 		label: 'Client Name',
	// 		property: 'client_name',
	// 		get selected() {
	// 			return filters?.['client_name'] ?? '';
	// 		},
	// 		set selected(v) {
	// 			filters['client_name'] = v;
	// 		},
	// 		get options() {
	// 			return filtersOptions['client_name'];
	// 		}
	// 	},
	// 	client_version: {
	// 		label: 'Client Version',
	// 		property: 'client_version',
	// 		get selected() {
	// 			return filters?.['client_version'] ?? '';
	// 		},
	// 		set selected(v) {
	// 			filters['client_version'] = v;
	// 		},
	// 		get options() {
	// 			return filtersOptions['client_version'];
	// 		}
	// 	},
	// 	response_status: {
	// 		label: 'Response Status',
	// 		property: 'response_status',
	// 		get selected() {
	// 			return filters?.['response_status'] ?? '';
	// 		},
	// 		set selected(v) {
	// 			filters['response_status'] = v;
	// 		},
	// 		get options() {
	// 			return filtersOptions['response_status'];
	// 		}
	// 	},
	// 	session_id: {
	// 		label: 'Session ID',
	// 		property: 'session_id',
	// 		get selected() {
	// 			return filters?.['session_id'] ?? '';
	// 		},
	// 		set selected(v) {
	// 			filters['session_id'] = v;
	// 		},
	// 		get options() {
	// 			return filtersOptions['session_id'];
	// 		}
	// 	},
	// 	client_ip: {
	// 		label: 'Client IP',
	// 		property: 'client_ip',
	// 		get selected() {
	// 			return filters?.['client_ip'] ?? '';
	// 		},
	// 		set selected(v) {
	// 			filters['client_ip'] = v;
	// 		},
	// 		get options() {
	// 			return filtersOptions['client_ip'];
	// 		}
	// 	}
	// } as FilterInputs;

	const filterInputsAsArray = $derived(Object.values(filterInputs));

	$effect(() => {
		const processLog = async (filterId: string) => {
			const response = await AdminService.listAuditLogFilterOptions(filterId);

			if (filterId === 'user_id') {
				return await Promise.all(
					response.options
						.map((d) => fetchUserById(d).then((user) => ({ id: d, label: user?.displayName ?? d })))
						.filter(Boolean)
				);
			}

			return response.options.map((d) => ({
				id: d,
				label: d
			}));
		};

		Object.keys(filterInputs).forEach((id) => {
			processLog(id).then((options) => {
				untrack(() => {
					filtersOptions[id] = options;
				});
			});
		});
	});

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
			<AuditFilter
				filter={filterInput}
				onSelect={(_, value) => {
					filterInput.selected = value ?? '';
				}}
				onClear={(_, value) => {
					const key = filterInputsAsArray[index].property;
					filterInputs[key].selected = value;
				}}
			></AuditFilter>
		{/each}
		<div class="mt-auto">
			<button
				class="button-primary text-md w-full rounded-lg px-4 py-2"
				onclick={handleApplyFilters}>Apply Filters</button
			>
		</div>
	</div>
</div>
