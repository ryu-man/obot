<script module>
	const MS_SECOND = 1000;
	const MS_MINUTE = MS_SECOND * 60;
	const MS_HOUR = MS_MINUTE * 60;
	const MS_DAY = MS_HOUR * 24;
	const MS_MONTH = MS_DAY * 30.436875;
</script>

<script lang="ts">
	import {
		scaleBand,
		scaleLinear,
		scaleOrdinal,
		scaleTime,
		stack,
		union,
		extent,
		select,
		axisBottom,
		rollup,
		timeHour,
		timeMonth,
		timeDay,
		timeMinute,
		max,
		min,
		timeDays,
		timeHours,
		timeWeeks,
		timeMinutes,
		timeMonths,
		axisLeft
	} from 'd3';
	import {
		set,
		startOfMonth,
		endOfMonth,
		isWithinInterval,
		startOfHour,
		endOfHour,
		startOfDay,
		endOfDay,
		startOfYear,
		endOfYear,
		type DateValues,
		intervalToDuration
	} from 'date-fns';
	import type { AuditLog } from '$lib/services';

	interface Props<T> {
		start: Date;
		end: Date;
		data: T[];
		padding?: number;
	}

	let { start, end, data }: Props<AuditLog> = $props();

	let tooltipElement = $state<HTMLElement>();

	let paddingLeft = $state(24);
	let paddingRight = $state(8);
	let paddingTop = $state(8);
	let paddingBottom = $state(16);

	let clientWidth = $state(0);
	let innerWidth = $derived(clientWidth - paddingLeft - paddingRight);

	let clientHeight = $state(0);
	let innerHeight = $derived(clientHeight - paddingTop - paddingBottom);

	const callTypes = $derived(union(data.map((d) => d.callType)));

	const duration = $derived(Math.abs(+end - +start));

	const frame = $derived.by(() => {
		if (duration >= MS_MONTH) {
			return 'monthly';
		}

		if (duration >= MS_DAY) {
			return 'daily';
		}

		if (duration > MS_HOUR) {
			return 'hourly';
		}

		return 'base';
	});

	const boundaries = $derived.by(() => {
		if (frame === 'hourly') return [startOfDay, endOfDay];

		if (frame === 'daily') return [startOfMonth, endOfMonth];

		if (frame === 'monthly') return [startOfYear, endOfYear];

		return [startOfHour, endOfHour];
	});

	const timeFrameDomain = $derived.by(() => {
		const [setStartBoundary, setEndBoundary] = boundaries;

		if (!start && !end) {
			const [mn, mx] = extent(data, (d) => new Date(d.createdAt));

			return [mn ? setStartBoundary(mn) : mn, mx ? setEndBoundary(mx) : mx];
		}

		if (start && end) {
			return [setStartBoundary(start), setEndBoundary(end)];
		}

		if (start) {
			const mx = max(data, (d) => new Date(d.createdAt));
			return [setStartBoundary(start), mx ? setEndBoundary(mx) : max];
		}

		const mn = min(data, (d) => new Date(d.createdAt));
		return [setStartBoundary(mn ?? new Date()), end ? setEndBoundary(end) : end];
	});

	const xAccessor = $derived.by(() => {
		let options: DateValues = { milliseconds: 0, seconds: 0 };

		if (frame === 'hourly') {
			options = { ...options, minutes: 0 };
		}

		if (frame === 'daily') {
			options = { ...options, minutes: 0, hours: 0 };
		}

		if (frame === 'monthly') {
			options = { ...options, minutes: 0, hours: 0, date: 1 };
		}

		return (d: AuditLog) => set(new Date(d.createdAt), options).toISOString();
	});

	const bands = $derived.by(() => {
		const [start, end] = timeFrameDomain as [Date, Date];

		type Generator =
			| typeof timeMinutes
			| typeof timeHours
			| typeof timeDays
			| typeof timeWeeks
			| typeof timeMonths;

		let generator: Generator = timeMinutes;
		let step = 1;

		if (frame === 'hourly') {
			generator = timeHours;
		}

		if (frame === 'daily') {
			generator = timeDays;
		}

		if (frame === 'monthly') {
			generator = timeMonths;
		}

		return union(generator(start, end, step).map((d) => d.toISOString()));
	});

	const xRange = $derived([0, innerWidth]);

	const timeScale = $derived(scaleTime(timeFrameDomain, xRange));

	const xScale = $derived(scaleBand(xRange).domain(bands).paddingInner(0.1).paddingOuter(0.1));

	const xAxisTicks = $derived.by(() => {
		let generator = timeMonth;
		let step = 1;

		if (frame === 'base') {
			generator = timeMinute;
			step = 5;
		}

		if (frame === 'hourly') {
			generator = timeHour;
			const duration = intervalToDuration({ start, end });
			const hours = duration.hours ?? 0;

			if (hours >= 8) {
				step = Math.ceil(12 / hours);
			}
		}

		if (frame === 'daily') {
			generator = timeDay;
			step = 2;

			if (duration > 2 * MS_MONTH) {
				step = 4;
			}
		}

		if (frame === 'monthly') {
			generator = timeMonth;
		}

		return generator.every(step);
	});

	const colorByCallType = {
		initialize: '#637BE9',
		'notifications/initialized': '#EB4773',
		'prompts/list': '#F5B13D',
		'resources/list': '#12D393',
		'tools/call': '#1EB1FA',
		'tools/list': '#EDEA26'
	};

	const callTypesArray = $derived(callTypes.values().toArray());
	const colorScale = $derived(
		scaleOrdinal(
			callTypesArray,
			callTypesArray.map((d) => colorByCallType[d] ?? '#999999')
		)
	);

	const group = $derived.by(() => {
		return rollup(
			$state.snapshot(data),
			(d) => d.length,
			xAccessor,
			(d) => d.callType
		);
	});

	const series = $derived.by(() => {
		const stacked = stack()
			.keys(callTypes)
			.value((d, key) => d[1].get(key) ?? 0);

		return stacked(group);
	});

	const yDomain = $derived(extent(series.map((serie) => extent(serie.flat())).flat(), (d) => d));
	const yScale = $derived(scaleLinear(yDomain, [innerHeight, 0]));

	let currentItem = $state<{ key: string; value: string; date: string }>();
