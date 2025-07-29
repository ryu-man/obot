<script lang="ts">
	import { twMerge } from 'tailwind-merge';
	import Select, { type SelectProps } from '$lib/components/Select.svelte';
	import type { FilterInput } from './AuditFilters.svelte';
	import type { AuditLog } from '$lib/services';

	interface Props {
		filter: FilterInput;
		onSelect?: SelectProps<AuditLog>['onSelect'];
		onClear?: SelectProps<AuditLog>['onClear'];
	}

	let { filter, onSelect, onClear }: Props = $props();

	let options = $derived(filter.options ?? []);
</script>

<div class={twMerge('mb-2 flex flex-col gap-1', !options.length && 'opacity-50')}>
	<label for={filter.property} class="text-md font-light">
		By {filter.label}
	</label>

	<Select
		class="dark:border-surface3 bg-surface1 border border-transparent shadow-inner dark:bg-black"
		classes={{
			root: 'w-full',
			clear: 'hover:bg-surface3 bg-transparent'
		}}
		{options}
		selected={filter.selected}
		multiple={true}
		{onSelect}
		{onClear}
		position="top"
	/>
</div>
