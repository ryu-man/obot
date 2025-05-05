<script lang="ts">
	import Self from './Step.svelte';
	import {
		ChatService,
		type Messages,
		type Project,
		type Task,
		type TaskStep
	} from '$lib/services';
	import Message from '$lib/components/messages/Message.svelte';
	import { Eye, EyeClosed, Plus, Trash2, Repeat } from 'lucide-svelte/icons';
	import {
		LoaderCircle,
		OctagonX,
		Play,
		RefreshCcw,
		ChevronLeft,
		ChevronRight
	} from 'lucide-svelte';
	import { tick } from 'svelte';
	import { autoHeight } from '$lib/actions/textarea.js';
	import Confirm from '$lib/components/Confirm.svelte';
	import { fade, slide } from 'svelte/transition';
	import { tooltip } from '$lib/actions/tooltip.svelte';

	interface Props {
		parentStale?: boolean;
		run?: (step: TaskStep) => Promise<void>;
		task: Task;
		index: number;
		step: TaskStep;
		runID?: string;
		pending?: boolean;
		stepMessages?: Map<string, Messages>;
		project: Project;
		showOutput?: boolean;
		readOnly?: boolean;
	}

	let {
		parentStale,
		run,
		task = $bindable(),
		index,
		step = $bindable(),
		runID,
		pending,
		stepMessages,
		project,
		showOutput: parentShowOutput,
		readOnly
	}: Props = $props();

	let isRunning = $derived(stepMessages?.get(step.id)?.inProgress ?? false);
	let isRunnedBefore = $derived(!!stepMessages?.get(step.id)?.lastRunID);
	let stale: boolean = $derived(parentStale || !parentMatches());
	let toDelete = $state<boolean>();
	let showOutput = $state(true);

	// Check whether the current step has looping steps (sub steps)
	// It should have the
	let isLoopStep = $derived((step?.loop?.length ?? 0) > 0);

	let messages = $derived(stepMessages?.get(step.id)?.messages ?? []);

	let loopDataMessages = $derived(stepMessages?.get(step.id + '{loopdata}')?.messages ?? []);

	let currentIteration = $state(0);
	let shouldFollowIteration = $state(true);

	type Iteration = Messages[];

	// Convert the steps messages map to an array of messages where each index represent the number of iteration
	let iterations: Iteration[] = $derived.by(() => {
		// Convert the keys into an array
		const keys = stepMessages?.keys().toArray() ?? [];

		// Define a regex pattern to extract iterations data
		const pattern = new RegExp(`^${step.id}{element=(\\d+)}`);

		// Initialize the iterations array
		const iterations: Iteration[] = [];

		keys
			// Filter out not matched items
			.filter((key) => pattern.test(key))

			.forEach((key) => {
				// Get the iteration number as a string
				const iterationAsString = key.match(pattern)?.at(1);

				if (iterationAsString === undefined) {
					return;
				}

				// Convert the iteration number to an integer
				const iteration = parseInt(iterationAsString);

				// Push the step messages to the same iteration array
				const steps = iterations.at(iteration) ?? [];
				const messages = stepMessages?.get(key);

				steps.push(messages!);

				iterations[iteration] = steps;
			});

		return iterations;
	});

	$effect(() => {
		if (parentShowOutput !== undefined) {
			showOutput = parentShowOutput;
		}
	});

	$effect(() => {
		// Check if task is running
		// If following iteration is true; then automatically navigate to the last iteration
		if (isRunning && shouldFollowIteration) {
			// Always navigation to the last iteration
			currentIteration = iterations.length - 1;
		}
	});

	function parentMatches() {
		if (isRunning) {
			return true;
		}
		if (index === 0) {
			return true;
		}
		const lastRun = stepMessages
			?.get(task.steps[index - 1].id)
			?.messages.findLast((msg) => msg.runID);
		const currentRun = stepMessages
			?.get(task.steps[index].id)
			?.messages.find((msg) => msg.parentRunID);
		return lastRun?.runID === currentRun?.parentRunID;
	}

	async function toggleLoop() {
		if (isLoopStep) {
			step.loop = undefined;
		} else {
			step.loop = [''];
		}
	}

	async function deleteStep() {
		task.steps = task.steps.filter((s) => s.id !== step.id);
	}

	async function addStep() {
		const newStep = createStep();
		task.steps.splice(index + 1, 0, newStep);
		await tick();

		document.getElementById('step' + newStep.id)?.focus();
	}

	async function onkeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.ctrlKey && !e.shiftKey) {
			e.preventDefault();
			await doRun();
		} else if (e.key === 'Enter' && e.ctrlKey && !e.shiftKey) {
			e.preventDefault();
			await addStep();
		}
	}

	function createStep(): TaskStep {
		return { id: Math.random().toString(36).substring(7), step: '' };
	}

	async function doRun() {
		if (isRunning || pending) {
			if (runID) {
				await ChatService.abort(project.assistantID, project.id, {
					taskID: task.id,
					runID: runID
				});
			}
			return;
		}
		if (isRunning || pending || !step.step || step.step?.trim() === '') {
			return;
		}

		// By default follow iteration when step is running
		shouldFollowIteration = true;

		await run?.(step);
	}

	function onclickNextIteration() {
		shouldFollowIteration = false;
		currentIteration = Math.min(iterations.length - 1, currentIteration + 1);

		// When user navigate back to the last iteration; activate following
		if (currentIteration === iterations.length - 1) {
			shouldFollowIteration = true;
		}
	}

	function onclickPreviousIteration() {
		shouldFollowIteration = false;
		currentIteration = Math.max(0, currentIteration - 1);
	}