</script>

<div bind:clientHeight bind:clientWidth class="group h-full w-full">
	<div
		class="tooltip pointer-events-none fixed top-0 left-0 flex flex-col"
		style="opacity: 0;"
		{@attach (node) => {
			tooltipElement = node;
		}}
	>
		<div class="flex flex-col gap-0 text-xs">
			<div>
				{currentItem?.date}
			</div>
			<div class="text-sm">
				{currentItem?.key}
			</div>
		</div>
		<div class="text-2xl font-bold">{currentItem?.value}</div>
	</div>

	<svg width={clientWidth} height={clientHeight} viewBox={`0 0 ${clientWidth} ${clientHeight}`}>
		<g transform="translate({paddingLeft}, {paddingTop})">
			<g
				class="x-axis text-on-surface3/10 dark:text-on-surface1/10"
				transform="translate(0 {innerHeight})"
				{@attach (node) => {
					const selection = select(node);
					const axis = axisBottom(timeScale).tickSizeOuter(0).ticks(xAxisTicks);

					selection
						.transition()
						.duration(200)
						.call(axis)
						.selectAll('.tick')
						.attr('transform', (d) => `translate(${timeScale(d) + xScale.bandwidth() / 2}, 0)`)
						.selectAll('line, text')
						.attr('class', function (d) {
							const element = this as SVGElement;

							const isActive = isWithinInterval(d, {
								start,
								end
							});

							const classNames = new Set(element.classList);
							const activeClassName = [
								'text-on-surface3',
								'dark:text-on-surface1',
								'duration-1000',
								'transiton-colors'
							];

							const callbackfn = (d: string) =>
								isActive ? classNames.add(d) : classNames.delete(d);
							activeClassName.forEach(callbackfn);

							// Keep old class names
							// Filter falsy values and join with a space
							return classNames.values().toArray().join(' ');
						});
				}}
			></g>

			<g
				class="y-axis text-on-surface3/10 dark:text-on-surface1/10"
				{@attach (node) => {
					select(node)
						.transition()
						.duration(100)
						.call(axisLeft(yScale).tickSizeOuter(0).ticks(3))
						.selectAll('.tick>line')
						.attr('x1', innerWidth);

					select(node).select('.domain').attr('opacity', 0);
				}}
			></g>

			<g
				class="data"
				{@attach (node) => {
					select(node)
						.selectAll('g')
						.data(series)
						.join('g')
						.attr('class', 'serie')
						.attr('data-type', (d) => d.key)
						.attr('fill', (d) => colorScale(d.key))
						.selectAll('rect')
						.data((d) => d)
						.join('rect')
						.attr('x', (d) => xScale(d.data[0]))
						.attr('y', (d) => yScale(d[1]))
						.attr('height', (d) => yScale(d[0]) - yScale(d[1]))
						.attr('width', xScale.bandwidth())
						.attr('cursor', 'pointer')
						.on('mouseover', function () {
							// Show tooltip
							if (tooltipElement) {
								select(tooltipElement).style('opacity', 1);
							}

							// Highlight the hovered bar
							select(this).style('stroke', 'white').style('stroke-width', 2);
						})
						.on('mousemove', function (event, d) {
							if (!tooltipElement) return;

							const t = select(tooltipElement);

							const item = {};

							const parentData = select(this.parentNode).datum();

							// Update tooltip content and position
							item.key = parentData.key;

							// The actual value of this segment
							item.value = d[1] - d[0];

							item.date = new Date(d.data[0]).toLocaleString();

							currentItem = { ...item };

							t.style('left', event.pageX + 10 + 'px').style('top', event.pageY - 28 + 'px');
						})
						.on('mouseout', function () {
							if (!tooltipElement) return;

							const t = select(tooltipElement);
							// Hide tooltip
							t.style('opacity', 0);
							// Remove highlight
							select(this).style('stroke', 'none');
						});
				}}
			>
			</g>
		</g>
	</svg>
</div>
