<script lang="ts">
	import { getHours, getMinutes, setHours, setMinutes } from 'date-fns';
	import { twMerge } from 'tailwind-merge';

	type Props = {
		date: Date;
		onChange?: (date: Date) => void;
	};
	let { date, onChange }: Props = $props();

	const hours = $derived(getHours(date));
	let isAm = $derived(getHours(date) < 12);
</script>

<div class="time-input flex h-14 items-center gap-2  bg-surface1/50 rounded-md">
	<div class="flex min-h-full flex-1 text-xl">
		<input
			class="min-h-full w-full bg-transparent px-4 text-end"
			type="number"
			max="12"
			min="0"
			bind:value={
				() => getHours(date) % 12,
				(v) => {
					date = setHours(date, v);
					onChange?.(date);
				}
			}
		/>
	</div>

	<div class="text-4xl font-bold">:</div>

	<div class=" flex min-h-full flex-1 rounded-md text-xl">
		<input
			class="min-h-full w-full bg-transparent px-4"
			type="number"
			max="12"
			min="0"
			bind:value={
				() => getMinutes(date),
				(v) => {
					date = setMinutes(date, v);
					onChange?.(date);
				}
			}
		/>
	</div>

	<div class="flex h-full flex-col gap-1 text-xs p-1">
		<button
			class={twMerge(
				'bg-surface3/30 flex-1 rounded-sm px-1',
				isAm && 'bg-primary/10 border-primary/50 text-primary'
			)}
			onclick={() => {
				if (isAm) return;
				date = setHours(date, hours - 12);
			}}>AM</button
		>

		<button
			class={twMerge(
				'bg-surface3/30 flex-1 rounded-sm px-1',
				!isAm && 'text-primary bg-primary/10'
			)}
			onclick={() => {
				if (!isAm) return;
				date = setHours(date, hours + 12);
			}}>PM</button
		>
	</div>
</div>

<style>
	/* For WebKit-based browsers (Chrome, Safari, Edge, Opera) */
	input::-webkit-outer-spin-button,
	input::-webkit-inner-spin-button {
		-webkit-appearance: none; /* Removes the default appearance */
		margin: 0; /* Removes any default margin */
	}

	/* For Mozilla Firefox */
	input[type='number'] {
		appearance: textfield; /* Standard property for compatibility */
		-moz-appearance: textfield; /* Hides the spin buttons in Firefox */
	}
</style>
