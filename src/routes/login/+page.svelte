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
	<!-- Left branding panel -->
	<div class="hidden lg:flex lg:w-5/12 xl:w-2/5 flex-col bg-slate-900 p-10">
		<div class="flex items-center gap-3 mb-auto">
			<div class="flex h-9 w-9 items-center justify-center rounded-lg bg-blue-600">
				<Store size={18} class="text-white" />
			</div>
			<span class="text-white font-bold text-base">Maestro POS</span>
		</div>

		<div class="py-12">
			<h1 class="text-3xl font-bold text-white leading-snug mb-3">
				The modern POS<br />built for retail.
			</h1>
			<p class="text-slate-400 text-sm leading-relaxed mb-10">
				Fast checkout, real-time inventory, customer management, and business intelligence — all in one place.
			</p>
			<div class="space-y-3">
				{#each [
					['Instant checkout', 'Process sales in seconds with barcode scanning'],
					['Live inventory', 'Track stock levels and get low-stock alerts'],
					['Business reports', 'Daily, weekly and monthly performance insights'],
				] as [title, desc]}
					<div class="flex items-start gap-3 p-3 rounded-lg bg-slate-800 border border-slate-700">
						<div class="h-1.5 w-1.5 rounded-full bg-blue-500 mt-2 shrink-0"></div>
						<div>
							<p class="text-sm font-semibold text-slate-200">{title}</p>
							<p class="text-xs text-slate-500 mt-0.5">{desc}</p>
						</div>
					</div>
				{/each}
			</div>
		</div>

		<p class="text-xs text-slate-600">© {new Date().getFullYear()} Maestro POS</p>
	</div>

	<!-- Right form panel -->
	<div class="flex flex-1 flex-col items-center justify-center bg-white p-6">
		<!-- Mobile logo -->
		<div class="mb-8 text-center lg:hidden">
			<div class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-blue-600">
				<Store size={24} class="text-white" />
			</div>
			<h1 class="text-xl font-bold text-slate-900">Maestro POS</h1>
			<p class="mt-1 text-sm text-slate-500">Sign in to continue</p>
		</div>

		<div class="w-full max-w-sm">
			<div class="hidden lg:block mb-8">
				<h2 class="text-2xl font-bold text-slate-900">Sign in</h2>
				<p class="mt-1 text-sm text-slate-500">Enter your credentials to access the system</p>
			</div>

			<form onsubmit={login} class="space-y-5">
				<div>
					<label for="email" class="block text-xs font-semibold text-slate-600 uppercase tracking-wide mb-1.5">Email address</label>
					<input
						id="email"
						type="email"
						bind:value={email}
						required
						autocomplete="email"
						class="w-full rounded-lg border border-slate-200 bg-slate-50 px-3.5 py-2.5 text-sm text-slate-900 outline-none transition focus:border-blue-500 focus:bg-white focus:ring-2 focus:ring-blue-500/15"
						placeholder="you@example.com"
					/>
				</div>

				<div>
					<label for="password" class="block text-xs font-semibold text-slate-600 uppercase tracking-wide mb-1.5">Password</label>
					<input
						id="password"
						type="password"
						bind:value={password}
						required
						autocomplete="current-password"
						class="w-full rounded-lg border border-slate-200 bg-slate-50 px-3.5 py-2.5 text-sm text-slate-900 outline-none transition focus:border-blue-500 focus:bg-white focus:ring-2 focus:ring-blue-500/15"
						placeholder="••••••••"
					/>
				</div>

				<button
					type="submit"
					disabled={loading}
					class="w-full rounded-lg bg-blue-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-blue-700 disabled:opacity-60 disabled:cursor-not-allowed transition-colors active:scale-[0.99]"
				>
					{loading ? 'Signing in…' : 'Sign in'}
				</button>
			</form>

			<p class="mt-6 text-center text-xs text-slate-400">
				Secure access · All sessions are encrypted
			</p>
		</div>
	</div>
</div>
