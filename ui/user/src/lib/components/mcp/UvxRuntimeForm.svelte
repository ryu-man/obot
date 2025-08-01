<script lang="ts">
	import type { UVXRuntimeConfig } from '$lib/services/chat/types';
	import { Plus, Trash2 } from 'lucide-svelte';
	import { tooltip } from '$lib/actions/tooltip.svelte';

	interface Props {
		config: UVXRuntimeConfig;
		readonly?: boolean;
	}
	let { config = $bindable(), readonly }: Props = $props();

	// Initialize args array if it doesn't exist
	if (!config.args) {
		config.args = [];
	}

	function addArgument() {
		if (!config.args) {
			config.args = [];
		}
		config.args.push('');
	}

	function removeArgument(index: number) {
		if (config.args) {
			config.args.splice(index, 1);
		}
	}

	function handlePaste(event: ClipboardEvent, index: number) {
		if (readonly || !config.args) return;

		event.preventDefault();
		const pastedText = event.clipboardData?.getData('text');
		if (!pastedText) return;

		const lines = pastedText.split(/[\r\n]+/).filter((line) => line.trim());
		if (lines.length <= 1) {
			config.args[index] = pastedText;
			return;
		}

		// Remove quotes, commas and trim each line
		const cleanedLines = lines.map((line) => {
			let trimmed = line.trim();
			if (trimmed.endsWith(',')) {
				trimmed = trimmed.slice(0, -1).trim();
			}

			if (
				(trimmed.startsWith('"') && trimmed.endsWith('"')) ||
				(trimmed.startsWith("'") && trimmed.endsWith("'"))
			) {
				trimmed = trimmed.slice(1, -1).trim();
			}
			return trimmed;
		});

		config.args[index] = cleanedLines[0];
		for (let j = 1; j < cleanedLines.length; j++) {
			config.args.splice(index + j, 0, cleanedLines[j]);
		}
	}
</script>

<div
	class="dark:bg-surface1 dark:border-surface3 flex flex-col gap-4 rounded-lg border border-transparent bg-white p-4 shadow-sm"
>
	<h4 class="text-sm font-semibold">UVX Runtime Configuration</h4>

	<!-- Package field (required) -->
	<div class="flex items-center gap-4">
		<label for="uvx-package" class="text-sm font-light">Package</label>
		<input
			id="uvx-package"
			class="text-input-filled w-full dark:bg-black"
			bind:value={config.package}
			disabled={readonly}
			placeholder="e.g. @modelcontextprotocol/server-filesystem"
			onblur={() => {
				if (config.package) {
					config.package = config.package.trim();
				}
			}}
			required
		/>
	</div>

	<!-- Command field (optional) -->
	<div class="flex items-center gap-4">
		<label for="uvx-command" class="text-sm font-light">Command</label>
		<input
			id="uvx-command"
			class="text-input-filled w-full dark:bg-black"
			bind:value={config.command}
			disabled={readonly}
			placeholder="e.g. mcp-server-filesystem"
			onblur={() => {
				if (config.command) {
					config.command = config.command.trim();
				}
			}}
		/>
	</div>

	<!-- Arguments field (optional) -->
	{#if config.args}
		<div class="flex gap-4">
			<span class="pt-2.5 text-sm font-light">Arguments</span>
			<div class="flex min-h-10 grow flex-col gap-4">
				{#each config.args as _arg, i (i)}
					<div class="flex items-center gap-2">
						<input
							class="text-input-filled w-full dark:bg-black"
							bind:value={config.args[i]}
							disabled={readonly}
							placeholder="e.g. /path/to/directory"
							onblur={() => {
								if (config.args && config.args[i]) {
									config.args[i] = config.args[i].trim();
								}
							}}
							onpaste={(e) => handlePaste(e, i)}
						/>
						{#if !readonly}
							<button
								class="icon-button"
								onclick={() => removeArgument(i)}
								use:tooltip={'Remove argument'}
							>
								<Trash2 class="size-4" />
							</button>
						{/if}
					</div>
				{/each}

				{#if !readonly}
					<div class="flex justify-end">
						<button class="button flex items-center gap-1 text-xs" onclick={addArgument}>
							<Plus class="size-4" /> Argument
						</button>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
