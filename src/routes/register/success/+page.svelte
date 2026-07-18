<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { authStore } from '$lib/stores/auth.svelte';
	import { onMount } from 'svelte';
	import { Store, CheckCircle2, ArrowRight, Sparkles } from '@lucide/svelte';

	let shopName = $state('');
	let ownerName = $state('');

	onMount(() => {
		if (!$page.state || !($page.state as Record<string, unknown>).shop) {
			if (authStore.isAuthenticated) {
				goto('/dashboard');
			} else {
				goto('/login');
			}
			return;
		}
		const state = $page.state as Record<string, unknown>;
		shopName = (state.shop as { name: string }).name;
		ownerName = (state.user as { name: string }).name;
	});
</script>

<svelte:head><title>Welcome — Maestro POS</title></svelte:head>

<div class="flex min-h-screen items-center justify-center bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 p-4">
	<div class="w-full max-w-md text-center">
		<!-- Success icon -->
		<div class="mx-auto mb-6 flex h-20 w-20 items-center justify-center rounded-full bg-emerald-500/15 animate-[scale-in_0.4s_ease-out]">
			<div class="flex h-14 w-14 items-center justify-center rounded-full bg-emerald-500 shadow-lg shadow-emerald-500/30">
				<CheckCircle2 size={32} class="text-white" />
			</div>
		</div>

		<!-- Title -->
		<h1 class="text-2xl font-bold text-white mb-2 animate-[fade-in_0.5s_ease-out]">
			Registration Successful
		</h1>
		<p class="text-slate-400 text-sm mb-8 animate-[fade-in_0.6s_ease-out]">
			Welcome to Maestro POS
		</p>

		<!-- Cards -->
		<div class="space-y-3 mb-8 animate-[fade-in_0.7s_ease-out]">
			<div class="flex items-center gap-3 rounded-xl bg-white/10 backdrop-blur-sm border border-white/10 p-4">
				<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-blue-500/20">
					<Store size={18} class="text-blue-400" />
				</div>
				<div class="text-left min-w-0">
					<p class="text-[10px] font-semibold uppercase tracking-wide text-slate-500">Shop</p>
					<p class="text-sm font-bold text-white truncate">{shopName}</p>
				</div>
			</div>

			<div class="flex items-center gap-3 rounded-xl bg-white/10 backdrop-blur-sm border border-white/10 p-4">
				<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-emerald-500/20">
					<Sparkles size={18} class="text-emerald-400" />
				</div>
				<div class="text-left min-w-0">
					<p class="text-[10px] font-semibold uppercase tracking-wide text-slate-500">Administrator</p>
					<p class="text-sm font-bold text-white truncate">{ownerName}</p>
				</div>
			</div>
		</div>

		<!-- Action -->
		<button onclick={() => goto('/dashboard')}
			class="inline-flex items-center gap-2 rounded-xl bg-blue-600 px-6 py-3 text-sm font-semibold text-white hover:bg-blue-700 transition-all active:scale-[0.98] shadow-lg shadow-blue-600/20 animate-[fade-in_0.8s_ease-out]">
			Go to Dashboard <ArrowRight size={16} />
		</button>

		<p class="text-xs text-slate-600 mt-8 animate-[fade-in_0.9s_ease-out]">
			Your shop is ready. You're now signed in as the administrator.
		</p>
	</div>
</div>

<style>
	@keyframes scale-in {
		from { transform: scale(0); opacity: 0; }
		to { transform: scale(1); opacity: 1; }
	}
	@keyframes fade-in {
		from { opacity: 0; transform: translateY(8px); }
		to { opacity: 1; transform: translateY(0); }
	}
</style>
