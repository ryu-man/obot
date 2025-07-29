<script lang="ts">
	import { clickOutside } from '$lib/actions/clickoutside';
	import { tooltip } from '$lib/actions/tooltip.svelte';
	import Calendar from '$lib/components/Calendar.svelte';
	import { formatTimeRange, getTimeRangeShorthand } from '$lib/time';
	import { twMerge } from 'tailwind-merge';

	let { start, end, onChange } = $props();

	let quickAccessPopover = $state<HTMLDialogElement>();

	const actions = [
		{
			label: 'Last Hour',
			onpointerdown: () => {
				end = new Date();
				const _start = new Date(end);
				_start.setHours(end.getHours() - 1);
				start = _start;

				onChange({ end: end, start: _start });
				quickAccessPopover?.close();
			}
		},
		{
			label: 'Last 6 Hour',
			onpointerdown: () => {
				end = new Date();
				const _start = new Date(end);
				_start.setHours(end.getHours() - 6);

				start = _start;

				onChange({ end, start: _start });
				quickAccessPopover?.close();
			}
		},
		{
			label: 'Last 24 Hour',
			onpointerdown: () => {
				end = new Date();
				const _start = new Date(end);
				_start.setHours(end.getHours() - 24);

				start = _start;

				onChange({ end, start: _start });
				quickAccessPopover?.close();
			}
		},
		{
			label: 'Last 7 Days',
			onpointerdown: () => {
				end = new Date();
				const _start = new Date(end);
				_start.setDate(end.getDate() - 7);

				start = _start;

				onChange({ end, start: _start });
				quickAccessPopover?.close();
			}
		}
	];
</script>

<div class="flex">
	<div class="relative flex items-center">
		<button
			class="dark:border-surface3 dark:hover:bg-surface2/70 dark:active:bg-surface2 dark:bg-surface1 hover:bg-surface1/70 active:bg-surface1 flex min-h-12.5 flex-shrink-0 items-center gap-2 truncate rounded-l-lg border border-r-0 border-transparent bg-white px-2 text-sm shadow-sm transition-colors duration-200"
			onpointerdown={() => {
				if (quickAccessPopover?.open) {
					quickAccessPopover?.close();
				} else {
					quickAccessPopover?.show();
				}
			}}
			use:tooltip={{
				text: 'Calendar Quick Actions',
				placement: 'top-end'
			}}
		>
			<span class="bg-surface3 rounded-md px-3 py-1 text-xs">
				{getTimeRangeShorthand(start, end)}
			</span>
			<span>
				{formatTimeRange(start, end)}
			</span>
		</button>

		<dialog
			use:clickOutside={[() => quickAccessPopover?.close(), true]}
			class={twMerge(
				'p-y absolute top-full right-0 left-[unset] z-50 m-0 mt-1 min-w-fit overflow-hidden'
			)}
			{@attach (node) => node.close()}
			{@attach (node) => (quickAccessPopover = node)}
		>
			<div class="flex flex-col items-start">
				{#each actions as action}
					<button
						class="hover:bg-surface3 w-full min-w-max px-4 py-2 text-start"
						onpointerdown={action.onpointerdown}
					>
						{action.label}
					</button>
				{/each}
			</div>
		</dialog>
	</div>

	<Calendar
		compact
		class="dark:border-surface3 hover:bg-surface1 dark:hover:bg-surface3 dark:bg-surface1 flex min-h-12.5 flex-shrink-0 items-center gap-2 truncate rounded-none rounded-r-lg border border-transparent bg-white px-4 text-sm shadow-sm"
		initialValue={{
			start: new Date(start),
			end: end ? new Date(end) : null
		}}
		{start}
		{end}
		{onChange}
	/>
</div>
