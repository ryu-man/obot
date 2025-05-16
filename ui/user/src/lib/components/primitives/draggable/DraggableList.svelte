<script lang="ts" generics="T">
	import { setDraggableContext, type DraggableContext, type DraggableItem } from './contextRoot';
	import { on } from 'svelte/events';
	import { twMerge } from 'tailwind-merge';

	type Props = {
		order: string[];
		onChange: (items: T[]) => void;
	};
	let {
		class: klass = '',
		as = 'div',
		order = [],
		disabled = false,
		onChange = undefined,
		children = undefined
	} = $props();

	let internalItems: DraggableItem<unknown>[] = $state([]);

	// Drag source Id
	let sourceItemId: string | undefined = $state();
	const sourceItem = $derived(internalItems.find((d) => d.id === sourceItemId));
	const sourceItemIndex = $derived(sourceItem ? internalItems.indexOf(sourceItem) : -1);

	// Drag target Id
	let targetItemId: string | undefined = $state();
	const targetItem = $derived(internalItems.find((d) => d.id === targetItemId));
	const targetItemIndex = $derived(targetItem ? internalItems.indexOf(targetItem) : -1);

	// timeout id before updating bindable data
	let synchTimeoutId: number | undefined = undefined;

	// sync number to trigger effect
	let sync: number | undefined = $state(undefined);

	// Share context with children
	const context: DraggableContext<unknown> = {
		get state() {
			return {
				items: internalItems,
				sourceItemId,
				targetItemId,
				disabled
			};
		},
		methods: {
			swap: () => {
				if (!sourceItem || !targetItem) {
					sourceItemId = undefined;
					targetItemId = undefined;
					return;
				}

				clearTimeout(synchTimeoutId);

				// take a snapshot
				let array = [...$state.snapshot(internalItems)] as DraggableItem<unknown>[];

				if (!sourceItem) return;

				if (!targetItem) return;

				array.splice(sourceItemIndex, 1, targetItem);
				array.splice(targetItemIndex, 1, sourceItem);

				internalItems = [...array];

				sourceItemId = undefined;
				targetItemId = undefined;

				synchTimeoutId = setTimeout(() => {
					// sync array
					sync = Date.now();
				}, 1000 / 60);
			},
			mount: (id, item) => {
				clearTimeout(synchTimeoutId);

				if (sync) {
					const preOrderArray = [...internalItems, item];

					const obj = preOrderArray.reduce(
						(acc, val) => {
							acc[val.id] = val;
							return acc;
						},
						{} as Record<string, DraggableItem<unknown>>
					);

					const orderedArray = [];

					for (const id of order) {
						const item = obj[id];
						if (item) {
							orderedArray.push(obj[id]);
						}
					}

					internalItems = [...orderedArray];
				} else {
					internalItems = [...internalItems, item];
				}

				synchTimeoutId = setTimeout(() => {
					// sync array
					sync = Date.now();
				}, 1000 / 60);

				return () => context.methods.unmount(id);
			},
			unmount: (id) => {
				clearTimeout(synchTimeoutId);

				internalItems = internalItems.filter((d) => d.id !== id);

				synchTimeoutId = setTimeout(() => {
					// sync arraysetSourceItem
					sync = Date.now();
				}, 1000 / 60);
			},
			setSourceItem: (id) => {
				sourceItemId = id;
			},
			setTargetItem: (id) => {
				targetItemId = id;
			}
		}
	};

	setDraggableContext(context);

	// only react if length changed
	const length = $derived(order.length);

	$effect(() => {
		if (sync === undefined) return;
		if (length === 0) return;
		if (length > internalItems.length) return;

		onChange?.(internalItems.map((d) => d.data));
	});

	// This code detects target id on touch-based devices, cause onpointerenter | onpointerleave does not works in touches
	$effect(() => {
		// check if current device supports touch events
		if (!navigator.maxTouchPoints) {
			// Not supported! Break
			return;
		}

		const onTouchMove = (ev: TouchEvent) => {
			requestAnimationFrame(() => {
				const touch = ev.touches[0];
				// Get the element under the touch
				const target = document.elementFromPoint(touch.clientX, touch.clientY);

				// find the root element of the draggable item
				const draggableItemElement = target?.closest('.draggable-element') as HTMLElement;

				// Not found! unset the target id
				if (!draggableItemElement) {
					targetItemId = undefined;
					return;
				}

				// get target id from the dataset
				targetItemId = draggableItemElement.dataset['id'];
			});
		};

		return on(window, 'touchmove', onTouchMove);
	});
</script>

<svelte:element this={as} class={twMerge('draggable-list flex flex-col', klass)}>
	{@render children?.()}
</svelte:element>
