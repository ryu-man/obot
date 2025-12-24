<script lang="ts">
	import { LoaderCircle, Group as GroupIcon, X } from 'lucide-svelte';

	import { Role } from '$lib/services/admin/types';
	import { getUserRoleLabel } from '$lib/utils';

	import GroupRoleForm from './GroupRoleForm.svelte';
	import type { GroupAssignment } from './types';
	import { responsive } from '$lib/stores';

	interface Props {
		groupAssignment?: GroupAssignment;
		loading?: boolean;
		onClose: () => void;
		onConfirm: (groupAssignment: GroupAssignment) => void;
		onAuditorConfirm: (groupAssignment: GroupAssignment) => void;
		onOwnerConfirm: (groupAssignment: GroupAssignment) => void;
	}

	// Helper functions to work with roles
	function getRoleId(role: number): number {
		return role & ~Role.AUDITOR;
	}

	function hasAuditorFlag(role: number): boolean {
		return (role & Role.AUDITOR) !== 0;
	}

	function addAuditorFlag(role: number): number {
		return role | Role.AUDITOR;
	}

	let {
		groupAssignment = $bindable(),
		loading = false,
		onClose,
		onConfirm,
		onAuditorConfirm,
		onOwnerConfirm
	}: Props = $props();

	let dialog = $state<HTMLDialogElement>();

	let draftRoleId = $state(0);
	let draftHaveAuditorPrevielage = $state(false);

	let isSmallScreen = $derived(responsive.isMobile);

	const hasRoleChanged = $derived.by(
		() => draftRoleId !== (groupAssignment ? groupAssignment.assignment.role : draftRoleId)
	);
	const hasAuditorChanged = $derived.by(
		() =>
			hasAuditorFlag(groupAssignment ? groupAssignment.assignment.role : 0) !==
			draftHaveAuditorPrevielage
	);

	// Check if any changes were made
	const hasChanges = $derived(hasRoleChanged || hasAuditorChanged);

	$effect(() => {
		if (groupAssignment) {
			// Initialize draft values from assignment
			const role = groupAssignment.assignment.role || 0;
			draftRoleId = getRoleId(role);
			draftHaveAuditorPrevielage = hasAuditorFlag(role);

			dialog?.showModal();
		} else {
			dialog?.close();
		}
	});

	function handleClose() {
		dialog?.close();
		onClose();
	}

	function handleConfirm() {
		if (!groupAssignment) return;

		const role = draftHaveAuditorPrevielage ? addAuditorFlag(draftRoleId) : draftRoleId;
		const result: GroupAssignment = {
			group: groupAssignment.group,
			assignment: {
				groupName: groupAssignment.group.name,
				role
			}
		};

		// Only description changed - update directly
		if (!hasRoleChanged && !hasAuditorChanged) {
			onConfirm(result);
			return;
		}

		// Auditor changed - show auditor confirmation
		if (hasAuditorChanged && draftHaveAuditorPrevielage && draftRoleId !== 0) {
			onAuditorConfirm(result);
			return;
		}

		// Changing to owner role - show owner confirmation
		const currentRoleId = getRoleId(groupAssignment.assignment.role || 0);
		if (draftRoleId === Role.OWNER && currentRoleId !== Role.OWNER) {
			onOwnerConfirm(result);
			return;
		}

		onConfirm(result);
	}
</script>

{#if groupAssignment}
	<dialog
		bind:this={dialog}
		class="flex max-h-[90svh] w-full max-w-[94svw] flex-col overflow-visible p-4 md:max-w-xl"
	>
		<div class="mb-6 flex flex-shrink-0 flex-col">
			<div class="flex items-center">
				{#if isSmallScreen}
					<div class="size-10"></div>
				{/if}
				<h3 class="default-dialog-title block flex-1 text-center md:text-start">
					{groupAssignment.assignment.role ? 'Update' : 'Assign'} Group Role
				</h3>

				<button onclick={handleClose} class="icon-button">
					<X class="size-5" />
				</button>
			</div>

			{#if groupAssignment.assignment.role}
				<div class="dark:bg-surface1 mt-3 flex flex-col gap-1 rounded-lg bg-gray-50 p-3">
					<div class="text-md flex items-center gap-1 text-black/50 dark:text-white/50">
						<GroupIcon class="size-5" />
						<span class="font-semibold">{groupAssignment.group.name}</span>
					</div>
					<div class="text-xs text-gray-600 dark:text-gray-400">
						{getUserRoleLabel(groupAssignment.assignment.role)}
					</div>
				</div>
			{/if}
		</div>

		<div class="flex-1 overflow-y-auto pr-2">
			<GroupRoleForm
				bind:roleId={draftRoleId}
				bind:hasAuditorPrivilege={draftHaveAuditorPrevielage}
				showRemoveOption={true}
			/>
		</div>

		<div class="mt-4 flex flex-shrink-0 justify-end gap-2">
			<button class="button" onclick={handleClose}>Cancel</button>
			<button
				class="button-primary"
				onclick={handleConfirm}
				disabled={loading || (!!groupAssignment.assignment.role && !hasChanges)}
			>
				{#if loading}
					<LoaderCircle class="size-4 animate-spin" />
				{:else}
					{groupAssignment.assignment.role ? 'Update' : 'Assign'}
				{/if}
			</button>
		</div>
	</dialog>
{/if}
