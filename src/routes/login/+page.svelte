<script lang="ts">
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/stores/auth.svelte';
	import { authService } from '$lib/services/auth';
	import { notify } from '$lib/stores/notification.svelte';
	import Notification from '$lib/components/Notification.svelte';
	import { Store, LoaderCircle } from '@lucide/svelte';

	let email = $state('');
	let password = $state('');
	let remember = $state(false);
	let loading = $state(false);

	async function login(e: SubmitEvent) {
		e.preventDefault();
		loading = true;
		try {
			const res = await authService.login(email.trim(), password);
			if (res.data) {
				authStore.set(res.data.user, res.data.token);
				if (remember) localStorage.setItem('pos_remember', email);
				else localStorage.removeItem('pos_remember');
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
	<div class="hidden lg:flex lg:w-5/12 xl:w-2/5 flex-col relative overflow-hidden" style="background:linear-gradient(135deg,#1a0a5e 0%,#3F00FF 50%,#3200CC 100%);">
		<div class="absolute inset-0 opacity-10" style="background-image:radial-gradient(circle at 20% 30%, white 0%, transparent 50%), radial-gradient(circle at 80% 70%, white 0%, transparent 50%);"></div>
		<div class="relative p-10 flex flex-col h-full">
			<div class="flex items-center gap-3 mb-auto">
				<div class="flex h-9 w-9 items-center justify-center">
					<Store size={20} class="text-white/90" />
				</div>
				<span class="text-white/90 font-bold text-base">Maestro POS</span>
			</div>

			<div class="py-12">
				<h1 class="text-4xl font-bold text-white leading-tight mb-4">
					The modern POS<br />built for retail.
				</h1>
				<p class="text-indigo-200 text-sm leading-relaxed mb-10 max-w-sm">
					Fast checkout, real-time inventory, customer management, and business intelligence — all in one place.
				</p>
				<div class="space-y-3">
					{#each [
						['Instant checkout', 'Process sales in seconds with barcode scanning'],
						['Live inventory', 'Track stock levels and get low-stock alerts'],
						['Business reports', 'Daily, weekly and monthly performance insights'],
					] as [title, desc]}
						<div class="flex items-start gap-3.5 p-3.5 rounded-lg bg-white/5 backdrop-blur-sm border border-white/10">
							<div class="h-2 w-2 rounded-full bg-indigo-300 mt-1.5 shrink-0"></div>
							<div>
								<p class="text-sm font-semibold text-white">{title}</p>
								<p class="text-xs text-indigo-300 mt-0.5">{desc}</p>
							</div>
						</div>
					{/each}
				</div>
			</div>

			<p class="text-xs text-indigo-300/60">© {new Date().getFullYear()} Maestro POS</p>
		</div>
	</div>

	<!-- Right form panel -->
	<div class="flex flex-1 flex-col items-center justify-center bg-white dark:bg-slate-950 p-6">
		<!-- Mobile logo -->
		<div class="mb-8 text-center lg:hidden">
			<div class="mx-auto mb-4 flex h-12 w-12 items-center justify-center">
				<Store size={28} class="text-blue-600" />
			</div>
			<h1 class="text-xl font-bold text-slate-900 dark:text-slate-100">Maestro POS</h1>
			<p class="mt-1 text-sm text-slate-500">Sign in to continue</p>
		</div>

		<div class="w-full max-w-sm">
			<div class="hidden lg:block mb-8">
				<h2 class="text-2xl font-bold text-slate-900 dark:text-slate-100">Welcome back</h2>
				<p class="mt-1.5 text-sm text-slate-500 dark:text-slate-400">Enter your credentials to access the system</p>
			</div>

			<form onsubmit={login} class="space-y-5">
				<div>
					<label for="email" class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Email address</label>
					<input
						id="email"
						type="email"
						bind:value={email}
						required
						autocomplete="email"
						class="w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 px-3.5 py-2.5 text-sm text-slate-900 dark:text-slate-100 outline-none transition focus:border-blue-600 focus:bg-white dark:focus:bg-slate-800 focus:ring-2 focus:ring-blue-600/15"
						placeholder="you@example.com"
					/>
				</div>

				<div>
					<label for="password" class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Password</label>
					<input
						id="password"
						type="password"
						bind:value={password}
						required
						autocomplete="current-password"
						class="w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 px-3.5 py-2.5 text-sm text-slate-900 dark:text-slate-100 outline-none transition focus:border-blue-600 focus:bg-white dark:focus:bg-slate-800 focus:ring-2 focus:ring-blue-600/15"
						placeholder="••••••••"
					/>
				</div>

				<div class="flex items-center justify-between">
					<label class="flex items-center gap-2 cursor-pointer">
						<input type="checkbox" bind:checked={remember}
							class="h-4 w-4 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-600/20 dark:bg-slate-800" />
						<span class="text-xs text-slate-600 dark:text-slate-400">Remember me</span>
					</label>
					<a href="/forgot-password" class="text-xs font-semibold text-blue-600 hover:text-blue-700 dark:text-blue-400">Forgot password?</a>
				</div>

				<button
					type="submit"
					disabled={loading}
					class="w-full rounded-lg px-4 py-2.5 text-sm font-semibold text-white disabled:opacity-60 disabled:cursor-not-allowed transition-all active:scale-[0.98]"
					style="background:linear-gradient(135deg,#3F00FF,#3200CC);"
				>
					{#if loading}
						<span class="inline-flex items-center gap-2"><LoaderCircle size={15} class="animate-spin" /> Signing in…</span>
					{:else}
						Sign in
					{/if}
				</button>
			</form>

			<div class="mt-6 text-center">
				<p class="text-xs text-slate-400 dark:text-slate-500">
					Don't have an account?
					<a href="/register" class="font-semibold text-blue-600 hover:text-blue-700 dark:text-blue-400">Create one</a>
				</p>
			</div>

			<p class="mt-4 text-center text-xs text-slate-400 dark:text-slate-500">
				Secure access · All sessions are encrypted
			</p>
		</div>
	</div>
</div>
