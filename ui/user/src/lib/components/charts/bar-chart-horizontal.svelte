<script lang="ts" generics="T">
	import { scaleBand, scaleLinear, select, axisBottom, axisLeft, scaleOrdinal, max } from 'd3';
	import { autoUpdate, computePosition, flip, offset } from '@floating-ui/dom';
	import { fade } from 'svelte/transition';
	import type { Snippet } from 'svelte';

	interface Props<T> {
		data: T[];
		xGet: (d: T) => number;
		yGet: (d: T) => string;
		padding?: number;
		formatTooltipValue?: (data: T) => string;
		formatTooltipTitle?: (d: T) => string;
		formatXLabel?: (d: T) => string;
		formatYLabel?: (d: T) => string;
		tooltip?: Snippet<
			[
				{
					data: T;
					x: number;
					y: string;
				}
			]
		>;
	}

	let {
		data,
		xGet,
		yGet,
		formatTooltipValue,
		formatTooltipTitle,
		formatXLabel,
		formatYLabel,
		tooltip
	}: Props<T> = $props();

	let highlightedRectElement = $state<SVGRectElement>();

	let paddingLeft = $state(1);
	let paddingRight = $state(1);
	let paddingTop = $state(0);
	let paddingBottom = $state(16);

	let clientWidth = $state(0);
	let innerWidth = $derived(clientWidth - paddingLeft - paddingRight);

	let clientHeight = $state(0);
	let innerHeight = $derived(clientHeight - paddingTop - paddingBottom);

	const xDomain = $derived.by(() => {
		const maxValue = max(data, xGet);

		return [0, maxValue ?? 0];
	});
	const xRange = $derived([0, innerWidth]);
	const xScale = $derived(scaleLinear(xDomain, xRange).nice());

	const yDomain = $derived(data.map(yGet));
	const yRange = $derived.by(() => {
		const minHeight = 32;

		if (yDomain.length * minHeight > innerHeight) {
			return [0, innerHeight];
		}

		return [0, yDomain.length * minHeight];
	});
	const yScale = $derived(scaleBand(yRange).domain(yDomain).paddingInner(0.1).paddingOuter(0.1));

	const colors: string[] = [
		'#254993',
		'#D65C7C',
		'#635DB6',
		'#D6A95C',
		'#2EB88A',
		'#47A3D1',
		'#D0CE43',
		'#D97850',
		'#5C9B72',
		'#D44968'
	];

	const colorScale = $derived(
		scaleOrdinal(
			yDomain,
			yDomain.map((_, i) => colors[i] ?? '#999999')
		)
	);

	let currentItem = $state<{ x: number; y: string; data: T }>();

	function createTooltip(reference: Element, floating: HTMLElement) {
		const compute = async () => {
			const position = await computePosition(reference, floating, {
				placement: 'top',
				middleware: [
					offset(8),
					flip({
						padding: {
							top: 0,
							right: 40,
							left: 40,
							bottom: 0
						},
						boundary: document.documentElement,
						fallbackPlacements: ['top', 'top-end', 'top-start', 'left-start', 'right-start']
					})
				]
			});

			const { x, y } = position;

			floating.style.transform = `translate(${x}px, ${y}px)`;
		};

		return autoUpdate(reference, floating, compute, {
			animationFrame: true,
			ancestorScroll: true,
			ancestorResize: true
		});
	}
</script>

