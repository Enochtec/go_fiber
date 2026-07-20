<script lang="ts">
	import { onDestroy } from 'svelte';
	import Chart from 'chart.js/auto';

	interface Props {
		labels: string[];
		data: number[];
		color?: string;
		highlightColor?: string;
		label?: string;
		height?: number;
		currency?: boolean;
	}

	let {
		labels, data,
		color = '#6366f1',
		highlightColor = '#4f46e5',
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

		const maxVal = Math.max(...data, 1);
		const colors = data.map(v =>
			v === maxVal ? highlightColor : color + 'b0'
		);

		chart = new Chart(canvas, {
			type: 'bar',
			data: {
				labels,
				datasets: [{
					label,
					data,
					backgroundColor: colors,
					hoverBackgroundColor: highlightColor,
					borderRadius: 4,
					borderSkipped: false,
				}],
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				animation: { duration: 600, easing: 'easeInOutQuart' },
				plugins: {
					legend: { display: false },
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
								if (!currency) return ` ${v}`;
								if (v >= 1_000_000) return ` KES ${(v / 1_000_000).toFixed(2)}M`;
								if (v >= 1_000) return ` KES ${(v / 1_000).toFixed(1)}K`;
								return ` KES ${v.toLocaleString()}`;
							},
						},
					},
				},
				scales: {
					x: {
						grid: { display: false },
						border: { display: false },
						ticks: { color: '#94a3b8', font: { size: 10 }, maxRotation: 0 },
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
		void data;
		build();
		return () => { chart?.destroy(); chart = null; };
	});

	onDestroy(() => chart?.destroy());
</script>

<div style="position:relative; height:{height}px; width:100%;">
	<canvas bind:this={canvas}></canvas>
</div>
