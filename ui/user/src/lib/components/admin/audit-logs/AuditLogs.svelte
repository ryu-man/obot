<script lang="ts">
	import { twMerge } from 'tailwind-merge';
	import { tooltip } from '$lib/actions/tooltip.svelte';
	import { VirtualTable } from '$lib/components/ui/virtual-table';

	let { data = [], onSelectRow, emptyContent, getUserDisplayName } = $props();

	let element: HTMLElement | undefined = $state();
	let navElementHeight = $state(64);

	const itemHeight = 56;

	$effect(() => {
		if (!element) return;

		const navElement = element.closest('main')?.querySelector('nav');

		navElementHeight = navElement?.clientHeight || 64;
	});
</script>

<!-- Data Table -->
<div
	bind:this={element}
	class="dark:bg-surface2 relative flex h-screen w-full min-w-full divide-y divide-gray-200 overflow-hidden overflow-x-auto rounded-lg border border-transparent bg-white shadow-sm"
	style="height: calc(100vh - {navElementHeight}px - 54px);"
>
	{#if data.length}
		<VirtualTable {data} {itemHeight} overscan={5} class={twMerge('w-full flex-1')}>
			{#snippet header()}
				<thead>
					<tr>
						<th
							scope="col"
							class="dark:bg-surface1 bg-surface2 sticky top-0 z-10 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
							style="width: 0px;">#</th
						>
						<th
							scope="col"
							class="dark:bg-surface1 bg-surface2 sticky top-0 z-10 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
							>Timestamp</th
						>
						<th
							scope="col"
							class="dark:bg-surface1 bg-surface2 sticky top-0 z-10 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
							>User</th
						>
						<th
							scope="col"
							class="dark:bg-surface1 bg-surface2 sticky top-0 z-10 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
							>Server</th
						>
						<th
							scope="col"
							class="dark:bg-surface1 bg-surface2 sticky top-0 z-10 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
							style="width: 100px;">Type</th
						>
						<th
							scope="col"
							class="dark:bg-surface1 bg-surface2 sticky top-0 z-10 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
							>Identifier</th
						>
						<th
							scope="col"
							class="dark:bg-surface1 bg-surface2 sticky top-0 z-10 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
							>Response Code</th
						>
						<th
							scope="col"
							class="dark:bg-surface1 bg-surface2 sticky top-0 z-10 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
							>Response Time (ms)</th
						>
						<th
							scope="col"
							class="dark:bg-surface1 bg-surface2 sticky top-0 z-10 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
							>Client</th
						>
						<th
							scope="col"
							class="dark:bg-surface1 bg-surface2 sticky top-0 z-10 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
							>IP Address</th
						>
					</tr>
				</thead>
			{/snippet}

			{#snippet children({ items })}
				{#each items as item (item.data.id)}
					{@const d = item.data}

					<tr
						class={twMerge(
							'virtual-list-row border-surface2 dark:border-surface2 h-14 border-t shadow-xs transition-colors duration-300',
							onSelectRow && ' hover:bg-surface1 dark:hover:bg-surface3 cursor-pointer'
						)}
						onclick={() => onSelectRow?.(d)}
					>
						<td class="px-6 py-4 text-sm whitespace-nowrap" style="width: 60px;"
							>{item.index + 1}</td
						>
						<td class="px-6 py-4 text-sm whitespace-nowrap" style="width: 160px;"
							>{new Date(d.createdAt)
								.toLocaleString(undefined, {
									year: 'numeric',
									month: 'short',
									day: 'numeric',
									hour: '2-digit',
									minute: '2-digit',
									second: '2-digit',
									hour12: true,
									timeZoneName: 'short'
								})
								.replace(/,/g, '')}</td
						>
						<td class="px-6 py-4 text-sm whitespace-nowrap" style="width: 120px;">
							{getUserDisplayName(d.userID)}
						</td>
						<td class="px-6 py-4 text-sm whitespace-nowrap" style="width: 150px;"
							>{d.mcpServerDisplayName}</td
						>
						<td class="px-6 py-4 text-sm whitespace-nowrap" style="width: 100px;">{d.callType}</td>
						<td class="px-6 py-4 text-sm whitespace-nowrap" style="width: 150px;"
							>{d.callIdentifier}</td
						>
						<td class="px-6 py-4 text-sm whitespace-nowrap" style="width: 120px;"
							>{d.responseStatus}</td
						>
						<td class="px-6 py-4 text-sm whitespace-nowrap" style="width: 140px;"
							>{d.processingTimeMs}</td
						>
						<td class="px-6 py-4 text-sm whitespace-nowrap" style="width: 120px;">
							<div class="max-w-[10ch] truncate" use:tooltip={d.client?.name}>
								{d.client?.name}
							</div>
						</td>
						<td class="px-6 py-4 text-sm whitespace-nowrap" style="width: 120px;">{d.clientIP}</td>
					</tr>
				{/each}
			{/snippet}
		</VirtualTable>
	{:else}
		{@render emptyContent?.()}
	{/if}
</div>
