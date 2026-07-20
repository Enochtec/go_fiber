<script lang="ts">
	import { onDestroy } from 'svelte';
	import Chart from 'chart.js/auto';

	export interface LineDataset {
		label: string;
		data: number[];
		color: string;
	}

	interface Props {
		labels: string[];
		datasets: LineDataset[];
		height?: number;
		currency?: boolean;
		smooth?: boolean;
	}

	let { labels, datasets, height = 220, currency = true, smooth = true }: Props = $props();

	let canvas = $state<HTMLCanvasElement | undefined>(undefined);
	let chart: Chart | null = null;

	function build() {
		if (!canvas) return;
		chart?.destroy();
		const ctx = canvas.getContext('2d')!;

		chart = new Chart(canvas, {
			type: 'line',
			data: {
				labels,
				datasets: datasets.map((d) => {
					const g = ctx.createLinearGradient(0, 0, 0, height);
					g.addColorStop(0, d.color + '38');
					g.addColorStop(1, d.color + '00');
					return {
						label: d.label,
						data: d.data,
						borderColor: d.color,
						backgroundColor: g,
						borderWidth: 2.5,
						fill: true,
						tension: smooth ? 0.42 : 0,
						pointRadius: 4,
						pointHoverRadius: 7,
						pointBackgroundColor: d.color,
						pointBorderColor: '#ffffff',
						pointBorderWidth: 2,
						pointHoverBorderWidth: 2.5,
					};
				}),
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				animation: { duration: 700, easing: 'easeInOutQuart' },
				interaction: { mode: 'index', intersect: false },
				plugins: {
					legend: {
						display: datasets.length > 1,
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
								const v = c.parsed.y;
								if (!currency) return ` ${c.dataset.label}: ${v}`;
								if (v >= 1_000_000) return ` ${c.dataset.label}: KES ${(v / 1_000_000).toFixed(2)}M`;
								if (v >= 1_000) return ` ${c.dataset.label}: KES ${(v / 1_000).toFixed(1)}K`;
								return ` ${c.dataset.label}: KES ${v.toLocaleString()}`;
							},
						},
					},
				},
				scales: {
					x: {
						grid: { color: '#e2e8f030' },
						border: { display: false },
						ticks: { color: '#94a3b8', font: { size: 10 }, maxRotation: 0, padding: 6 },
					},
					y: {
						grid: { color: '#e2e8f050' },
						border: { display: false },
						beginAtZero: true,
						ticks: {
							color: '#94a3b8',
							font: { size: 10 },
							padding: 8,
							callback: (v: string | number) => {
								const n = Number(v);
								if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
								if (n >= 1_000) return `${(n / 1_000).toFixed(0)}K`;
								return String(n);
							},
						},
					},
				},
			},
		});
	}

	$effect(() => {
		void labels;
		void datasets;
		build();
		return () => {
			chart?.destroy();
			chart = null;
		};
	});

	onDestroy(() => chart?.destroy());
</script>

<div style="position:relative; height:{height}px; width:100%;">
	<canvas bind:this={canvas}></canvas>
</div>
