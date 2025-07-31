<script module>
	export interface SelectProps<T> {
		id?: string;
		disabled?: boolean;
		options: T[];
		selected?: string | number;
		multiple?: boolean;
		onSelect: (option: T, value?: string | number) => void;
		class?: string;
		classes?: {
			root?: string;
			clear?: string;
			option?: string;
			buttonContent?: string;
		};
		position?: 'top' | 'bottom';
		onClear?: (option?: T, value?: string | number) => void;
		buttonStartContent?: Snippet;
	}
</script>

<script lang="ts" generics="T extends { id: string | number; label: string }">
	import { clickOutside } from '$lib/actions/clickoutside';
	import { ChevronDown, X, Check } from 'lucide-svelte';
	import type { Snippet } from 'svelte';
	import { slide } from 'svelte/transition';
	import { twMerge } from 'tailwind-merge';

	let {
		id,
		disabled,
		options,
		onSelect,
		selected = $bindable(),
		multiple = false,
		class: klass,
		classes,
		position = 'bottom',
		onClear,
		buttonStartContent
	}: SelectProps<T> = $props();

	const selectedValues = $derived.by(() => {
		if (multiple) {
			if (typeof selected === 'string') {
				const values =
					selected
						.split(',')
						.map((d) => d.trim())
						.filter(Boolean) ?? [];
				return values;
			}

			if (typeof selected === 'number') {
				return [selected] as number[];
			}

			return [];
		}

		return [selected].filter(Boolean) as (string | number)[];
	});

	let search = $state('');

	let availableOptions = $derived(
		options.filter((option) => option.label.toLowerCase().includes(search.toLowerCase()))
	);

	let selectedOptions = $derived(
		selectedValues
			.map((selectedValue) => options.find((option) => option.id === selectedValue))
			.filter(Boolean)
	);

	$inspect(selectedOptions, selectedValues, selected);

	let popover = $state<HTMLDialogElement>();

	function onInput(e: Event) {
		search = (e.target as HTMLInputElement).value;
	}
</script>

<div class={twMerge('relative', classes?.root)}>
	<div class="relative flex items-center">
		<button
			{id}
			{disabled}
			class={twMerge(
				'dark:bg-surface1 text-md flex min-h-10 w-full grow resize-none items-center justify-between gap-2 rounded-lg bg-white px-2 py-2 text-left shadow-sm',
				disabled && 'cursor-not-allowed opacity-50',
				klass
			)}
			placeholder="Enter a task"
			oninput={onInput}
			onpointerdown={() => {
				if (popover?.open) {
					popover?.close();
				} else {
					popover?.show();
				}
			}}
		>
			<div class="no-scrollbar inset-0 flex flex-1 items-center justify-start overflow-x-scroll">
				<div class="flex items-center justify-center gap-2">
					{#each selectedOptions as selectedOption (selectedOption.id)}
						<div
							class={twMerge(
								'text-md bg-surface3/50 dark:bg-surface2 flex items-center gap-1 truncate rounded-sm px-1',
								onClear && '',
								classes?.buttonContent
							)}
							in:slide={{ duration: 100, axis: 'x' }}
							out:slide={{ duration: 50, axis: 'x' }}
						>
							{#if buttonStartContent}
								{@render buttonStartContent()}
							{/if}
							<di>{selectedOption?.label ?? ''}</di>

							<div
								class={twMerge(
									'button rounded-xs p-0 transition-colors duration-300',
									classes?.clear
								)}
								role="button"
								tabindex="0"
								onclick={(ev) => {
									ev.stopPropagation();

									const filteredValues = selectedValues.filter((d) => d !== selectedOption.id);

									selected = filteredValues.join(',');

									onClear?.(selectedOption, selected);
								}}
								onkeydown={() => {}}
							>
								<X class="size-4" />
							</div>
						</div>
					{/each}
				</div>
			</div>

			<ChevronDown class="size-5 flex-shrink-0" />
		</button>

		{#if !multiple && onClear}
			<button
				class={twMerge(
					'button absolute right-12 top-1/2 -translate-y-1/2 p-1 transition-colors duration-300',
					classes?.clear
				)}
				onclick={() => {
					onClear();
				}}
			>
				<X class="size-4" />
			</button>
		{/if}
	</div>

	<dialog
		use:clickOutside={[
			() => {
				popover?.close();
			},
			true
		]}
		bind:this={popover}
		class={twMerge(
			'default-scrollbar-thin absolute left-0 top-0 z-10 max-h-[300px] w-full overflow-y-auto rounded-sm',
			position === 'top' && 'translate-y-10',
			position === 'bottom' && '-translate-y-full'
		)}
	>
		{#each availableOptions as option (option.id)}
			{@const isSelected = selectedValues.some((d) => d === option.id)}

			<button
				class={twMerge(
					'dark:hover:bg-surface3 hover:bg-surface2 text-md flex w-full items-center px-4 py-2 text-left transition-colors duration-100',
					isSelected && 'dark:bg-surface1 bg-surface2',
					classes?.option
				)}
				onclick={(e) => {
					e.stopPropagation();

					const key = option.id.toString();
					const values = new Set(selectedValues);

					if (isSelected) {
						values.delete(key);
					} else {
						values.add(key);
					}

					selected = values.values().toArray().toReversed().join(',');

					console.log(selected);

					onSelect(option, selected);

					popover?.close();
				}}
			>
				<div>{option.label}</div>

				{#if multiple && isSelected}
					<Check class="ml-auto size-4" />
				{/if}
			</button>
		{/each}
	</dialog>
</div>
