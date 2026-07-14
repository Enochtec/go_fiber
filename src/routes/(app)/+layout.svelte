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

<div class="flex h-screen overflow-hidden bg-slate-100 dark:bg-slate-950">
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
		<header class="flex items-center gap-3 px-4 py-3 shadow-md lg:hidden shrink-0" style="background-color: #005A9C;">
			<button
				onclick={() => drawerOpen = true}
				class="flex h-9 w-9 items-center justify-center rounded-lg text-white/70 hover:bg-white/10 active:bg-white/20"
			>
				<Menu size={20} />
			</button>
			<div class="flex items-center gap-2">
				<div class="flex h-7 w-7 items-center justify-center rounded-xl bg-white/20">
					<Store size={14} class="text-white" />
				</div>
				<span class="font-bold text-white text-sm">POS System</span>
			</div>
			{#if authStore.user}
				<div class="ml-auto flex h-8 w-8 items-center justify-center rounded-full bg-white text-xs font-bold" style="color: #005A9C;">
					{authStore.user.name.charAt(0).toUpperCase()}
				</div>
			{/if}
		</header>

		<!-- Page content -->
		<main class="flex-1 overflow-y-auto pb-20 lg:pb-0 dark:bg-slate-950">
			{@render children()}
		</main>

		<!-- Mobile bottom navigation -->
		<nav class="fixed bottom-0 left-0 right-0 z-30 lg:hidden shadow-lg" style="background-color: #005A9C;">
			<div class="flex">
				{#each bottomNav as item}
					{@const active = $page.url.pathname === item.href}
					<a
						href={item.href}
						class="flex flex-1 flex-col items-center gap-1 py-3 text-xs font-medium transition-colors {active ? 'text-white' : 'text-white/50'}"
					>
						<item.icon size={20} />
						{item.label}
					</a>
				{/each}
				<button
					onclick={() => drawerOpen = true}
					class="flex flex-1 flex-col items-center gap-1 py-3 text-xs font-medium text-white/50 hover:text-white"
				>
					<MoreHorizontal size={20} />
					More
				</button>
			</div>
		</nav>
	</div>
</div>
