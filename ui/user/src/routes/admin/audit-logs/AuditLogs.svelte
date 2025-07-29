<script lang="ts">
	import { twMerge } from 'tailwind-merge';

	let { data = [], onSelectRow, emptyContent, fetchUserById } = $props();

	// 	{
	//     "id": 59,
	//     "createdAt": "2025-07-24T13:02:29.776039125-04:00",
	//     "userID": "2",
	//     "mcpID": "ms14rtsg",
	//     "mcpServerDisplayName": "Simple Server",
	//     "mcpServerCatalogEntryName": "default-simple-server-cc9435a7c8srb",
	//     "client": {
	//         "name": "nanobot",
	//         "version": "v0.0.0-dev"
	//     },
	//     "clientIP": "127.0.0.1",
	//     "callType": "notifications/initialized",
	//     "responseStatus": 200,
	//     "processingTimeMs": 0,
	//     "sessionID": "99d0cfb0-67f8-4c20-bb97-77694d62767a",
	//     "requestID": "3c5ba0b3-477c-42d2-83f0-d1b0e64a1ea0",
	//     "userAgent": "Go-http-client/1.1",
	//     "requestHeaders": {
	//         "Accept": [
	//             "application/json, text/event-stream"
	//         ],
	//         "Accept-Encoding": [
	//             "gzip"
	//         ],
	//         "Content-Length": [
	//             "54"
	//         ],
	//         "Content-Type": [
	//             "application/json"
	//         ],
	//         "Mcp-Session-Id": [
	//             "99d0cfb0-67f8-4c20-bb97-77694d62767a"
	//         ],
	//         "User-Agent": [
	//             "Go-http-client/1.1"
	//         ]
	//     },
	//     "responseHeaders": {
	//         "Vary": [
	//             "Origin"
	//         ],
	//         "X-Ratelimit-Limit": [
	//             "1000"
	//         ],
	//         "X-Ratelimit-Remaining": [
	//             "997"
	//         ],
	//         "X-Ratelimit-Reset": [
	//             "Thu, 24 Jul 2025 17:02:29 UTC"
	//         ]
	//     }
	// }
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
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>Timestamp</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>User</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>Server</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>Type</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>Identifier</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>Response Code</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>Response Time (ms)</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>Client</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>Client Version</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>IP Address</th
					>
				</tr>
			</thead>

			<tbody class="">
				<!-- Sample Data Rows -->
				{#each data as item (item.id)}
					<tr
						class={twMerge(
							'border-surface2 dark:border-surface2 border-t shadow-xs transition-colors duration-300',
							onSelectRow && ' hover:bg-surface1 dark:hover:bg-surface3 cursor-pointer'
						)}
						onclick={() => onSelectRow?.(item)}
					>
						<td class="px-6 py-4 text-sm whitespace-nowrap"
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
						<td class="px-6 py-4 text-sm whitespace-nowrap">
							{#await fetchUserById(item.userID)}
								<span class="text-gray-500">Loading...</span>
							{:then user}
								{user?.displayName || 'Unknown User'}
							{/await}
						</td>
						<td class="px-6 py-4 text-sm whitespace-nowrap">{item.mcpServerDisplayName}</td>
						<td class="px-6 py-4 text-sm whitespace-nowrap">{item.callType}</td>
						<td class="px-6 py-4 text-sm whitespace-nowrap"></td>
						<td class="px-6 py-4 text-sm whitespace-nowrap">{item.responseStatus}</td>
						<td class="px-6 py-4 text-sm whitespace-nowrap">{item.processingTimeMs}</td>
						<td class="px-6 py-4 text-sm whitespace-nowrap">{item.client?.name}</td>
						<td class="px-6 py-4 text-sm whitespace-nowrap">{item.client?.version}</td>
						<td class="px-6 py-4 text-sm whitespace-nowrap">{item.clientIP}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{:else}
		{@render emptyContent?.()}
	{/if}
</div>
