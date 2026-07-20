<script lang="ts">
	import { onDestroy } from 'svelte';
	import Chart from 'chart.js/auto';

	interface Props {
		labels: string[];
		data: number[];
		subData?: number[];
		subLabel?: string;
		color?: string;
		subColor?: string;
		label?: string;
		height?: number;
		currency?: boolean;
	}

	let {
		labels, data,
		subData = [],
		subLabel = 'Units',
		color = '#6366f1',
		subColor = '#10b981',
		label = 'Revenue',
		height = 220,
		currency = true,
	}: Props = $props();

	let canvas = $state<HTMLCanvasElement | undefined>(undefined);
	let chart: Chart | null = null;

	function build() {
		if (!canvas) return;
		chart?.destroy();
		const ctx = canvas.getContext('2d')!;

		const grad = ctx.createLinearGradient(0, 0, canvas.offsetWidth || 400, 0);
		grad.addColorStop(0, color + 'b0');
		grad.addColorStop(1, color);

		const datasets: object[] = [{
			label,
			data,
			backgroundColor: grad,
			hoverBackgroundColor: color,
			borderRadius: 4,
			borderSkipped: false,
			xAxisID: 'x',
		}];

		if (subData.length > 0) {
			datasets.push({
				label: subLabel,
				data: subData,
				backgroundColor: subColor + '80',
				hoverBackgroundColor: subColor,
				borderRadius: 4,
				borderSkipped: false,
				xAxisID: 'x2',
			});
		}

		chart = new Chart(canvas, {
			type: 'bar',
			data: { labels, datasets },
			options: {
				indexAxis: 'y',
				responsive: true,
				maintainAspectRatio: false,
				animation: { duration: 600, easing: 'easeInOutQuart' },
				plugins: {
					legend: {
						display: subData.length > 0,
						position: 'top',
						labels: { color: '#94a3b8', font: { size: 11 }, usePointStyle: true, padding: 16 },
					},
					tooltip: {
						backgroundColor: '#0f172a',
						titleColor: '#64748b',
						bodyColor: '#f1f5f9',
						borderColor: '#1e293b',
						borderWidth: 1,
						padding: 12,
						cornerRadius: 8,
						callbacks: {
							label: (c) => {
								const v = c.parsed.x;
								if (c.datasetIndex === 1) return ` ${subLabel}: ${v} units`;
								if (!currency) return ` ${v}`;
								if (v >= 1_000_000) return ` KES ${(v / 1_000_000).toFixed(2)}M`;
								if (v >= 1_000) return ` KES ${(v / 1_000).toFixed(1)}K`;
								return ` KES ${v.toLocaleString()}`;
							},
						},
					},
				},
				scales: {
					y: {
						grid: { display: false },
						border: { display: false },
						ticks: { color: '#475569', font: { size: 11, weight: '500' } },
					},
					x: {
						grid: { color: '#e2e8f050' },
						border: { display: false },
						position: 'bottom',
						ticks: {
							color: '#94a3b8',
							font: { size: 10 },
							callback: (v: string | number) => {
								const n = Number(v);
								if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
								if (n >= 1_000) return `${(n / 1_000).toFixed(0)}K`;
								return String(n);
							},
						},
					},
					...(subData.length > 0 ? {
						x2: {
							position: 'top',
							grid: { display: false },
							border: { display: false },
							ticks: { color: '#94a3b8', font: { size: 9 } },
						},
					} : {}),
				},
			},
		});
	}

	$effect(() => {
		void labels; void data; void subData;
		build();
		return () => { chart?.destroy(); chart = null; };
	});

	onDestroy(() => chart?.destroy());
</script>

<div style="position:relative; height:{height}px; width:100%;">
	<canvas bind:this={canvas}></canvas>
</div>
