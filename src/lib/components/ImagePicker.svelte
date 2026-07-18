<script lang="ts">
	import { uploadImage } from '$lib/services/upload';
	import { Camera, Upload, Trash2, LoaderCircle, Image } from '@lucide/svelte';

	interface Props {
		value: string | null | undefined;
		onchange: (url: string | null) => void;
	}

	let { value, onchange }: Props = $props();

	let preview = $state<string | null>(null);
	let uploading = $state(false);
	let error = $state('');
	let dragOver = $state(false);

	let fileInput: HTMLInputElement | undefined = $state();
	let cameraInput: HTMLInputElement | undefined = $state();

	function openFilePicker() {
		fileInput?.click();
	}

	function openCamera() {
		cameraInput?.click();
	}

	function handleFile(file: File) {
		error = '';

		const valid = ['image/png', 'image/jpeg', 'image/webp'];
		if (!valid.includes(file.type)) {
			error = 'Unsupported format. Use PNG, JPG, or WEBP.';
			return;
		}
		if (file.size > 5 * 1024 * 1024) {
			error = 'File too large (max 5MB).';
			return;
		}

		preview = URL.createObjectURL(file);
		upload(file);
	}

	async function upload(file: File) {
		uploading = true;
		error = '';
		try {
			const url = await uploadImage(file);
			onchange(url);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Upload failed';
			preview = null;
		} finally {
			uploading = false;
		}
	}

	function remove() {
		preview = null;
		onchange(null);
		error = '';
	}

	let hasCamera = $state(false);
	$effect(() => {
		if (typeof navigator !== 'undefined' && 'mediaDevices' in navigator) {
			navigator.mediaDevices?.enumerateDevices().then(devices => {
				hasCamera = devices.some(d => d.kind === 'videoinput');
			}).catch(() => { hasCamera = false; });
		}
	});

	function onDrop(e: DragEvent) {
		e.preventDefault();
		dragOver = false;
		const f = e.dataTransfer?.files?.[0];
		if (f) handleFile(f);
	}

	function onDragOver(e: DragEvent) {
		e.preventDefault();
		dragOver = true;
	}

	function onDragLeave() {
		dragOver = false;
	}
</script>

<div class="space-y-2">
	<!-- Current image preview -->
	{#if value || preview}
		<div class="relative overflow-hidden rounded-xl bg-slate-50 border border-slate-200">
			<img
				src={preview || value || ''}
				alt="Product preview"
				class="w-full h-48 object-contain"
				loading="lazy"
			/>
			<button
				onclick={remove}
				disabled={uploading}
				class="absolute top-2 right-2 flex h-7 w-7 items-center justify-center rounded-lg bg-white/90 text-red-500 hover:bg-red-500 hover:text-white shadow-sm transition-colors disabled:opacity-50"
				title="Remove image"
			>
				<Trash2 size={13} />
			</button>
			{#if uploading}
				<div class="absolute inset-0 flex items-center justify-center bg-white/60 rounded-xl">
					<div class="flex items-center gap-2 text-sm font-medium text-slate-600">
						<LoaderCircle size={16} class="animate-spin" /> Uploading…
					</div>
				</div>
			{/if}
		</div>
	{:else}
		<!-- Drop zone -->
		<div
			role="button"
			tabindex="0"
			onclick={openFilePicker}
			ondrop={onDrop}
			ondragover={onDragOver}
			ondragleave={onDragLeave}
			onkeydown={(e) => e.key === 'Enter' && openFilePicker()}
			class="flex cursor-pointer flex-col items-center justify-center gap-2 rounded-xl border-2 border-dashed p-8 transition-colors {dragOver ? 'border-blue-400 bg-blue-50' : 'border-slate-300 hover:border-blue-400 hover:bg-slate-50'}"
		>
			<div class="flex h-12 w-12 items-center justify-center rounded-full bg-slate-100">
				<Image size={22} class="text-slate-400" />
			</div>
			<div class="text-center">
				<p class="text-sm font-semibold text-slate-600">Drop image here or click to browse</p>
				<p class="text-xs text-slate-400 mt-0.5">PNG, JPG, WEBP — max 5MB</p>
			</div>
		</div>
	{/if}

	<!-- Action buttons -->
	<div class="flex gap-2">
		<button
			type="button"
			onclick={openFilePicker}
			disabled={uploading}
			class="flex items-center gap-1.5 rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-600 hover:bg-slate-50 transition-colors disabled:opacity-50"
		>
			<Upload size={13} /> Browse Files
		</button>
		{#if hasCamera}
			<button
				type="button"
				onclick={openCamera}
				disabled={uploading}
				class="flex items-center gap-1.5 rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-600 hover:bg-slate-50 transition-colors disabled:opacity-50"
			>
				<Camera size={13} /> {typeof navigator !== 'undefined' && /Mobi|Android/i.test(navigator.userAgent) ? 'Capture' : 'Camera'}
			</button>
		{/if}
	</div>

	<!-- Hidden inputs -->
	<input
		type="file"
		accept="image/png,image/jpeg,image/webp"
		class="hidden"
		bind:this={fileInput}
		onchange={(e) => {
			const f = (e.target as HTMLInputElement).files?.[0];
			if (f) handleFile(f);
			(e.target as HTMLInputElement).value = '';
		}}
	/>
	<input
		type="file"
		accept="image/png,image/jpeg,image/webp"
		capture="environment"
		class="hidden"
		bind:this={cameraInput}
		onchange={(e) => {
			const f = (e.target as HTMLInputElement).files?.[0];
			if (f) handleFile(f);
			(e.target as HTMLInputElement).value = '';
		}}
	/>

	<!-- Error -->
	{#if error}
		<p class="text-xs text-red-500">{error}</p>
	{/if}
</div>
