<script lang="ts">
	import type { ModelProvider, Project } from '$lib/services';
	import { CheckCircleIcon, Loader2, Search } from 'lucide-svelte';
	import { toHTMLFromMarkdown } from '$lib/markdown';
	import { darkMode } from '$lib/stores';
	import {
		updateProject,
		listAvailableModels,
		configureModelProvider,
		deconfigureModelProvider,
		getModelProviderConfig
	} from '$lib/services/chat/operations';
	import { twMerge } from 'tailwind-merge';
	import { fade, slide } from 'svelte/transition';
	import { delay, throttle } from 'es-toolkit';
	import { untrack } from 'svelte';

	type Props = {
		provider: ModelProvider;
		project: Project;
		configuringProvider?: string | null;
		availableModels: string[];
		onError: (error: string) => void;
	};

	let {
		provider,
		project = $bindable(),
		configuringProvider = $bindable(),
		availableModels: models = [],
		onError
	}: Props = $props();

	let selectedModels: string[] = $derived(project.models?.[provider.id] ?? []);
	let error: string | null = $state(null);

	let defaultModelProvider = $derived(project.defaultModelProvider || '');
	let defaultModel = $derived(project.defaultModel || '');

	let isConfigured = $derived(provider?.configured ?? false);
	let providerId = $derived(provider.id);
	// let selectedModels = $derived<Record<string, string[]>>(project.models || {});

	let configuration: Record<string, string> = $state({});
	// svelte-ignore state_referenced_locally
	let oldConfiguration: Record<string, string> = $state(
		structuredClone($state.snapshot(configuration))
	);

	let isConfigurationLoading = $state(false);
	let isModelsLoading = $state(false);
	let isSaving = $state(false);
	let isProviderConfigurationShown = $state(false);

	let modelQuery: string = $state('');
	let filteredModels: string[] = $state([]);

	// Throttle search model for better performance
	const searchModels = throttle((query: string, models: string[] = []) => {
		filteredModels = models.filter((model) =>
			model.trim().toLowerCase().includes(query.trim().toLowerCase())
		);
	}, 100);

	$effect(() => {
		searchModels(modelQuery, models);
	});

	// When a provider is selected for configuration, load its config
	$effect(() => {
		if (isConfigured) {
			untrack(async () => {
				try {
					await loadProviderConfig(providerId);
					await delay(1000);
					models = await loadAvailableModels(providerId);
				} catch (err) {
					console.error(err);
				}
			});
		} else {
			models = [];
		}
	});

	// Deconfigure model provider
	async function handleDeconfigureModelProvider(provider: ModelProvider) {
		try {
			// Unselect all model first
			await unselectModels(selectedModels);
			// Unconfigure model provider
			await deconfigureModelProvider(project.assistantID, project.id, provider.id);

			models = [];
			provider.configured = false;
			delete project.models?.[provider.id];

			// clear current configurations
			configuration = {};
			// clear previous configurations
			oldConfiguration = {};
		} catch (err) {
			console.error(`Failed to deconfigure ${provider.name}`, err);
		}
	}

	type UpdateProjectParams = {
		models: Project['models'];
		defaultModel?: Project['defaultModel'];
		defaultModelProvider?: Project['defaultModelProvider'];
	};
	// Update the project's model selection directly and trigger auto-save
	function updateProject_(params: UpdateProjectParams) {
		project = {
			...project,
			...params
		};

		saveChanges();
	}

	// Load configuration for a model provider
	async function loadProviderConfig(providerId: string) {
		try {
			configuration = await getModelProviderConfig(project.assistantID, project.id, providerId);
			oldConfiguration = structuredClone($state.snapshot(configuration));
		} catch (err) {
			console.error(`Failed to get configuration for provider ${providerId}`, err);
		} finally {
			isConfigurationLoading = false;
		}
	}

	// Configure model provider
	async function handleConfigureModelProvider(
		provider: ModelProvider,
		config: Record<string, string>
	) {
		try {
			await configureModelProvider(project.assistantID, project.id, provider.id, config);
			const newProject = setProjectModels(project, providerId, []);

			await updateProject(newProject);

			project = newProject;
		} catch (err) {
			onError?.((error = `Failed to configure ${provider.name}`));

			console.error(error, err);
		}
	}

	// Load available models for a provider
	async function loadAvailableModels(providerId: string) {
		isModelsLoading = true;

		try {
			const response = await listAvailableModels(project.assistantID, project.id, providerId);

			return (response.data || [])
				.filter((m) => m.metadata && m.metadata.usage === 'llm')
				.map((m) => m.id)
				.sort((a, b) => a.localeCompare(b));
		} catch (err) {
			console.error(`Failed to load models for provider ${providerId}`, err);
			return [];
		} finally {
			await delay(500);
			isModelsLoading = false;
		}
	}

	async function saveHandler(ev: Event) {
		ev.preventDefault();

		try {
			isModelsLoading = true;
			await handleConfigureModelProvider(provider, configuration);

			provider.configured = true;

			await delay(400);

			const array = await loadAvailableModels(provider.id);

			isProviderConfigurationShown = false;

			await delay(300);

			models = array;
		} catch (err) {
			console.log(err);
			provider.configured = false;
		} finally {
			isModelsLoading = false;
		}
	}

	// Toggle model selection
	function toggleModel(model: string, isChecked: boolean) {
		// Take a snapshot of current models
		let projectModels = { ...(project.models || {}) };
		const currentProviderModels = new Set([...(projectModels[provider.id] || [])]);

		if (isChecked) {
			currentProviderModels.add(model);
		} else {
			currentProviderModels.delete(model);
		}

		let p = project;

		if (!isChecked && defaultModelProvider === provider.id) {
			p = setProjectDefaultProvider(p);
		}

		if (!isChecked && defaultModel === model) {
			p = setProjectDefaultModel(p, model);
		}

		p = setProjectModels(p, provider.id, currentProviderModels.values().toArray());

		project = p;
	}

	function selectModels(models: string[]) {
		// Take a snapshot of current models
		let projectModels = { ...(project.models || {}) };

		projectModels = {
			...projectModels,
			[provider.id]: models
		};

		return updateProject_({ models: projectModels });
	}

	function unselectModels(models: string[]) {
		// Take a snapshot of current models
		let projectModels = { ...(project.models || {}) };

		const array = new Set(projectModels[providerId] ?? []);

		for (const model of models) {
			array.delete(model);
		}

		projectModels = {
			...projectModels,
			[provider.id]: array.values().toArray()
		};

		return updateProject_({ models: projectModels });
	}

	function setProjectDefaultProvider(project: Project, providerId?: string) {
		const clone = structuredClone($state.snapshot(project));

		if (provider) {
			clone.defaultModelProvider = providerId;
		} else {
			delete clone.defaultModelProvider;
		}

		return clone;
	}

	function setProjectDefaultModel(project: Project, modelId: string) {
		const clone = structuredClone($state.snapshot(project));

		if (provider) {
			clone.defaultModel = modelId;
		} else {
			delete clone.defaultModel;
		}

		return clone;
	}

	function setProjectModels(project: Project, providerId: string, models: string[]) {
		const clone = structuredClone($state.snapshot(project));

		if (!clone.models) {
			clone.models = {};
		}

		if (!models) {
			delete clone.models[providerId];
		} else {
			clone.models[providerId] = models;
		}

		return clone;
	}

	// Save changes to the server
	async function saveChanges() {
		try {
			isSaving = true;
			// isSaving = true;
			// Update the project prop directly
			project = await updateProject(project);
		} catch (err) {
			console.error('Failed to save model configuration', err);
			error = 'Failed to save changes. Please try again.';
		} finally {
			await delay(500);
			isSaving = false;
		}
	}

	function onConfigureProviderClickHandler() {
		configuringProvider = provider.id;
		isProviderConfigurationShown = true;
	}