</script>

{#snippet outputVisibilityButton()}
	<div class="size-10">
		{#if messages.length > 0}
			<button
				class="icon-button"
				data-testid="step-toggle-output-btn"
				onclick={() => (showOutput = !showOutput)}
				use:tooltip={'Toggle Output Visibility'}
				transition:fade={{ duration: 200 }}
			>
				{#if showOutput}
					<Eye class="size-4" />
				{:else}
					<EyeClosed class="size-4" />
				{/if}
			</button>
		{/if}
	</div>
{/snippet}

<li class="ms-4">
	<div class="flex items-start justify-between gap-6">
		<div class="flex grow flex-col gap-2">
			<div class="flex items-center gap-2">
				<textarea
					{onkeydown}
					rows="1"
					placeholder={isLoopStep ? 'Description of the data to loop over...' : 'Instructions...'}
					use:autoHeight
					id={'step' + step.id}
					bind:value={step.step}
					class="ghost-input border-surface2 ml-1 grow resize-none"
					disabled={readOnly}
				></textarea>
			</div>

			{#if isLoopStep}
				{#if loopDataMessages.length > 0 && showOutput}
					<!-- Show step message -->
					<div
						class="relative my-3 -ml-4 flex min-h-[150px] flex-col gap-4 rounded-lg bg-white p-5 transition-transform dark:bg-black"
						class:border-2={isRunning}
						class:border-blue={isRunning}
						transition:slide
					>
						{#each loopDataMessages as msg}
							{#if !msg.sent}
								<Message {msg} {project} disableMessageToEditor />
							{/if}
						{/each}
						{#if stale}
							<div
								class="absolute inset-0 h-full w-full rounded-3xl bg-white opacity-80 dark:bg-black"
							></div>
						{/if}
					</div>
				{/if}

				<div class="iterations-container flex flex-col gap-4">
					{#if (isRunning || isRunnedBefore || readOnly) && iterations.length && showOutput}
						<!-- Display the iterations header only in case of task is running or was runned before or in read-only mode -->
						<div class="iterations-header flex justify-between">
							<div class="flex items-baseline gap-4 opacity-50">
								<div>Iterations:</div>

								<div class="text-sm">
									<span>{currentIteration + 1}</span>
									<span class="opacity-50">/ {iterations.length}</span>
								</div>
							</div>

							<!-- Show iterations navigation button -->
							<div class="flex gap-2">
								<button
									class="flex aspect-square h-8 items-center justify-center rounded-md bg-black transition-colors duration-200 hover:bg-black/90 active:bg-black/80"
									disabled={currentIteration <= 0}
									onclick={onclickPreviousIteration}
								>
									<ChevronLeft class="h-5 opacity-50" />
								</button>

								<button
									class="flex aspect-square h-8 items-center justify-center rounded-md bg-black transition-colors duration-200 hover:bg-black/90 active:bg-black/80"
									disabled={currentIteration >= iterations.length - 1}
									onclick={onclickNextIteration}
								>
									<ChevronRight class="h-5 opacity-50" />
								</button>
							</div>
						</div>
					{/if}

					<div class="iterations-body flex flex-col gap-2 pl-6">
						{#each step.loop! as _, i}
							<!-- Get the current iteration steps messages array -->
							{@const messages = iterations[currentIteration] ?? []}

							<!-- Get the current step messages array -->
							{@const stepMessages = messages[i] ?? []}

							<div class="iteration-step flex flex-col gap-2">
								<div class="flex items-center gap-2">
									<textarea
										{onkeydown}
										rows="1"
										placeholder="Instructions..."
										use:autoHeight
										bind:value={step.loop![i]}
										class="ghost-input border-surface2 grow resize-none"
										disabled={readOnly}
									></textarea>

									{#if !readOnly}
										<button
											class="icon-button"
											onclick={() => step.loop!.splice(i, 1)}
											use:tooltip={'Remove step from loop'}
										>
											<Trash2 class="size-4" />
										</button>
									{/if}
								</div>

								{#if stepMessages.messages?.length > 0 && showOutput}
									<div
										class="relative my-3 -ml-4 flex min-h-[150px] flex-col gap-4 rounded-lg bg-white p-5 transition-transform dark:bg-black"
										class:border-2={isRunning}
										class:border-blue={isRunning}
										transition:slide
									>
										{#each stepMessages.messages as msg}
											{#if !msg.sent}
												<Message {msg} {project} disableMessageToEditor />
											{/if}
										{/each}
										{#if stale}
											<div
												class="absolute inset-0 h-full w-full rounded-3xl bg-white opacity-80 dark:bg-black"
											></div>
										{/if}
									</div>
								{/if}
							</div>
						{/each}
						{#if !readOnly}
							<button
								class="icon-button self-start"
								onclick={() => step.loop!.push('')}
								use:tooltip={'Add step to loop'}
							>
								<Plus class="size-4" />
							</button>
						{/if}
					</div>
				</div>
			{/if}
		</div>

		<div class="flex items-start">
			{#if !readOnly}
				<button
					class="icon-button"
					class:text-blue={isLoopStep}
					data-testid="step-loop-btn"
					onclick={toggleLoop}
					use:tooltip={isLoopStep ? 'Convert to regular step' : 'Convert to loop step'}
				>
					<Repeat class="size-4" />
				</button>

				<button
					class="icon-button"
					data-testid="step-run-btn"
					onclick={doRun}
					use:tooltip={isRunning
						? 'Abort'
						: pending
							? 'Running...'
							: messages.length > 0
								? 'Re-run Step'
								: 'Run Step'}
				>
					{#if isRunning}
						<OctagonX class="size-4" />
					{:else if pending}
						<LoaderCircle class="size-4 animate-spin" />
					{:else if messages.length > 0}
						<RefreshCcw class="size-4" />
					{:else}
						<Play class="size-4" />
					{/if}
				</button>
				<button
					class="icon-button"
					data-testid="step-delete-btn"
					onclick={() => {
						if (step.step?.trim()) {
							toDelete = true;
						} else {
							deleteStep();
						}
					}}
					use:tooltip={'Delete Step'}
				>
					<Trash2 class="size-4" />
				</button>
				<div class="flex grow">
					<div class="size-10">
						{#if (step.step?.trim() || '').length > 0}
							<button
								class="icon-button"
								data-testid="step-add-btn"
								onclick={addStep}
								use:tooltip={'Add Step'}
								transition:fade={{ duration: 200 }}
							>
								<Plus class="size-4" />
							</button>
						{/if}
					</div>
					{@render outputVisibilityButton()}
				</div>
			{/if}
		</div>
	</div>
</li>

<!-- This code section is responsible for showing messages in a !loop task -->
{#if !isLoopStep && messages.length > 0 && showOutput}
	<div
		class="relative my-3 -ml-4 flex min-h-[150px] flex-col gap-4 rounded-lg bg-white p-5 transition-transform dark:bg-black"
		class:border-2={isRunning}
		class:border-blue={isRunning}
		transition:slide
	>
		{#each messages as msg}
			{#if !msg.sent}
				<Message {msg} {project} disableMessageToEditor />
			{/if}
		{/each}
		{#if stale}
			<div
				class="absolute inset-0 h-full w-full rounded-3xl bg-white opacity-80 dark:bg-black"
			></div>
		{/if}
	</div>
{/if}

<!-- This code section shows other task steps recursively; -->
<!-- REFACTOR: This should be moved out to the steps component and render steps in an each loop, Step.svelte should only be responsibe for displaying a single step -->
{#if task.steps.length > index + 1}
	{#key task.steps[index + 1].id}
		<Self
			{run}
			{runID}
			{pending}
			{task}
			index={index + 1}
			bind:step={task.steps[index + 1]}
			{stepMessages}
			parentStale={stale}
			{project}
			showOutput={parentShowOutput}
			{readOnly}
		/>
	{/key}
{/if}

<!-- This code section show dialog to confirm task delete -->
<!-- REFACTOR: Move out to the Steps.svelte component; having one dialog shared with many steps is better than each steps has its own dialog-->
<Confirm
	show={toDelete !== undefined}
	msg={`Are you sure you want to delete this step`}
	onsuccess={deleteStep}
	oncancel={() => (toDelete = undefined)}
/>
