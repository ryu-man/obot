<script lang="ts">
	import { fade, slide } from 'svelte/transition';
	import { X, ChevronLeft, ChevronRight } from 'lucide-svelte';
	import { throttle } from 'es-toolkit';
	import { page } from '$app/state';
	import { afterNavigate, goto } from '$app/navigation';
	import { type DateRange } from '$lib/components/Calendar.svelte';
	import Layout from '$lib/components/Layout.svelte';
	import Search from '$lib/components/Search.svelte';
	import { PAGE_TRANSITION_DURATION } from '$lib/constants';
	import { type OrgUser, type AuditLogFilters, AdminService, type AuditLog } from '$lib/services';
	import { getUser, type PaginatedResponse } from '$lib/services/admin/operations';
	import { clickOutside } from '$lib/actions/clickoutside';
	import { dialogAnimation } from '$lib/actions/dialogAnimation';
	import AuditLogDetails from '$lib/components/admin/audit-logs/AuditLogDetails.svelte';
	import AuditFilters from '$lib/components/admin/audit-logs/AuditFilters.svelte';
	import { tooltip } from '$lib/actions/tooltip.svelte';
	import AuditLogsTable from './AuditLogs.svelte';
	import AuditLogsTimeline from './AuditLogsTimeline.svelte';
	import AuditLogCalendar from './AuditLogCalendar.svelte';
	import { endOfDay, set } from 'date-fns';

	const duration = PAGE_TRANSITION_DURATION;

	let auditLogsResponse = $state<PaginatedResponse<AuditLog>>();
	const auditLogsTotalItems = $derived(auditLogsResponse?.total ?? 0);

	let pageIndex = $state(0);
	const pageLimit = $state(40);

	const numberOfPages = $derived(Math.ceil(auditLogsTotalItems / pageLimit));
	const pageOffset = $derived(pageIndex * pageLimit);

	const remoteAuditLogs = $derived(
		(auditLogsResponse?.items ?? []).map(({ createdAt, ...restProps }) => ({
			...restProps,
			createdAt: new Date(createdAt.slice(0, -6))+'Z'
		}))
	);

	const isReachedMax = $derived(pageIndex >= numberOfPages - 1);
	const isReachedMin = $derived(pageIndex <= 0);

	let fragmentIndex = $state(0);
	const fragmentLimit = $state(10);
	const numberOfFragments = $derived(Math.ceil(remoteAuditLogs.length / fragmentLimit));
	const fragmentSliceStart = $derived(0);
	const fragmentSliceEnd = $derived(
		Math.min(remoteAuditLogs.length, (fragmentIndex + 1) * fragmentLimit)
	);

	const fragmentedAuditLogs = $derived(remoteAuditLogs.slice(fragmentSliceStart, fragmentSliceEnd));
	const haveMoreFragments = $derived(fragmentedAuditLogs.length < remoteAuditLogs.length);

	const users = new Map<string, OrgUser>();

	let showFilters = $state(false);
	let selectedAuditLog = $state<AuditLog & { user: string }>();
	let rightSidebar = $state<HTMLDialogElement>();

	let query = $state('');

	const searchParamFilters = $derived.by<AuditLogFilters & { mcpId?: string | null }>(() => {
		return page.url.searchParams.entries().reduce((acc, [key, value]) => {
			acc[key] = decodeURIComponent(value ?? '');
			return acc;
		}, {});
	});

	let sortFilters = $derived.by(() => {
		if (searchParamFilters.sortBy && searchParamFilters.sortOrder) {
			return {
				sortBy: searchParamFilters.sortBy,
				sortOrder: searchParamFilters.sortOrder as 'asc' | 'desc'
			};
		}

		return {};
	});

	let timeRangeFilters = $derived.by(() => {
		if (searchParamFilters.start_time || searchParamFilters.end_time) {
			return {
				start_time: searchParamFilters.start_time ?? '',
				end_time: searchParamFilters.end_time ?? ''
			};
		}

		return {
			start_time: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
			end_time: new Date().toISOString()
		};
	});

	const allFilters = $derived({
		...searchParamFilters,
		...sortFilters,
		limit: pageLimit,
		offset: pageOffset,
		query
	});

	afterNavigate(() => {
		AdminService.listUsers().then((userData) => {
			for (const user of userData) {
				users.set(user.id, user);
			}
		});
	});

	$effect(() => {
		if (!allFilters) return;

		fetchAuditLogs({ ...allFilters });
	});

	// Throttle query update
	const handleQueryChange = throttle((value: string) => {
		query = value;
	}, 100);

	async function nextPage() {
		if (isReachedMax) return;

		pageIndex = Math.min(numberOfPages, pageIndex + 1);

		fetchAuditLogs({ ...allFilters });
	}

	async function prevPage() {
		if (isReachedMin) return;

		pageIndex = Math.max(0, pageIndex - 1);

		fetchAuditLogs({ ...allFilters });
	}

	async function fetchAuditLogs(filters: typeof searchParamFilters) {
		const mcpId = filters.mcpId;

		if (mcpId) {
			auditLogsResponse = await AdminService.listServerOrInstanceAuditLogs(mcpId, filters);
		} else {
			auditLogsResponse = await AdminService.listAuditLogs(filters);
		}
	}

	async function fetchUserById(id: string) {
		const cache = users.get(id);

		if (cache) {
			return cache;
		}

		const remote = await getUser(id);
		users.set(id, remote);

		return remote;
	}

	function getFilterDisplayLabel(key: string) {
		if (key === 'mcpServerDisplayName') return 'Server';
		if (key === 'mcpServerCatalogEntryName') return 'Server ID';
		if (key === 'mcpId') return 'Server ID';
		if (key === 'startTime') return 'Start Time';
		if (key === 'endTime') return 'End Time';
		if (key === 'userId') return 'User ID';
		if (key === 'client') return 'Client';
		if (key === 'callType') return 'Call Type';
		if (key === 'sessionId') return 'Session ID';

		return key.replace(/_(\w)/g, ' $1');
	}

	function getFilterValue(label: string, value: string | number) {
		if (label === 'start_time' || label === 'end_time') {
			return new Date(value).toLocaleString(undefined, {
				year: 'numeric',
				month: 'short',
				day: 'numeric',
				hour: '2-digit',
				minute: '2-digit',
				second: '2-digit',
				hour12: true,
				timeZoneName: 'short'
			});
		}

		return value + '';
	}

	function handleRightSidebarClose() {
		rightSidebar?.close();
		setTimeout(() => {
			showFilters = false;
			selectedAuditLog = undefined;
		}, 300);
	}

	function handleDateChange({ start, end }: DateRange) {
		const url = page.url;

		if (start) {
			url.searchParams.set('start_time', start.toISOString());

			if (end) {
				url.searchParams.set('end_time', end.toISOString());
			} else {
				const end = endOfDay(start);
				url.searchParams.set('end_time', end.toISOString());
			}
		}

		goto(url.toString(), { noScroll: true });
	}
