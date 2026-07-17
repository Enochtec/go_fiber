<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { authStore } from '$lib/stores/auth.svelte';
	import { authService } from '$lib/services/auth';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import Notification from '$lib/components/Notification.svelte';
	import { onMount } from 'svelte';
	import {
		Menu, LayoutDashboard, ShoppingCart, Package, Users, MoreHorizontal, Store
	} from '@lucide/svelte';
	import { themeStore } from '$lib/stores/theme.svelte';

	let { children } = $props();
	let drawerOpen = $state(false);

	const bottomNav = [
		{ href: '/dashboard', label: 'Home', icon: LayoutDashboard },
		{ href: '/sales', label: 'Sale', icon: ShoppingCart },
		{ href: '/products', label: 'Products', icon: Package },
		{ href: '/customers', label: 'Customers', icon: Users },
	];

	onMount(async () => {
		themeStore.init();
		if (!authStore.isAuthenticated) {
			goto('/login');
			return;
		}
		if (!authStore.user) {
			try {
				const res = await authService.me();
				if (res.data) authStore.setUser(res.data);
			} catch {
				authStore.clear();
				goto('/login');
			}
		}
	});
</script>

<Notification />

<!-- Mobile overlay -->
{#if drawerOpen}
	<div
		class="fixed inset-0 z-40 bg-black/60 backdrop-blur-sm lg:hidden"
		role="presentation"
		onclick={() => drawerOpen = false}
	></div>
{/if}

<div class="flex h-screen overflow-hidden bg-slate-50 dark:bg-slate-950">
	<!-- Sidebar: always visible on desktop, drawer on mobile -->
	<div
		class="fixed inset-y-0 left-0 z-50 transition-transform duration-300 lg:relative lg:translate-x-0 lg:z-auto"
		class:translate-x-0={drawerOpen}
		class:-translate-x-full={!drawerOpen}
	>
		<Sidebar onclose={() => drawerOpen = false} />
	</div>

	<!-- Main area -->
	<div class="flex flex-1 flex-col min-w-0 overflow-hidden">
		<!-- Mobile top header -->
		<header class="flex items-center gap-3 px-4 py-3 bg-slate-900 border-b border-slate-800 lg:hidden shrink-0">
			<button
				onclick={() => drawerOpen = true}
				class="flex h-8 w-8 items-center justify-center rounded-lg text-slate-400 hover:bg-slate-800 hover:text-slate-200 transition-colors"
			>
				<Menu size={18} />
			</button>
			<div class="flex items-center gap-2">
				<div class="flex h-7 w-7 items-center justify-center rounded-lg bg-blue-600">
					<Store size={13} class="text-white" />
				</div>
				<span class="font-bold text-white text-sm">Maestro POS</span>
			</div>
			{#if authStore.user}
				<div class="ml-auto flex h-7 w-7 items-center justify-center rounded-full bg-blue-600 text-xs font-bold text-white">
					{authStore.user.name.charAt(0).toUpperCase()}
				</div>
			{/if}
		</header>

		<!-- Page content -->
		<main class="flex-1 overflow-y-auto pb-20 lg:pb-0 dark:bg-slate-950">
			{@render children()}
		</main>

		<!-- Mobile bottom navigation -->
		<nav class="fixed bottom-0 left-0 right-0 z-30 lg:hidden bg-slate-900 border-t border-slate-800">
			<div class="flex">
				{#each bottomNav as item}
					{@const active = $page.url.pathname === item.href}
					<a
						href={item.href}
						data-sveltekit-preload-code
						class="flex flex-1 flex-col items-center gap-1 py-2.5 text-xs font-medium transition-colors {active ? 'text-blue-400' : 'text-slate-500 hover:text-slate-300'}"
					>
						<item.icon size={19} />
						{item.label}
					</a>
				{/each}
				<button
					onclick={() => drawerOpen = true}
					class="flex flex-1 flex-col items-center gap-1 py-2.5 text-xs font-medium text-slate-500 hover:text-slate-300 transition-colors"
				>
					<MoreHorizontal size={19} />
					More
				</button>
			</div>
		</nav>
	</div>
</div>
