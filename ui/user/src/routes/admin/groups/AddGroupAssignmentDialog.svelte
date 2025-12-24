<script lang="ts">
	import { debounce } from 'es-toolkit';
	import { LoaderCircle, Group as GroupIcon, X, Search as SearchIcon } from 'lucide-svelte';
	import { twMerge } from 'tailwind-merge';
	import { Role, type OrgGroup, type GroupRoleAssignment } from '$lib/services/admin/types';
	import { responsive } from '$lib/stores/index.js';
	import { getUserRoleLabel } from '$lib/utils';
	import GroupRoleForm from './GroupRoleForm.svelte';
	import type { GroupAssignment } from './types';

	interface Props {
		open?: boolean;
		groups: OrgGroup[];
		groupRoleMap: Record<string, GroupRoleAssignment>;
		loading?: boolean;
		onClose: () => void;
		onConfirm: (groupAssignment: GroupAssignment) => void;
		onAuditorConfirm: (groupAssignment: GroupAssignment) => void;
		onOwnerConfirm: (groupAssignment: GroupAssignment) => void;
	}

	function hasAuditorFlag(role: number): boolean {
		return (role & Role.AUDITOR) !== 0;
	}

	function addAuditorFlag(role: number): number {
		return role | Role.AUDITOR;
	}

	let {
		open = $bindable(),
		groups,
		groupRoleMap,
		loading = false,
		onClose,
		onConfirm,
		onAuditorConfirm,
		onOwnerConfirm
	}: Props = $props();

	let dialogElement = $state<HTMLDialogElement>();
	let searchQuery = $state('');
	let selectedGroup = $state<OrgGroup | undefined>();
	let draftRoleId = $state(0);
	let draftHaveAuditorPrevielage = $state(false);

	let isSmallScreen = $derived(responsive.isMobile);

	// Filter groups by search query
	const availableGroups = $derived(
		groups.filter((group) => group.name.toLowerCase().includes(searchQuery.toLowerCase()))
	);

	function resetForm() {
		searchQuery = '';
		selectedGroup = undefined;
		draftRoleId = 0;
		draftHaveAuditorPrevielage = false;
	}

	$effect(() => {
		if (open) {
			resetForm();
			dialogElement?.showModal();
		} else {
			dialogElement?.close();
		}
	});

	function handleClose() {
		dialogElement?.close();
		open = false;
		onClose();
	}

	function handleGroupSelect(group: OrgGroup) {
		selectedGroup = group;
		// Load existing assignment if available
		const existingAssignment = groupRoleMap[group.name];
		if (existingAssignment) {
			const role = existingAssignment.role || 0;
			draftRoleId = role & ~Role.AUDITOR;
			draftHaveAuditorPrevielage = hasAuditorFlag(role);
		} else {
			draftRoleId = 0;
			draftHaveAuditorPrevielage = false;
		}
	}

	function handleBack() {
		resetForm();
	}

	function handleConfirm() {
		if (!selectedGroup) return;

		const role = draftHaveAuditorPrevielage ? addAuditorFlag(draftRoleId) : draftRoleId;
		const result: GroupAssignment = {
			group: selectedGroup,
			assignment: {
				groupName: selectedGroup.name,
				role
			}
		};

		// Auditor changed - show auditor confirmation
		if (draftHaveAuditorPrevielage && draftRoleId !== 0) {
			onAuditorConfirm(result);
			return;
		}

		// Changing to owner role - show owner confirmation
		if (draftRoleId === Role.OWNER) {
			onOwnerConfirm(result);
			return;
		}

		onConfirm(result);
	}

	const updateSearch = debounce((value: string) => {
		searchQuery = value;
	}, 100);
</script>

