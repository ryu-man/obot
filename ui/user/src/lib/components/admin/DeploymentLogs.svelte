<script lang="ts">
	import { AlertTriangle, Maximize, Minimize, RefreshCw, Search, X } from 'lucide-svelte';
	import { fade } from 'svelte/transition';
	import { tooltip } from '$lib/actions/tooltip.svelte';
	import { twMerge } from 'tailwind-merge';

	interface Props {
		messages: string[];
		error?: string;
		refreshing?: boolean;
		onRefresh?: () => void;
		onClear?: () => void;
		title?: string;
		showRefresh?: boolean;
	}

	const {
		messages = $bindable([]),
		error,
		refreshing = false,
		onRefresh,
		onClear,
		title = 'Deployment Logs',
		showRefresh = true
	}: Props = $props();

	let logsContainer: HTMLDivElement;
	let modalContainer: HTMLDivElement;
	let isMaximized = $state(false);
	let query = $state('');
	let userScrolledUp = $state(false);
	let searchInput = $state<HTMLInputElement>();

	let filteredMessages = $derived.by(() => {
		if (!query) return messages;
		return messages.filter((msg) => msg.toLowerCase().includes(query.toLowerCase()));
	});

	const hasMessages = $derived(messages.length > 0);
	const hasResults = $derived(filteredMessages.length > 0);

	$effect(() => {
		if (!messages.length) return;

		// Auto-scroll to bottom when new messages arrive, unless user scrolled up
		if (logsContainer && !userScrolledUp) {
			setTimeout(() => {
				if (!userScrolledUp) scrollToBottom(logsContainer);
			}, 50);
		}
	});

	function isScrolledToBottom(element: HTMLElement): boolean {
		return Math.abs(element.scrollHeight - element.clientHeight - element.scrollTop) < 10;
	}

	function scrollToBottom(element: HTMLElement) {
		element.scrollTop = element.scrollHeight;
	}

	function handleUserScroll() {
		if (logsContainer) {
			userScrolledUp = !isScrolledToBottom(logsContainer);
		}
	}

	function clearLogs() {
		// Mark the current position - future messages after this will be shown
		query = ''; // Also clear search
		userScrolledUp = false; // Reset scroll tracking

		// Call parent callback if provided
		onClear?.();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape' && isMaximized) {
			isMaximized = false;
		} else if ((e.ctrlKey || e.metaKey) && e.key === 'f' && hasMessages) {
			e.preventDefault();
			searchInput?.focus();
		}
	}

	function handleModalClick(e: MouseEvent) {
		if (e.target === modalContainer) {
			isMaximized = false;
		}
	}

	function highlightText(text: string, search: string): string {
		if (!search) return text;
		const regex = new RegExp(`(${search.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi');
		return text.replace(regex, '<mark class="bg-yellow-300 dark:bg-yellow-600">$1</mark>');
	}

	export function scroll() {
		if (logsContainer) {
			scrollToBottom(logsContainer);
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<div>
	<div class="mb-2 flex items-center gap-2">
		<h2 class="text-lg font-semibold">{title}</h2>
		{#if showRefresh && onRefresh}
			<button
				onclick={onRefresh}
				use:tooltip={'Refresh logs'}
				class="rounded-md p-1 text-gray-500 hover:bg-gray-100 hover:text-gray-700 disabled:opacity-50 dark:text-gray-400 dark:hover:bg-gray-700 dark:hover:text-gray-300"
				disabled={refreshing}
				aria-label="Refresh logs"
			>
				<RefreshCw class="size-4 {refreshing ? 'animate-spin' : ''}" />
			</button>
		{/if}
		{#if error}
			<div
				use:tooltip={`An error occurred in connecting to the event stream. This is normal if the server is still starting up.`}
			>
				<AlertTriangle class="size-4 text-yellow-500" />
			</div>
		{/if}

		<div class="ml-auto flex items-center gap-1">
			<button
				onclick={clearLogs}
				use:tooltip={'Clear logs'}
				class="rounded-md p-1 text-gray-500 hover:bg-gray-100 hover:text-gray-700 disabled:opacity-50 dark:text-gray-400 dark:hover:bg-gray-700 dark:hover:text-gray-300"
				disabled={!hasMessages || refreshing}
				aria-label="Clear logs"
			>
				<X class="size-4" />
			</button>

			<button
				onclick={() => {
					isMaximized = true;
				}}
				use:tooltip={'Maximize (Esc to close)'}
				class="rounded-md p-1 text-gray-500 hover:bg-gray-100 hover:text-gray-700 disabled:opacity-50 dark:text-gray-400 dark:hover:bg-gray-700 dark:hover:text-gray-300"
				disabled={!hasMessages}
				aria-label="Maximize logs"
			>
				<Maximize class="size-4" />
			</button>
		</div>
	</div>

	<div
		bind:this={modalContainer}
		onclick={handleModalClick}
		class={twMerge(
			isMaximized
				? 'fixed inset-0 z-50 flex flex-col items-center justify-center bg-black/50 p-4 md:p-8 lg:p-10'
				: 'contents'
		)}
		role={isMaximized ? 'dialog' : undefined}
		aria-modal={isMaximized ? 'true' : undefined}
	>
		<div
			onscroll={handleUserScroll}
			bind:this={logsContainer}
			class={twMerge(
				'dark:bg-surface1 dark:border-surface3 default-scrollbar-thin bg-background flex min-h-64 flex-col overflow-y-auto rounded-lg border border-transparent shadow-sm',
				isMaximized ? 'h-full max-h-full w-full' : 'max-h-84 '
			)}
		>
			{#if hasMessages}
				<div
					class="dark:bg-surface1 bg-background sticky top-0 z-10 border-b border-gray-200 p-4 dark:border-gray-700"
				>
					<div
						class={twMerge(
							'border-surface2 bg-surface1/50 flex h-10 w-full items-center gap-2 rounded-sm border pr-2 pl-2 text-xs focus-within:outline focus-within:outline-2 focus-within:outline-blue-500',
							isMaximized && 'text-md h-12'
						)}
					>
						<div class="flex h-full max-h-8 items-center py-1.5">
							<Search class="h-full opacity-30" />
						</div>

						<input
							bind:this={searchInput}
							class="flex-1 bg-transparent py-3 outline-none placeholder:text-gray-400"
							type="text"
							placeholder="Search logs... (Ctrl+F)"
							bind:value={query}
							aria-label="Search logs"
						/>

						<div class="flex h-full items-center gap-1 p-0.5">
							{#if query}
								<span class="text-xs text-gray-500 dark:text-gray-400">
									{filteredMessages.length} / {messages.length}
								</span>
								<button
									class="hover:bg-surface2/80 active:bg-surface2/100 flex h-full max-h-8 items-center justify-center rounded-md p-1.5 opacity-30 hover:opacity-60"
									onclick={() => {
										query = '';
									}}
									aria-label="Clear search"
								>
									<X class="size-full text-current" />
								</button>
							{/if}

							{#if isMaximized}
								<button
									class="hover:bg-surface2/80 active:bg-surface2/100 flex h-full max-h-8 items-center justify-center rounded-md p-1.5 opacity-30 hover:opacity-60"
									onclick={() => {
										isMaximized = false;
									}}
									use:tooltip={'Close (Esc)'}
									aria-label="Close maximized view"
								>
									<Minimize class="size-full text-current" />
								</button>
							{/if}
						</div>
					</div>
				</div>
			{/if}

			{#if hasResults}
				<div class="messages-grid space-y-1 p-4">
					{#each filteredMessages as message, i (i)}
						<div
							class={twMerge(
								'group col-span-full grid grid-cols-subgrid gap-2 rounded px-2 py-1 font-mono text-sm hover:bg-gray-50 dark:hover:bg-gray-800',
								isMaximized && 'text-base'
							)}
							in:fade
						>
							<div class="border-surface3 border-r pr-1 text-right">
								<span class="text-gray-400 select-none dark:text-gray-600">{i + 1}</span>
							</div>
							<span class="text-on-surface1 flex-1">
								{@html highlightText(message, query)}
							</span>
						</div>
					{/each}
				</div>
			{:else if hasMessages && query}
				<div class="flex w-full flex-1 items-center justify-center p-6">
					<div class="text-center">
						<div class="text-on-surface1/80 font-medium">No matches found</div>
						<p class="text-on-surface1/60 mt-1 text-sm">Try a different search term</p>
					</div>
				</div>
			{:else}
				<div class="flex w-full flex-1 items-center justify-center p-6">
					<div class="text-center">
						<div class="text-on-surface1/80 font-medium">No deployment logs.</div>
						<p class="text-on-surface1/60 mt-1 text-sm">Try refreshing the logs.</p>
					</div>
				</div>
			{/if}
		</div>
	</div>
</div>

<style>
	.messages-grid {
		/* Show line numbers in a separate column */
		display: grid;
		grid-template-columns: auto 1fr;
	}
</style>