</script>

<div
	class="model-provider-card border-surface2 flex h-fit w-full flex-col gap-2 rounded-md border py-4 shadow-sm 2xl:mb-4 2xl:last:mb-0"
>
	<div class="flex flex-col px-4">
		<div class="mb-2 flex items-center gap-2">
			{#if provider.icon || provider.iconDark}
				<img
					src={darkMode.isDark && provider.iconDark ? provider.iconDark : provider.icon}
					alt={provider.name}
					class="h-6 w-6 {darkMode.isDark && !provider.iconDark ? 'dark:invert' : ''}"
				/>
			{/if}
			<div class="flex items-center gap-4">
				<h4 class="truncate text-lg font-medium">{provider.name}</h4>
				{#if [isModelsLoading, isSaving, isConfigurationLoading].some(Boolean)}
					<div
						class="flex justify-center"
						in:fade={{ duration: 1000 }}
						out:fade={{ duration: 600 }}
					>
						<Loader2 class="size-5 animate-spin" />
					</div>
				{/if}
			</div>

			{#if provider.configured}
				<CheckCircleIcon class="text-blue ml-auto aspect-square h-5" />
			{/if}
		</div>

		{#if provider.description}
			<div class="markdown-content text-gray text-sm">
				{@html toHTMLFromMarkdown(provider.description)}
			</div>
		{:else}
			<div class="bg-surface1 text-gray flex w-full items-center rounded-lg px-4 py-3 text-sm">
				<p>Description is not available</p>
			</div>
		{/if}
	</div>

	<div class="flex flex-col gap-4 px-4 pb-4 pt-4">
		{#if isConfigured}
			<!-- Models Selection Section -->
			<div
				class={twMerge(
					'flex flex-col',
					isProviderConfigurationShown && 'pointer-events-none opacity-50'
				)}
				transition:slide={{ duration: 100 }}
			>
				{#if models.length}
					<div class="bg-surface1/0 flex flex-col gap-2 rounded-md">
						<div
							class="model-provider-search-input bg-surface1 dark:bg-surface2 relative flex h-10 w-full items-center gap-2 rounded-md text-sm"
						>
							<div class="absolute inset-x-0 left-2 aspect-square h-5 opacity-50">
								<Search class="h-full" />
							</div>
							<input
								bind:value={modelQuery}
								class="h-full w-full bg-transparent px-3 pl-9 pr-8"
								type="text"
								placeholder="Search your model here..."
							/>

							{#if filteredModels.length !== models.length}
								<div
									class="absolute inset-y-0 right-4 flex h-full items-center justify-center text-sm font-medium opacity-50"
									transition:fade={{ duration: 100 }}
								>
									<div>{filteredModels.length}</div>
								</div>
							{/if}
						</div>

						<div class="mb-2 flex justify-between">
							<div class="flex items-center gap-2 px-3 text-sm font-medium">
								<h5 class="flex items-center">
									<input
										class="bg-surface3 mr-2 h-4 w-4"
										type="checkbox"
										bind:checked={
											() => models.length === selectedModels.length,
											(v) => {
												const indeterminate =
													selectedModels.length > 0 && models.length > selectedModels.length;

												if (indeterminate || v) {
													selectModels(models);
												} else {
													unselectModels(selectedModels);
												}
											}
										}
										indeterminate={selectedModels.length > 0 &&
											models.length > selectedModels.length}
									/>
									<label
										class="inline hover:underline"
										onclick={() => {
											if (selectedModels.length) {
												unselectModels(selectedModels);
											} else {
												selectModels(models);
											}
										}}>Available Models</label
									>
									{#if models.length}
										<span class="opacity-50">({models.length})</span>
									{/if}
								</h5>

								{#if selectedModels.length}
									<div class="h-full border-l"></div>
									<button class="inline font-normal">
										<span class="">Selected</span>
										<span class="opacity-50">({selectedModels.length})</span>
									</button>
								{/if}
							</div>
						</div>

						<div
							class="default-scrollbar-thin scrollbar-track-rounded-full max-h-48 overflow-y-auto pr-2"
						>
							{#each filteredModels as model, i (model)}
								<div class="hover:bg-surface1 bored flex items-center rounded px-3 py-2">
									<input
										class="bg-surface3 mr-2 h-4 w-4"
										type="checkbox"
										id={`model-${provider.id}-${model}`}
										bind:checked={
											() => (selectedModels ?? []).includes(model),
											(checked) => toggleModel(model, checked)
										}
									/>

									<label
										for={`model-${provider.id}-${model}`}
										class="flex-1 cursor-pointer select-none truncate text-sm"
									>
										{model}
										<!-- {#if defaultModelProvider === provider.id && defaultModel === model}
									<span class="text-primary ml-2 text-xs font-medium">(Default Model)</span>
								{/if} -->
									</label>
								</div>
							{:else}
								<p class="text-muted text-sm">No models available for this provider.</p>
							{/each}
						</div>
					</div>
				{/if}
			</div>
		{/if}

		{#if isProviderConfigurationShown}
			<div transition:slide={{ duration: 200 }} class="flex flex-col gap-2">
				<div class="flex items-center gap-2">
					<div class="flex items-center font-medium">
						<span>Configurations</span>
					</div>
					<div class="border-surface2 flex-1 border-b"></div>
				</div>

				<div class="flex flex-col gap-4">
					{#each provider.requiredConfigurationParameters || [] as param}
						<div class="flex w-full flex-col gap-1">
							<label class="mb-1 block truncate text-sm font-medium" for={param.name}>
								{param.friendlyName || param.name}
								{#if param.description}
									<span class="text-muted text-xs">({param.description})</span>
								{/if}
							</label>

							<input
								type={param.sensitive ? 'password' : 'text'}
								id={param.name}
								class="w-full rounded-md border p-2 text-sm"
								bind:value={configuration[param.name]}
								required
							/>
						</div>
					{/each}
				</div>
			</div>
		{/if}
	</div>

	<div class="border-surface2 flex w-full justify-between gap-4 border-t px-4 pt-4">
		{#if isProviderConfigurationShown}
			{@const isDirty =
				JSON.stringify($state.snapshot(configuration)) !==
				JSON.stringify($state.snapshot(oldConfiguration))}

			<div class="flex gap-2">
				{#if isDirty}
					<button
						in:fade={{ duration: 100 }}
						out:fade={{ duration: 0 }}
						type="button"
						class="button hover:bg-surface1 rounded-full px-4 py-2 text-sm transition-colors duration-100"
						onclick={() => {
							configuration = $state.snapshot(oldConfiguration);
						}}
					>
						Reset
					</button>
				{:else}
					<button
						in:fade={{ duration: 100 }}
						out:fade={{ duration: 0 }}
						type="button"
						class={twMerge(
							'button hover:bg-surface1 rounded-full px-4 py-2 text-sm transition-colors duration-100'
						)}
						onclick={() => {
							// Close configuration UI
							isProviderConfigurationShown = false;
						}}
					>
						Cancel
					</button>
				{/if}

				{#if isDirty}
					<button
						in:fade={{ duration: 100 }}
						out:fade={{ duration: 0 }}
						type="submit"
						class={twMerge(
							'button bg-blue/10 text-blue hover:bg-blue/15 active:bg-blue/20 rounded-full border-none px-4 py-2 text-sm transition-colors duration-100'
						)}
						disabled={!isDirty}
						onclick={(ev) => saveHandler(ev)}
					>
						Save
					</button>
				{/if}
			</div>
		{/if}

		<div class="ml-auto flex gap-2">
			{#if provider.configured}
				<button
					class="button bg-red-500/0 text-sm font-medium text-red-500 transition-colors duration-100 hover:bg-red-500/10 active:bg-red-500/15"
					onclick={() => {
						if (confirm(`Are you sure you want to deconfigure ${provider.name}?`)) {
							handleDeconfigureModelProvider(provider);
						}
					}}
				>
					Unconfigure
				</button>
			{/if}
			{#if !isProviderConfigurationShown}
				<button class="button text-sm font-medium" onclick={onConfigureProviderClickHandler}>
					{provider.configured ? 'Reconfigure' : 'Configure'}
				</button>
			{/if}
		</div>
	</div>
</div>

<style>
	.model-provider-card {
		break-inside: avoid;
	}
	.markdown-content :global(a) {
		text-decoration: underline;
		font-weight: 500;
	}
	.markdown-content :global(a:hover) {
		opacity: 0.8;
	}
	.markdown-content :global(p) {
		margin-bottom: 0.5rem;
	}
	.markdown-content :global(ul),
	.markdown-content :global(ol) {
		margin-left: 1.5rem;
		margin-bottom: 0.5rem;
	}
	.markdown-content :global(ul) {
		list-style-type: disc;
	}
	.markdown-content :global(ol) {
		list-style-type: decimal;
	}
	.markdown-content :global(code) {
		font-family: monospace;
		background-color: rgba(0, 0, 0, 0.1);
		padding: 0.125rem 0.25rem;
		border-radius: 0.25rem;
	}
	.markdown-content :global(h1),
	.markdown-content :global(h2),
	.markdown-content :global(h3),
	.markdown-content :global(h4),
	.markdown-content :global(h5),
	.markdown-content :global(h6) {
		font-weight: 600;
		margin-top: 0.5rem;
		margin-bottom: 0.5rem;
	}
</style>