{#snippet groupList()}
	<div class="flex flex-col gap-4 overflow-y-auto pr-2">
		<div class="relative sticky top-0 z-10 flex-shrink-0 bg-white">
			<SearchIcon class="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-gray-400" />
			<input
				type="text"
				value={searchQuery}
				oninput={(e) => updateSearch(e.currentTarget.value)}
				placeholder="Search groups..."
				class="dark:bg-surface2 dark:border-surface3 w-full rounded-lg border py-2 pr-3 pl-10 text-sm"
			/>
		</div>

		<div class="flex flex-col gap-2">
			{#if availableGroups.length === 0}
				<p class="py-8 text-center text-sm text-gray-500">
					{searchQuery ? 'No groups found matching your search.' : 'No groups available.'}
				</p>
			{:else}
				{#each availableGroups as group (group.id)}
					{@const hasAssignment = !!groupRoleMap[group.name]}
					{@const assignedRole = groupRoleMap[group.name]?.role}
					<button
						onclick={() => handleGroupSelect(group)}
						class={twMerge(
							'border-surface3 flex items-center gap-3 rounded-lg border p-3 text-left transition-colors hover:bg-black/5 dark:hover:bg-white/5',
							selectedGroup?.id === group.id && 'bg-primary/10 border-primary'
						)}
					>
						<div class="flex flex-1 items-center gap-3">
							{#if group.iconURL}
								<img src={group.iconURL} alt={group.name} class="size-8 rounded-full" />
							{:else}
								<div
									class="dark:bg-surface3 flex size-8 items-center justify-center rounded-full bg-gray-200"
								>
									<GroupIcon class="size-4" />
								</div>
							{/if}
							<div class="flex flex-1 flex-col">
								<span class="font-medium">{group.name}</span>
								{#if hasAssignment && assignedRole}
									<span class="text-xs text-gray-500">{getUserRoleLabel(assignedRole)}</span>
								{/if}
							</div>
						</div>
					</button>
				{/each}
			{/if}
		</div>
	</div>
{/snippet}

{#snippet roleForm()}
	<div class="flex flex-col gap-4 overflow-y-auto pr-2">
		{#if selectedGroup}
			<div class="dark:bg-surface1 flex flex-col gap-1 rounded-lg bg-gray-50 p-3">
				<div class="text-md flex items-center gap-2">
					{#if selectedGroup.iconURL}
						<img src={selectedGroup.iconURL} alt={selectedGroup.name} class="size-6 rounded-full" />
					{:else}
						<GroupIcon class="size-5" />
					{/if}
					<span class="font-semibold">{selectedGroup.name}</span>
				</div>
				<div class="text-xs text-gray-600 dark:text-gray-400">
					{#if groupRoleMap[selectedGroup.name]}
						Update the role for this group
					{:else}
						Select a role to assign to this group
					{/if}
				</div>
			</div>

			<GroupRoleForm
				bind:roleId={draftRoleId}
				bind:hasAuditorPrivilege={draftHaveAuditorPrevielage}
				showRemoveOption={selectedGroup && !!groupRoleMap[selectedGroup.name]}
			/>
		{:else}
			<div class="flex h-full items-center justify-center py-12 text-sm text-gray-500">
				Select a group to assign a role
			</div>
		{/if}
	</div>
{/snippet}

{#if open}
	<dialog
		bind:this={dialogElement}
		class={twMerge(
			'flex max-h-[90svh] max-w-[94svw] flex-col overflow-visible p-4 md:min-h-[768px]',
			!isSmallScreen ? 'w-full max-w-4xl' : 'w-full'
		)}
	>
		<div class="mb-6 flex items-center gap-0">
			{#if isSmallScreen && selectedGroup}
				<button onclick={handleBack} class="icon-button">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						width="20"
						height="20"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
					>
						<path d="m15 18-6-6 6-6" />
					</svg>
				</button>
			{:else if isSmallScreen}
				<div class="size-10"></div>
			{/if}

			<h3 class="default-dialog-title block flex-1 text-center md:text-start">
				{#if selectedGroup && groupRoleMap[selectedGroup.name]}
					Update Group Role
				{:else}
					Assign Group Role
				{/if}
			</h3>

			<button onclick={handleClose} class="icon-button">
				<X class="size-5" />
			</button>
		</div>

		{#if !isSmallScreen}
			<!-- Large screen: two-column layout -->
			<div class="grid flex-1 grid-cols-2 gap-8 overflow-hidden">
				<div class="flex flex-col overflow-hidden">
					<h4 class="mb-4 flex-shrink-0 text-sm font-semibold">Select Group</h4>
					{@render groupList()}
				</div>
				<div class="flex flex-col overflow-hidden">
					<h4 class="mb-4 flex-shrink-0 text-sm font-semibold">Assign Role</h4>
					{@render roleForm()}
				</div>
			</div>
		{:else}
			<!-- Small screen: single column with conditional rendering -->
			{#if !selectedGroup}
				<div class="flex flex-1 flex-col overflow-hidden">
					<h4 class="mb-4 flex-shrink-0 text-sm font-semibold">Select Group</h4>
					{@render groupList()}
				</div>
			{:else}
				<div class="flex flex-1 flex-col overflow-hidden">
					{@render roleForm()}
				</div>
			{/if}
		{/if}

		<div class="mt-6 flex flex-shrink-0 justify-end gap-2">
			<button class="button" onclick={handleClose}>Cancel</button>
			<button
				class="button-primary"
				onclick={handleConfirm}
				disabled={loading || !selectedGroup || draftRoleId === 0}
			>
				{#if loading}
					<LoaderCircle class="size-4 animate-spin" />
				{:else if selectedGroup && groupRoleMap[selectedGroup.name]}
					Update Role
				{:else}
					Assign Role
				{/if}
			</button>
		</div>
	</dialog>
{/if}
