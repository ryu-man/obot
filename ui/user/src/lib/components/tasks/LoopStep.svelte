<script lang="ts">
	import type { Messages, Project } from '$lib/services';
	import Message from '$lib/components/messages/Message.svelte';
	import { Trash2 } from 'lucide-svelte/icons';
	import { autoHeight } from '$lib/actions/textarea.js';
	import { tooltip } from '$lib/actions/tooltip.svelte';
	import type { KeyboardEventHandler } from 'svelte/elements';
	import { transitionParentHeight } from '$lib/actions/size.svelte';
	import { slide } from 'svelte/transition';

	type Props = {
		value: string;
		messages: Messages;
		project: Project;
		isLoopStepRunning?: boolean;
		isStepRunning?: boolean;
		isStepRunned?: boolean;
		isReadOnly?: boolean;
		shouldShowOutput?: boolean;
		stale?: boolean;
		onkeydown?: KeyboardEventHandler<HTMLTextAreaElement>;
		ondelete?: () => void;
	};

	let {
		value = $bindable(),
		messages,
		project,
		isLoopStepRunning = false,
		isStepRunning = false,
		isStepRunned = false,
		isReadOnly = false,
		shouldShowOutput = false,
		stale = false,
		onkeydown = undefined,
		ondelete = undefined
	}: Props = $props();
</script>

<div
	class="iteration-step flex flex-col gap-2 transition-opacity duration-100"
	class:opacity-50={isStepRunning && !isLoopStepRunning}
>
	<div class="flex items-center gap-2 overflow-hidden">
		<textarea
			use:autoHeight
			bind:value
			rows="1"
			placeholder="Instructions..."
			class="ghost-input border-surface2 h-auto grow resize-none"
			disabled={isReadOnly}
			{onkeydown}
		></textarea>

		{#if !isReadOnly}
			<button class="icon-button" onclick={ondelete} use:tooltip={'Remove step from loop'}>
				<Trash2 class="size-4" />
			</button>
		{/if}
	</div>

	{#if (isStepRunning || isStepRunned) && shouldShowOutput}
		<div
			class="transition-height relative my-3 -ml-4 box-content flex min-h-[96px] flex-col gap-4 rounded-lg bg-white p-5 transition-all duration-100 dark:bg-black"
			class:border-2={isStepRunning && isLoopStepRunning}
			class:border-blue={isStepRunning && isLoopStepRunning}
			transition:slide
		>
			{#if messages.messages?.length > 0}
				<div
					class="messages-list flex w-full flex-col gap-4"
					use:transitionParentHeight={() => (isStepRunning && shouldShowOutput) || messages.messages}
				>
					{#each messages.messages as msg}
						{#if !msg.sent}
							<Message {msg} {project} disableMessageToEditor />
						{/if}
					{/each}
				</div>

				{#if stale}
					<div
						class="absolute inset-0 h-full w-full rounded-3xl bg-white opacity-80 dark:bg-black"
					></div>
				{/if}
			{/if}
		</div>
	{/if}
</div>
