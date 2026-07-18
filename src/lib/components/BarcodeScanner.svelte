<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Html5Qrcode } from 'html5-qrcode';
	import { X, CameraOff, LoaderCircle, Scan, Smartphone } from '@lucide/svelte';

	interface Props {
		onscan: (barcode: string) => void;
		onclose: () => void;
	}

	let { onscan, onclose }: Props = $props();

	let scanner: Html5Qrcode | null = null;
	let status = $state<'init' | 'scanning' | 'error' | 'denied'>('init');
	let errorMsg = $state('');
	let cameraCount = $state(0);
	let facingMode: 'environment' | 'user' = 'environment';

	const SCANNER_ID = 'barcode-scanner';

	onMount(async () => {
		try {
			const devices = await Html5Qrcode.getCameras();
			cameraCount = devices.length;
			if (cameraCount === 0) {
				status = 'error';
				errorMsg = 'No camera found. Use the search field to find products by name or barcode.';
				return;
			}
			await startScanning();
		} catch (e) {
			status = 'error';
			errorMsg = 'Camera access denied. Please enable camera permissions in your browser settings, or use the search field.';
		}
	});

	async function startScanning() {
		status = 'init';
		errorMsg = '';
		scanner = new Html5Qrcode(SCANNER_ID);

		const config = {
			fps: 10,
			qrbox: { width: 280, height: 120 },
			aspectRatio: 1.7777778,
		};

		try {
			await scanner.start(
				{ facingMode },
				config,
				onScanSuccess,
				() => {}
			);
			status = 'scanning';
		} catch (e) {
			if (facingMode === 'environment' && cameraCount > 1) {
				facingMode = 'user';
				try {
					await scanner.start({ facingMode }, config, onScanSuccess, () => {});
					status = 'scanning';
					return;
				} catch {}
			}
			status = 'denied';
			errorMsg = 'Camera permission denied. Enable camera access in your browser settings, or use the search field.';
		}
	}

	function onScanSuccess(decodedText: string) {
		stopScanner();
		onscan(decodedText);
	}

	async function stopScanner() {
		if (scanner) {
			try { await scanner.stop(); } catch {}
			scanner = null;
		}
	}

	async function switchCamera() {
		facingMode = facingMode === 'environment' ? 'user' : 'environment';
		await stopScanner();
		await startScanning();
	}

	onDestroy(() => { stopScanner(); });
</script>

<!-- Overlay -->
<div class="fixed inset-0 z-[100] flex flex-col bg-black">
	<!-- Header -->
	<div class="flex items-center justify-between px-4 py-3 bg-black/80 shrink-0">
		<button onclick={onclose} class="flex h-9 w-9 items-center justify-center rounded-lg text-white/70 hover:bg-white/10">
			<X size={20} />
		</button>
		<h2 class="text-sm font-semibold text-white">Scan Barcode</h2>
		<div class="w-9"></div>
	</div>

	<!-- Scanner area -->
	<div class="flex-1 relative flex items-center justify-center">
		{#if status === 'init'}
			<div class="flex flex-col items-center gap-3 text-white/80">
				<LoaderCircle size={28} class="animate-spin" />
				<p class="text-sm">Initializing camera…</p>
			</div>
		{:else if status === 'denied' || status === 'error'}
			<div class="flex flex-col items-center gap-4 px-8 text-center">
				<div class="flex h-16 w-16 items-center justify-center rounded-full bg-white/10">
					<CameraOff size={28} class="text-white/60" />
				</div>
				<p class="text-white font-semibold text-sm">Camera Unavailable</p>
				<p class="text-white/60 text-xs max-w-xs">{errorMsg}</p>
				<button onclick={onclose} class="rounded-lg bg-white/10 px-5 py-2.5 text-sm font-semibold text-white hover:bg-white/20 transition-colors">
					Use Manual Search
				</button>
			</div>
		{:else}
			<div id={SCANNER_ID} class="w-full h-full"></div>

			<!-- Scanning overlay frame -->
			<div class="absolute inset-0 pointer-events-none flex items-center justify-center">
				<div class="relative w-64 h-28">
					<div class="absolute inset-0 rounded-2xl border-2 border-blue-400/60"></div>
					<div class="absolute top-0 left-1/2 -translate-x-1/2 -translate-y-3">
						<div class="flex items-center gap-1.5 rounded-full bg-blue-600 px-3 py-1 shadow-lg">
							<Scan size={11} class="text-white" />
							<span class="text-[10px] font-semibold text-white">Scanning</span>
						</div>
					</div>
				</div>
			</div>
		{/if}
	</div>

	<!-- Footer controls -->
	<div class="flex items-center justify-center gap-4 px-4 py-5 bg-black/80 shrink-0">
		{#if status === 'scanning' && cameraCount > 1}
			<button onclick={switchCamera} class="flex flex-col items-center gap-1 text-white/60 hover:text-white transition-colors">
				<div class="flex h-10 w-10 items-center justify-center rounded-full bg-white/10">
					<Smartphone size={18} />
				</div>
				<span class="text-[10px] font-medium">Switch</span>
			</button>
		{/if}

		<button onclick={onclose} class="flex flex-col items-center gap-1 text-white/60 hover:text-white transition-colors">
			<div class="flex h-10 w-10 items-center justify-center rounded-full bg-white/10">
				<X size={18} />
			</div>
			<span class="text-[10px] font-medium">Cancel</span>
		</button>
	</div>
</div>
