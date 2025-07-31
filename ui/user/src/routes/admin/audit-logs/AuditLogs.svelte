<script lang="ts">
	import { twMerge } from 'tailwind-merge';

	let {
		data = [],
		onSelectRow,
		emptyContent,
		fetchUserById,
		currentFragmentIndex = 0,
		getFragmentIndex,
		getFragmentRowIndex,
		onLoadNextFragment
	} = $props();
</script>

<!-- Data Table -->
<div
	class="dark:bg-surface2 w-full overflow-hidden overflow-x-auto rounded-lg border border-transparent bg-white shadow-sm"
>
	{#if data.length}
		<table class="min-w-full divide-y divide-gray-200">
			<thead class="dark:bg-surface1 bg-surface2">
				<tr class="sticky top-0">
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500"
						>Timestamp</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500"
						>User</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500"
						>Server</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500"
						>Type</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500"
						>Identifier</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500"
						>Response Code</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500"
						>Response Time (ms)</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500"
						>Client</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500"
						>Client Version</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500"
						>IP Address</th
					>
				</tr>
			</thead>

			<tbody class="">
				<!-- Audit Data Rows -->
				{#each data as item, i (item.id)}
					{@const fragmentIndex = getFragmentIndex?.(i)}
					{@const fragmentRowIndex = getFragmentRowIndex?.(i)}
					<tr
						class={twMerge(
							'border-surface2 dark:border-surface2 shadow-xs border-t transition-colors duration-300',
							onSelectRow && ' hover:bg-surface1 dark:hover:bg-surface3 cursor-pointer',
							fragmentIndex && fragmentRowIndex === 0 && 'bg-surface3/50'
						)}
						data-fragment-index={fragmentIndex}
						data-fragment-row-index={fragmentRowIndex}
						onclick={() => onSelectRow?.(item)}
						{@attach (node) => {
							if (fragmentIndex < currentFragmentIndex) return;
							if (fragmentRowIndex > 0) return;

							const callback: IntersectionObserverCallback = (entries) => {
								const isIntersection = entries.some(
									(entry) => entry.target === node && entry.isIntersecting
								);

								if (isIntersection) {
									onLoadNextFragment?.(fragmentIndex);
								}
							};

							const rootElement = document.body;

							const observer = new IntersectionObserver(callback, {
								root: rootElement
							});

							observer.observe(node);

							return () => {
								observer.disconnect();
							};
						}}
					>
						<td class="whitespace-nowrap px-6 py-4 text-sm"
							>{new Date(item.createdAt)
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
						<td class="whitespace-nowrap px-6 py-4 text-sm">
							{#await fetchUserById(item.userID)}
								<span class="text-gray-500">Loading...</span>
							{:then user}
								{user?.displayName || 'Unknown User'}
							{/await}
						</td>
						<td class="whitespace-nowrap px-6 py-4 text-sm">{item.mcpServerDisplayName}</td>
						<td class="whitespace-nowrap px-6 py-4 text-sm">{item.callType}</td>
						<td class="whitespace-nowrap px-6 py-4 text-sm"></td>
						<td class="whitespace-nowrap px-6 py-4 text-sm">{item.responseStatus}</td>
						<td class="whitespace-nowrap px-6 py-4 text-sm">{item.processingTimeMs}</td>
						<td class="whitespace-nowrap px-6 py-4 text-sm">{item.client?.name}</td>
						<td class="whitespace-nowrap px-6 py-4 text-sm">{item.client?.version}</td>
						<td class="whitespace-nowrap px-6 py-4 text-sm">{item.clientIP}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{:else}
		{@render emptyContent?.()}
	{/if}
</div>