<div bind:clientHeight bind:clientWidth class="group relative h-full w-full">
	{#if highlightedRectElement && currentItem}
		<div
			class="pointer-events-none fixed top-0 left-0 flex flex-col"
			{@attach (node) => createTooltip(highlightedRectElement!, node)}
			in:fade={{ duration: 100, delay: 10 }}
			out:fade={{ duration: 100 }}
		>
			{#if tooltip}
				{@render tooltip({ data: currentItem.data, x: currentItem.x, y: currentItem.y })}
			{:else}
				<div
					class="bg-background border-surface3 dark:border-surface3 flex flex-col gap-1 rounded-lg border px-4 py-3 shadow-lg backdrop-blur-sm"
				>
					<div class="text-on-surface3 dark:text-on-surface1 text-xs font-medium opacity-90">
						{formatTooltipTitle ? formatTooltipTitle(currentItem.data) : currentItem.y}
					</div>
					<div class="text-on-surface3 dark:text-on-surface1 text-2xl font-semibold leading-tight">
						{formatTooltipValue ? formatTooltipValue(currentItem.data) : currentItem?.x}
					</div>
				</div>
			{/if}
		</div>
	{/if}

	<svg width={clientWidth} height={clientHeight} viewBox={`0 0 ${clientWidth} ${clientHeight}`}>
		<g transform="translate({paddingLeft}, {paddingTop})">
			<g
				class="y-axis text-on-surface3/20 dark:text-on-surface1/10"
				{@attach (node: SVGGElement) => {
					const selection = select(node);

					const axis = axisLeft(yScale).tickSizeOuter(0);

					selection
						.transition()
						.duration(100)
						.call(axis as any)
						.selectAll('.tick')
						.selectAll('line, text')
						.attr('class', function (_d) {
							return '';
						});
				}}
			></g>

			<g
				class="x-axis text-on-surface3 dark:text-on-surface1"
				transform="translate(0, {innerHeight})"
				{@attach (node: SVGGElement) => {
					const selection = select(node);

					const axis = axisBottom(xScale);

					selection.selectAll('.tick>line').attr('y1', -innerHeight).attr('opacity', 0.1);
					selection.selectAll('.tick>text').attr('text-anchor', (_d, i, array) => {
						if (i === 0) return 'start';
						if (i === array.length - 1) return 'end';
						return 'middle';
					});

					selection.select('.domain').attr('opacity', 0);
					selection.transition().duration(100).call(axis);
				}}
			></g>

			<g
				class="data"
				{@attach (node: SVGGElement) => {
					const selection = select(node);

					// Create groups with data bound once
					const groups = selection
						.selectAll('.bar-group')
						.data(data)
						.join('g')
						.attr('class', 'bar-group')
						.attr('color', (d) => colorScale(yGet(d)))
						.attr('transform', (d) => `translate(0, ${yScale(yGet(d)) ?? 0})`);

					// Append one rect per group (data is already bound to the group)
					groups
						.selectAll('rect')
						.data((d) => [d]) // Wrap in array to ensure one rect per group
						.join('rect')
						.attr('class', 'bar')
						.attr('cursor', 'pointer')
						.attr('width', (d) => Math.max(0, xScale(xGet(d)) ?? 0))
						.attr('height', yScale.bandwidth())
						.attr('fill', 'currentColor')
						.on('pointerenter', function (_ev, d) {
							highlightedRectElement = this as SVGRectElement;

							currentItem = {
								y: yGet(d) + '',
								x: xGet(d),
								data: d
							};

							select(this)
								.attr('stroke', 'currentColor')
								.attr('stroke-width', 2)
								.attr('fill-opacity', 0.8);
						})
						.on('pointerleave', function () {
							if (this === highlightedRectElement) {
								highlightedRectElement = undefined;
							}

							select(this).attr('stroke-width', 0).attr('fill-opacity', 1);
						});

					// Build labels - one text per group
					const minWidthForBothInside = 120; // minimum bar width to place both labels inside

					const textGroups = groups
						.selectAll('text')
						.data((d) => [d]) // Wrap in array to ensure one text per group
						.join('text');

					textGroups
						.attr('dominant-baseline', 'middle')
						.attr('pointer-events', 'none')
						.attr('x', (d) => {
							const barWidth = Math.max(0, xScale(xGet(d)) ?? 0);
							if (barWidth > minWidthForBothInside) {
								// Place inside, near the left edge
								return 8;
							} else {
								// Place outside, to the right
								return barWidth + 6;
							}
						})
						.attr('y', () => {
							// Since we're using transform on the group, y is relative to the group
							const barHeight = yScale.bandwidth();
							return barHeight / 2;
						})
						.attr('text-anchor', 'start')
						.each(function (d) {
							const barWidth = Math.max(0, xScale(xGet(d)) ?? 0);
							const isInside = barWidth > minWidthForBothInside;
							const textElement = select(this);

							// Clear existing content
							textElement.selectAll('*').remove();

							// Add band name tspan
							const nameTspan = textElement
								.append('tspan')
								.attr('class', 'band-name')
								.attr('font-size', isInside ? '13px' : '12px')
								.attr('font-weight', isInside ? '600' : '500')
								.attr('fill', isInside ? 'rgba(255, 255, 255, 0.9)' : 'currentColor')
								.attr('opacity', isInside ? 1 : 0.85)
								.text(formatYLabel ? formatYLabel(d) : yGet(d) + '');

							// Add value tspan
							const valueTspan = textElement
								.append('tspan')
								.attr('class', 'band-value')
								.attr('font-size', '11px')
								.attr('font-weight', '400')
								.attr('fill', isInside ? 'rgba(255, 255, 255, 0.7)' : 'currentColor')
								.attr('opacity', isInside ? 1 : 0.7)
								.text(formatXLabel ? formatXLabel(d) : xGet(d) + '');

							// Position value at the far right when inside
							if (isInside) {
								const nameWidth = (nameTspan.node() as SVGTextElement).getComputedTextLength();
								const valueWidth = (valueTspan.node() as SVGTextElement).getComputedTextLength();
								const availableSpace = barWidth - 16; // 8px left padding + 8px right padding
								const spaceBetween = availableSpace - nameWidth - valueWidth;

								if (spaceBetween > 5) {
									// Add space to push value to the right
									valueTspan.attr('dx', spaceBetween);
								} else {
									// Not enough space, just add minimal space and truncate name if needed
									valueTspan.attr('dx', 5);

									const totalWidth = nameWidth + 5 + valueWidth;
									if (totalWidth > availableSpace) {
										let truncated = yGet(d);
										while (
											(nameTspan.node() as SVGTextElement).getComputedTextLength() +
												5 +
												valueWidth >
												availableSpace &&
											truncated.length > 3
										) {
											truncated = truncated.slice(0, -1);
											nameTspan.text(truncated + '...');
										}
									}
								}
							} else {
								// Outside, add more space for better breathing room
								valueTspan.attr('dx', 8);
							}
						});
				}}
			>
			</g>
		</g>
	</svg>
</div>
