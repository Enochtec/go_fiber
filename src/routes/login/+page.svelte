<script lang="ts">
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/stores/auth.svelte';
	import { authService } from '$lib/services/auth';
	import { notify } from '$lib/stores/notification.svelte';
	import Notification from '$lib/components/Notification.svelte';
	import { Store, LoaderCircle, ShieldCheck, ArrowUpRight } from '@lucide/svelte';

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

	<div class="min-h-[100svh] bg-[#f4f7fb] text-slate-900 lg:grid lg:grid-cols-[minmax(360px,0.9fr)_1.1fr]">
	<!-- Brand panel -->
	<div class="relative hidden overflow-hidden bg-[#08152f] lg:flex lg:min-h-screen lg:flex-col lg:justify-between lg:p-12 xl:p-16">
		<div class="absolute -right-28 -top-28 h-96 w-96 rounded-full border border-blue-400/20"></div>
		<div class="absolute -bottom-40 -left-32 h-[30rem] w-[30rem] rounded-full border border-blue-400/10"></div>
		<div class="absolute inset-0 opacity-20" style="background-image:linear-gradient(rgba(85,146,255,.18) 1px, transparent 1px),linear-gradient(90deg,rgba(85,146,255,.18) 1px, transparent 1px);background-size:42px 42px;"></div>
		<div class="relative">
			<div class="flex items-center gap-3">
				<div class="flex h-10 w-10 items-center justify-center rounded-xl bg-[#0000CD] text-white shadow-lg shadow-blue-950/40">
					<Store size={20} />
				</div>
				<div>
					<p class="text-sm font-bold tracking-wide text-white">Maestro POS</p>
					<p class="text-[11px] uppercase tracking-[0.2em] text-blue-200/70">Retail operations</p>
				</div>
			</div>
		</div>

		<div class="relative max-w-lg py-16">
			<div class="mb-6 inline-flex items-center gap-2 rounded-full border border-blue-300/20 bg-blue-300/10 px-3 py-1.5 text-[11px] font-semibold uppercase tracking-[0.16em] text-blue-100">
				<span class="h-1.5 w-1.5 rounded-full bg-[#6ea8ff]"></span>
				Your store, in sync
			</div>
			<h1 class="max-w-md text-5xl font-bold leading-[1.04] tracking-tight text-white xl:text-6xl">Make every sale count.</h1>
			<p class="mt-6 max-w-md text-base leading-7 text-blue-100/70">A focused workspace for checkout, stock, customers, and the decisions that keep your business moving.</p>

			<div class="mt-10 flex items-center gap-3 border-t border-white/10 pt-5 text-sm text-blue-100/80">
				<ShieldCheck size={18} class="text-blue-300" />
				<span>Protected access for your team</span>
				<ArrowUpRight size={15} class="ml-auto text-blue-300/70" />
			</div>
		</div>

		<p class="relative text-xs text-blue-100/45">© {new Date().getFullYear()} Maestro POS</p>
	</div>

	<!-- Form panel -->
	<div class="flex min-h-[100svh] flex-col items-center justify-center px-4 py-8 sm:px-8 sm:py-10">
		<div class="w-full max-w-[430px]">
			<div class="mb-8 flex items-center gap-3 lg:hidden">
				<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-[#0000CD] text-white shadow-lg shadow-blue-900/20">
					<Store size={21} />
				</div>
				<div>
					<p class="text-sm font-bold tracking-wide text-slate-900">Maestro POS</p>
					<p class="mt-0.5 text-xs uppercase tracking-[0.18em] text-slate-400">Retail operations</p>
				</div>
			</div>

			<div class="mb-7 sm:mb-8">
				<p class="mb-3 text-xs font-semibold uppercase tracking-[0.18em] text-[#0000CD]">Secure sign in</p>
				<h2 class="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">Welcome back</h2>
				<p class="mt-2 text-sm leading-6 text-slate-500">Sign in to continue to your workspace.</p>
			</div>

			<form onsubmit={login} class="space-y-4 sm:space-y-5">
				<div>
					<label for="email" class="mb-2 block text-sm font-semibold text-slate-700">Email address</label>
					<input
						id="email"
						type="email"
						bind:value={email}
						required
						autocomplete="email"
						class="w-full rounded-lg border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 shadow-sm outline-none transition placeholder:text-slate-400 focus:border-[#0000CD] focus:ring-4 focus:ring-blue-600/10"
						placeholder="you@example.com"
					/>
				</div>

				<div>
					<label for="password" class="mb-2 block text-sm font-semibold text-slate-700">Password</label>
					<input
						id="password"
						type="password"
						bind:value={password}
						required
						autocomplete="current-password"
						class="w-full rounded-lg border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 shadow-sm outline-none transition placeholder:text-slate-400 focus:border-[#0000CD] focus:ring-4 focus:ring-blue-600/10"
						placeholder="••••••••"
					/>
				</div>

				<div class="flex flex-wrap items-center justify-between gap-x-4 gap-y-2">
					<label class="flex items-center gap-2 cursor-pointer">
						<input type="checkbox" bind:checked={remember}
							class="h-4 w-4 rounded border-slate-300 text-[#0000CD] focus:ring-blue-600/20" />
						<span class="text-sm text-slate-500">Remember me</span>
					</label>
					<a href="/forgot-password" class="text-sm font-semibold text-[#0000CD] hover:text-blue-800">Forgot password?</a>
				</div>

				<button
					type="submit"
					disabled={loading}
					class="w-full rounded-lg bg-[#0000CD] px-4 py-3 text-sm font-semibold text-white shadow-lg shadow-blue-900/20 transition-all hover:bg-blue-800 hover:shadow-blue-900/30 active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-60"
				>
					{#if loading}
						<span class="inline-flex items-center gap-2"><LoaderCircle size={15} class="animate-spin" /> Signing in…</span>
					{:else}
						Sign in
					{/if}
				</button>
			</form>

			<div class="mt-7 text-center">
				<p class="text-sm text-slate-500">
					Don't have an account?
					<a href="/register" class="font-semibold text-[#0000CD] hover:text-blue-800">Create one</a>
				</p>
			</div>

			<p class="mt-8 text-center text-xs text-slate-400">
				Your session is protected with encrypted access
			</p>
		</div>
	</div>
</div>