</script>

<svelte:head>
	<title>Obot | Audit Logs</title>
</svelte:head>

<Layout>
	<div class="my-4 h-screen" in:fade={{ duration }} out:fade={{ duration }}>
		<div class="flex min-h-full flex-col gap-8 pb-8">
			<div class="flex items-center justify-between gap-4">
				<h1 class="text-2xl font-semibold">Audit Logs</h1>
			</div>

			<div class="flex gap-4">
				<Search
					class="dark:bg-surface1 dark:border-surface3 border border-transparent bg-white shadow-sm"
					onChange={handleQueryChange}
					placeholder="Search..."
				/>

				<AuditLogCalendar
					start={timeRangeFilters.start_time
						? set(new Date(timeRangeFilters.start_time), { seconds: 0, milliseconds: 0 })
						: null}
					end={timeRangeFilters.end_time
						? set(new Date(timeRangeFilters.end_time), { milliseconds: 0, seconds: 0 })
						: null}
					onChange={handleDateChange}
				/>

				<button
					class="dark:bg-surface1 dark:hover:bg-surface2/70 dark:active:bg-surface2 dark:border-surface3 flex w-full items-center justify-center rounded-lg border border-transparent bg-white px-4 py-2 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 sm:w-auto"
					onclick={() => {
						showFilters = true;
						selectedAuditLog = undefined;
						rightSidebar?.show();
					}}
					use:tooltip={'Filter Logs'}
				>
					<svg
						class="mr-2 h-5 w-5"
						xmlns="http://www.w3.org/2000/svg"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
					>
						<polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3"></polygon>
					</svg>

					Filters
				</button>
			</div>

			{@render filters()}

			<!-- Timeline Graph (Placeholder) -->
			<div
				class="dark:bg-surface2 dark:border-surface3 rounded-lg border border-transparent bg-white text-black shadow-sm dark:text-white"
			>
				<h3 class="mb-2 px-4 pt-4 text-lg font-medium">Timeline</h3>
				<div class="px-4">
					<div class="flex h-40 items-center justify-center rounded-md text-gray-500">
						<AuditLogsTimeline
							data={remoteAuditLogs}
							start={timeRangeFilters.start_time
								? set(new Date(timeRangeFilters.start_time), { milliseconds: 0, seconds: 0 })
								: null}
							end={timeRangeFilters.end_time
								? set(new Date(timeRangeFilters.end_time), { milliseconds: 0, seconds: 0 })
								: null}
						/>
					</div>
				</div>
				<hr class="dark:border-surface3 my-4 border" />
				<div class="flex items-center justify-between gap-2 px-4 pb-4 text-xs text-gray-600">
					<div class="flex gap-4">
						<div>{Intl.NumberFormat().format(remoteAuditLogs.length)} results</div>

						<div class="flex items-center">
							<sapn>{Intl.NumberFormat().format(pageIndex + 1)}</sapn>/
							<span>{Intl.NumberFormat().format(numberOfPages)}</span>
							<span class="ml-1">pages</span>
						</div>
					</div>

					<div class="flex gap-4">
						<button class="flex items-center text-xs" disabled={isReachedMin} onclick={prevPage}>
							<ChevronLeft class="size-[1.4em]" />
							<div>Previous Page</div>
						</button>

						<button class="flex items-center text-xs" disabled={isReachedMax} onclick={nextPage}>
							<div>Next Page</div>
							<ChevronRight class="size-[1.4em]" />
						</button>
					</div>
				</div>
			</div>

			<AuditLogsTable
				data={fragmentedAuditLogs}
				currentFragmentIndex={fragmentIndex}
				getFragmentIndex={(rowIndex: number) => Math.floor(rowIndex / fragmentLimit)}
				getFragmentRowIndex={(rowIndex: number) => {
					const fragIndex = Math.floor(rowIndex / fragmentLimit);

					return rowIndex - fragIndex * fragmentLimit;
				}}
				{haveMoreFragments}
				onLoadNextFragment={(rowFragmentIndex: number) => {
					fragmentIndex = Math.min(numberOfFragments - 1, rowFragmentIndex + 1);
				}}
				onSelectRow={(d) => {
					selectedAuditLog = d;
					showFilters = false;
					rightSidebar?.show();
				}}
				{fetchUserById}
			/>
		</div>
	</div>
