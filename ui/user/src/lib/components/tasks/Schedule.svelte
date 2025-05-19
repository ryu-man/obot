<script lang="ts">
	import type { Schedule } from '$lib/services';
	import Dropdown from '$lib/components/tasks/Dropdown.svelte';
	import { GlobeIcon } from 'lucide-svelte';
	import Combobox from './Combobox.svelte';

	interface Props {
		schedule?: Schedule;
		readOnly?: boolean;
	}

	let { schedule = $bindable(), readOnly }: Props = $props();

	let defaultTimezone = $state(Intl.DateTimeFormat().resolvedOptions().timeZone);
</script>

<div class="flex h-12 items-center self-start">
	<h4 class="text-base font-medium">Schedule</h4>
</div>
<div class="flex flex-col gap-4 md:flex-row">
	<Dropdown
		class="schedule-dropdown max-w-[144px] md:max-w-[204px]"
		values={{
			hourly: 'hourly',
			daily: 'daily',
			weekly: 'weekly',
			monthly: 'monthly'
		}}
		selected={schedule?.interval}
		onSelected={(value) => {
			if (schedule) {
				schedule.interval = value;
			}
		}}
		disabled={readOnly}
	/>

	{#if schedule?.interval === 'hourly'}
		<Combobox
			class="schedule-dropdown max-w-[144px] md:max-w-[204px]"
			type="number"
			values={{
				'0': 'on the hour',
				'15': '15 minutes past',
				'30': '30 minutes past',
				'45': '45 minutes past'
			}}
			selected={schedule?.minute.toString()}
			onSelected={(value) => {
				if (schedule) {
					schedule.minute = parseInt(value);
				}
			}}
			disabled={readOnly}
		/>
	{/if}

	{#if schedule?.interval === 'daily'}
		<Combobox
			class="schedule-dropdown max-w-[144px] md:max-w-[204px]"
			type="number"
			values={{
				'0': 'midnight',
				'3': '3 AM',
				'6': '6 AM',
				'9': '9 AM',
				'12': 'noon',
				'15': '3 PM',
				'18': '6 PM',
				'21': '9 PM'
			}}
			selected={schedule?.hour.toString()}
			onSelected={(value) => {
				if (schedule) {
					schedule.hour = parseInt(value);
				}
			}}
			disabled={readOnly}
		/>
		{#if schedule?.timezone && schedule.timezone !== defaultTimezone}
			<div class="flex items-center gap-1">
				<GlobeIcon class="text-muted-foreground mr-1 h-4 w-4" />
				<span class="text-muted-foreground text-sm">{schedule.timezone}</span>
			</div>
		{/if}
	{/if}

	{#if schedule?.interval === 'weekly'}
		<Dropdown
			class="schedule-dropdown max-w-[144px] md:max-w-[204px]"
			values={{
				'0': 'Sunday',
				'1': 'Monday',
				'2': 'Tuesday',
				'3': 'Wednesday',
				'4': 'Thursday',
				'5': 'Friday',
				'6': 'Saturday'
			}}
			selected={schedule?.weekday.toString()}
			onSelected={(value) => {
				if (schedule) {
					schedule.weekday = parseInt(value);
				}
			}}
			disabled={readOnly}
		/>
		{#if schedule?.timezone && schedule.timezone !== defaultTimezone}
			<div class="flex items-center gap-1">
				<GlobeIcon class="text-muted-foreground mr-1 h-4 w-4" />
				<span class="text-muted-foreground text-sm">{schedule.timezone}</span>
			</div>
		{/if}
	{/if}

	{#if schedule?.interval === 'monthly'}
		<!-- The value of schedule.day is saved on the backend based on zero-based numbering, whereas the user is familliar with one-based numbering -->
		<!-- The solution will be to convert back and forth -->
		<Combobox
			class="schedule-dropdown max-w-[144px] md:max-w-[204px]"
			type="number"
			values={{
				'1': '1st',
				'2': '2nd',
				'3': '3rd',
				'5': '5th',
				'15': '15th',
				'20': '20th',
				'25': '25th',
				'-1': 'last day'
			}}
			selected={(schedule?.day >= 0 ? schedule?.day + 1 : schedule?.day).toString()}
			onSelected={(value) => {
				if (schedule) {
					const valueAsNumber = parseInt(value);

					if (valueAsNumber >= 0) {
						// reset counting to 0-based
						schedule.day = valueAsNumber - 1;
					} else {
						// negative set as is
						schedule.day = valueAsNumber;
					}
				}
			}}
			disabled={readOnly}
		/>
		{#if schedule?.timezone && schedule.timezone !== defaultTimezone}
			<div class="flex items-center gap-1">
				<GlobeIcon class="text-muted-foreground mr-1 h-4 w-4" />
				<span class="text-muted-foreground text-sm">{schedule.timezone}</span>
			</div>
		{/if}
	{/if}
</div>

<style lang="postcss">
	:global(.schedule-dropdown) {
		background-color: var(--surface2);
		font-size: var(--text-md);
		display: flex;
		flex-grow: 1;
	}
</style>
