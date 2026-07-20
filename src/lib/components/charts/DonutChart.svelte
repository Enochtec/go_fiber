<script lang="ts">
	import { onDestroy } from 'svelte';
	import Chart from 'chart.js/auto';

	interface Props {
		labels: string[];
		data: number[];
		colors: string[];
		centerLabel?: string;
		centerValue?: string;
		height?: number;
		cutout?: string;
	}

	let {
		labels, data, colors,
		centerLabel = '',
		centerValue = '',
		height = 200,
		cutout = '70%',
	}: Props = $props();

	let canvas = $state<HTMLCanvasElement | undefined>(undefined);
	let chart: Chart | null = null;

	function build() {
		if (!canvas) return;
		chart?.destroy();

		chart = new Chart(canvas, {
			type: 'doughnut',
			data: {
				labels,
				datasets: [{
					data,
					backgroundColor: colors,
					hoverBackgroundColor: colors.map(c => c + 'ee'),
					borderWidth: 2,
					borderColor: '#ffffff',
					hoverOffset: 6,
				}],
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				cutout,
				animation: { duration: 700, easing: 'easeInOutQuart' },
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
								const total = (c.dataset.data as number[]).reduce((a, b) => a + b, 0);
								const pct = total > 0 ? ((c.parsed / total) * 100).toFixed(1) : '0';
								const v = c.parsed;
								if (v >= 1_000_000) return ` KES ${(v / 1_000_000).toFixed(2)}M (${pct}%)`;
								if (v >= 1_000) return ` KES ${(v / 1_000).toFixed(1)}K (${pct}%)`;
								return ` KES ${v.toLocaleString()} (${pct}%)`;
							},
						},
					},
				},
			},
			plugins: centerLabel ? [{
				id: 'centerText',
				beforeDraw(ch) {
					const { width, height, ctx } = ch;
					ctx.save();
					ctx.textAlign = 'center';
					ctx.textBaseline = 'middle';
					const cx = width / 2;
					const cy = height / 2;
					ctx.font = 'bold 14px Inter, system-ui, sans-serif';
					ctx.fillStyle = '#1e293b';
					ctx.fillText(centerValue, cx, cy - 8);
					ctx.font = '10px Inter, system-ui, sans-serif';
					ctx.fillStyle = '#94a3b8';
					ctx.fillText(centerLabel.toUpperCase(), cx, cy + 10);
					ctx.restore();
				},
			}] : [],
		});
	}

	$effect(() => {
		void labels; void data; void colors; void centerValue;
		build();
		return () => { chart?.destroy(); chart = null; };
	});

	onDestroy(() => chart?.destroy());
</script>

<div style="position:relative; height:{height}px; width:100%;">
	<canvas bind:this={canvas}></canvas>
</div>