</Layout>

<dialog
	bind:this={rightSidebar}
	use:clickOutside={[handleRightSidebarClose, true]}
	use:dialogAnimation={{ type: 'drawer' }}
	class="dark:border-surface1 dark:bg-surface1 fixed! top-0! right-0! bottom-0! left-auto! outline-none! z-40 h-screen w-auto max-w-none rounded-none border-0 bg-white shadow-lg"
>
	{#if selectedAuditLog}
		<AuditLogDetails onClose={handleRightSidebarClose} auditLog={selectedAuditLog} />
	{/if}
	{#if showFilters}
		<AuditFilters
			onClose={handleRightSidebarClose}
			filters={{ ...searchParamFilters }}
			{fetchUserById}
		/>
	{/if}
</dialog>

{#snippet filters()}
	{@const keys = Object.keys(searchParamFilters)}
	{@const hasFilters = Object.values(searchParamFilters).some((value) => !!value)}

	{#if hasFilters}
		<div
			class="flex flex-wrap items-center gap-2"
			in:slide={{ duration: 200 }}
			out:slide={{ duration: 100 }}
		>
			{#each keys.filter((key) => key !== 'startTime' && key !== 'endTime' && key !== 'sortBy' && key !== 'sortOrder') as key (key)}
				{@const displayLabel = getFilterDisplayLabel(key)}
				{@const value = searchParamFilters[key as keyof typeof searchParamFilters]}
				{#if value}
					<div
						class="bg-blue-500/33 flex items-center gap-1 rounded-lg border border-blue-500 px-4 py-2"
					>
						<p class="text-xs font-semibold">
							{displayLabel}: <span class="font-light">{getFilterValue(key, value)}</span>
						</p>

						<button
							class="rounded-full p-1 transition-colors duration-200 hover:bg-blue-500/50"
							onclick={() => {
								const url = page.url;
								url.searchParams.delete(key);

								goto(url, { noScroll: true });
							}}
						>
							<X class="size-3" />
						</button>
					</div>
				{/if}
			{/each}
		</div>
	{/if}
{/snippet}
