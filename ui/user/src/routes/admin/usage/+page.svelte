<script lang="ts">
	import { fade } from 'svelte/transition';
	import Layout from '$lib/components/Layout.svelte';
	import { PAGE_TRANSITION_DURATION } from '$lib/constants';
	import UsageGraphs from '$lib/components/admin/usage/UsageGraphs.svelte';

	const duration = PAGE_TRANSITION_DURATION;

	// function compileSortAndFilters(): UsageStatsFilters & {
	// 	mcpId?: string | null;
	// 	sortBy?: string | null;
	// 	sortOrder?: string | null;
	// } {
	// 	if (!browser) return {};

	// 	const url = new URL(window.location.href);
	// 	const startTime = url.searchParams.get('startTime')
	// 		? decodeURIComponent(url.searchParams.get('startTime')!)
	// 		: null;
	// 	const endTime = url.searchParams.get('endTime')
	// 		? decodeURIComponent(url.searchParams.get('endTime')!)
	// 		: null;
	// 	// Handle both single and array-based parameters for backward compatibility
	// 	const userIds = url.searchParams.get('userIds')
	// 		? decodeURIComponent(url.searchParams.get('userIds')!)
	// 				.split(',')
	// 				.map((s) => s.trim())
	// 				.filter(Boolean)
	// 		: undefined;

	// 	const mcpServerDisplayNames = url.searchParams.get('mcpServerDisplayNames')
	// 		? decodeURIComponent(url.searchParams.get('mcpServerDisplayNames')!)
	// 				.split(',')
	// 				.map((s) => s.trim())
	// 				.filter(Boolean)
	// 		: undefined;

	// 	return {
	// 		startTime,
	// 		endTime,
	// 		userIds,
	// 		mcpServerDisplayNames
	// 	};
	// }
</script>

<Layout>
	<div class="my-4 h-dvh" in:fade={{ duration }} out:fade={{ duration }}>
		<div class="flex min-h-full flex-col gap-8 pb-8">
			<!-- {@render usageContent()} -->

			<div class="flex flex-col gap-8" in:fade={{ duration }}>
				<UsageGraphs />
			</div>
		</div>
	</div>
</Layout>

<svelte:head>
	<title>Obot | Usage</title>
</svelte:head>
