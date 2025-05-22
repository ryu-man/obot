<script lang="ts">
	import { ChevronDown } from 'lucide-svelte/icons';
	import { popover } from '$lib/actions';
	import { twMerge } from 'tailwind-merge';
	import type { FocusEventHandler, HTMLInputTypeAttribute } from 'svelte/elements';
	import type { Snippet } from 'svelte';

	interface Props {
		class?: string;
		values: Record<string, string>;
		selected?: string;
		disabled?: boolean;
		type?: HTMLInputTypeAttribute | null | undefined;
		children?: Snippet<[]>;
		onselect?: (value: string) => void | Promise<void>;
		onfocus?: FocusEventHandler<HTMLInputElement> | null;
		onblur?: FocusEventHandler<HTMLInputElement> | null;
	}

	const popoverController = popover({
		placement: 'bottom-start'
	});

	const { ref, tooltip, toggle } = popoverController;

	let {
		values,
		selected,
		disabled = false,
		class: kclass = '',
		type = 'text',
		onselect,
		onfocus,
		onblur,
		children
	}: Props = $props();

	async function select(value: string) {
		await onselect?.(value);
		toggle();
	}
</script>

{#if disabled}
	<span
		class={twMerge(
			'combobox text-gray flex items-center justify-between gap-2 rounded-3xl p-3 px-4 capitalize dark:hover:bg-gray-900',
			kclass
		)}
	>
		{selected ? values[selected] : values[''] || ''}
		<ChevronDown class="text-gray" />
	</span>
{:else}
	<button
		use:ref
		onclick={() => {
			toggle();
		}}
		class={twMerge(
			'hover:bg-gray-70 relative flex items-center justify-between gap-2 rounded-3xl px-4 capitalize focus-within:outline dark:hover:bg-gray-900',
			kclass
		)}
	>
		<input
			class="w-full min-w-0 flex-1 bg-transparent py-3 outline-none"
			{type}
			bind:value={
				() => selected ?? '',
				(v) => {
					onselect?.(v);
				}
			}
			onclick={(ev) => {
				if (popoverController.open) {
					ev.stopPropagation();
				}
			}}
			{onfocus}
			{onblur}
		/>
		{@render children?.()}
		<ChevronDown class="z-10" />
	</button>
	<div use:tooltip class="min-w-[150px] rounded-3xl bg-white shadow dark:bg-gray-900">
		<ul>
			{#each Object.keys(values) as key}
				{@const value = values[key]}
				<li>
					<button
						class:bg-gray-70={selected === key}
						class:dark:bg-gray-800={selected === key}
						class="w-full px-6 py-2.5 text-start capitalize hover:bg-gray-100 dark:hover:bg-gray-800"
						onclick={() => select(key)}
					>
						{value}
					</button>
				</li>
			{/each}
		</ul>
	</div>
{/if}

<style lang="postcss">
	li:first-child button {
		border-top-left-radius: 1.5rem;
		border-top-right-radius: 1.5rem;
		padding-top: 1rem;
	}
	li:last-child button {
		border-bottom-left-radius: 1.5rem;
		border-bottom-right-radius: 1.5rem;
		padding-bottom: 1rem;
	}

	input::-webkit-outer-spin-button,
	input::-webkit-inner-spin-button {
		-webkit-appearance: none;
		margin: 0;
	}

	input[type='number'] {
		-moz-appearance: textfield;
	}
</style>
