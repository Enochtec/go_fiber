<script lang="ts">
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/stores/auth.svelte';
	import { authService } from '$lib/services/auth';
	import { notify } from '$lib/stores/notification.svelte';
	import Notification from '$lib/components/Notification.svelte';
	import { Store } from '@lucide/svelte';

	let email = $state('');
	let password = $state('');
	let loading = $state(false);

	async function login(e: SubmitEvent) {
		e.preventDefault();
		loading = true;
		try {
			const res = await authService.login(email, password);
			if (res.data) {
				authStore.set(res.data.user, res.data.token);
				goto('/dashboard');
			}
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Login failed');
		} finally {
			loading = false;
		}
	}
</script>

<Notification />

<div class="flex min-h-screen">
	<!-- Left branding panel (hidden on mobile) -->
	<div class="hidden lg:flex lg:w-1/2 xl:w-3/5 flex-col items-center justify-center bg-gradient-to-br from-slate-900 via-blue-950 to-slate-900 p-12 relative overflow-hidden">
		<!-- Background decoration -->
		<div class="absolute top-0 left-0 right-0 bottom-0 opacity-10">
			<div class="absolute top-1/4 left-1/4 h-64 w-64 rounded-full bg-blue-400 blur-3xl"></div>
			<div class="absolute bottom-1/4 right-1/4 h-48 w-48 rounded-full bg-cyan-400 blur-3xl"></div>
		</div>
		<div class="relative z-10 text-center">
			<div class="mx-auto mb-6 flex h-20 w-20 items-center justify-center rounded-2xl bg-gradient-to-br from-blue-400 to-blue-600 shadow-2xl shadow-blue-900/50">
				<Store size={40} class="text-white" />
			</div>
			<h1 class="text-4xl font-bold text-white mb-3">POS System</h1>
			<p class="text-blue-300 text-lg">Modern Point of Sale</p>
			<div class="mt-10 grid grid-cols-3 gap-4 text-center">
				{#each [['Fast', 'Instant checkout'], ['Smart', 'Real-time reports'], ['Simple', 'Easy to use']] as [t, s]}
					<div class="rounded-xl bg-white/5 border border-white/10 p-4">
						<p class="text-white font-semibold text-sm">{t}</p>
						<p class="text-blue-300 text-xs mt-1">{s}</p>
					</div>
				{/each}
			</div>
		</div>
	</div>

	<!-- Right form panel -->
	<div class="flex flex-1 flex-col items-center justify-center bg-slate-50 p-6">
		<!-- Mobile logo -->
		<div class="mb-8 text-center lg:hidden">
			<div class="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-blue-500 to-blue-700 shadow-lg">
				<Store size={28} class="text-white" />
			</div>
			<h1 class="text-2xl font-bold text-slate-900">POS System</h1>
			<p class="mt-1 text-sm text-slate-500">Sign in to continue</p>
		</div>

		<div class="w-full max-w-sm">
			<div class="hidden lg:block mb-8">
				<h2 class="text-2xl font-bold text-slate-900">Welcome back</h2>
				<p class="mt-1 text-sm text-slate-500">Sign in to your account</p>
			</div>

			<form onsubmit={login} class="rounded-2xl bg-white p-8 shadow-sm border border-slate-200">
				<div class="space-y-5">
					<div>
						<label for="email" class="block text-sm font-semibold text-slate-700 mb-1.5">Email address</label>
						<input
							id="email"
							type="email"
							bind:value={email}
							required
							autocomplete="email"
							class="w-full rounded-xl border border-slate-300 px-4 py-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20"
							placeholder="you@example.com"
						/>
					</div>

					<div>
						<label for="password" class="block text-sm font-semibold text-slate-700 mb-1.5">Password</label>
						<input
							id="password"
							type="password"
							bind:value={password}
							required
							autocomplete="current-password"
							class="w-full rounded-xl border border-slate-300 px-4 py-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20"
							placeholder="••••••••"
						/>
					</div>

					<button
						type="submit"
						disabled={loading}
						class="w-full rounded-xl bg-gradient-to-r from-blue-600 to-blue-700 px-4 py-3 text-sm font-semibold text-white shadow-md shadow-blue-500/20 hover:from-blue-700 hover:to-blue-800 disabled:opacity-60 disabled:cursor-not-allowed transition-all active:scale-[0.98]"
					>
						{loading ? 'Signing in…' : 'Sign in'}
					</button>
				</div>
			</form>
		</div>
	</div>
</div>
