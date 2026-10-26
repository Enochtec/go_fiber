<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { authStore } from '$lib/stores/auth.svelte';
	import { authService } from '$lib/services/auth';
	import { shopService } from '$lib/services/shop';
	import { refreshLocalCache } from '$lib/services/offline';
	import { shopStore } from '$lib/stores/shop.svelte';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import Notification from '$lib/components/Notification.svelte';
	import SyncStatus from '$lib/components/SyncStatus.svelte';
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
				return;
			}
		}
		try {
			const info = await shopService.getInfo();
			shopStore.setName(info.shop.name);
		} catch {
		}
		await refreshLocalCache();
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
		<Sidebar onclose={() => drawerOpen = false} shopName={shopStore.name} />
	</div>

	<!-- Main area -->
	<div class="flex flex-1 flex-col min-w-0 overflow-hidden">
		<!-- Mobile top header -->
		<header class="flex items-center gap-3 px-4 py-3 bg-sidebar border-b border-white/20 lg:hidden shrink-0">
			<button
				onclick={() => drawerOpen = true}
				class="flex h-8 w-8 items-center justify-center rounded-lg text-white/75 hover:bg-white/15 hover:text-white transition-colors"
			>
				<Menu size={18} />
			</button>
			<div class="flex items-center gap-2">
				<div class="flex h-7 w-7 items-center justify-center">
					<Store size={13} class="text-white" />
				</div>
				<span class="font-bold text-white text-sm truncate" title={shopStore.name}>{shopStore.name}</span>
			</div>
			<div class="ml-auto flex items-center gap-2">
				<SyncStatus />
				{#if authStore.user}
					<div class="flex h-7 w-7 items-center justify-center rounded-full bg-blue-600 text-xs font-bold text-white">
						{authStore.user.name.charAt(0).toUpperCase()}
					</div>
				{/if}
			</div>
		</header>

		<!-- Page content -->
		<main class="flex-1 overflow-y-auto pb-20 lg:pb-0 dark:bg-slate-950">
			{@render children()}
		</main>

		<!-- Mobile bottom navigation -->
		<nav class="fixed bottom-0 left-0 right-0 z-30 lg:hidden bg-sidebar border-t border-white/20 pb-[env(safe-area-inset-bottom,0px)]">
			<div class="flex">
				{#each bottomNav as item}
					{@const active = $page.url.pathname === item.href}
					<a
						href={item.href}
						data-sveltekit-preload-code
						class="flex flex-1 flex-col items-center gap-0.5 py-2.5 text-[10px] font-medium transition-colors {active ? 'text-white' : 'text-white/60 hover:text-white'}"
					>
						<div class="h-9 w-9 flex items-center justify-center {active ? 'bg-white/20 rounded-lg' : ''}">
							<item.icon size={20} />
						</div>
						{item.label}
					</a>
				{/each}
				<button
					onclick={() => drawerOpen = true}
					class="flex flex-1 flex-col items-center gap-0.5 py-2.5 text-[10px] font-medium text-white/60 hover:text-white transition-colors"
				>
					<div class="h-9 w-9 flex items-center justify-center">
						<MoreHorizontal size={20} />
					</div>
					More
				</button>
			</div>
		</nav>
	</div>
</div>
