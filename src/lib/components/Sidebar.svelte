<script lang="ts">
	import { page } from '$app/stores';
	import { authStore } from '$lib/stores/auth.svelte';
	import {
		LayoutDashboard,
		ShoppingCart,
		Package,
		Warehouse,
		Users,
		Truck,
		ShoppingBag,
		BarChart2,
		UserCog,
		Settings,
		LogOut,
		Store,
		X,
		Sun,
		Moon
	} from '@lucide/svelte';
	import { goto } from '$app/navigation';
	import { themeStore } from '$lib/stores/theme.svelte';

	interface Props {
		onclose?: () => void;
	}
	let { onclose }: Props = $props();

	const links = [
		{ href: '/dashboard', label: 'Dashboard', icon: LayoutDashboard, roles: ['admin', 'manager', 'cashier'] },
		{ href: '/sales', label: 'New Sale', icon: ShoppingCart, roles: ['admin', 'manager', 'cashier'] },
		{ href: '/sales/history', label: 'Sales History', icon: BarChart2, roles: ['admin', 'manager', 'cashier'], indent: true },
		{ href: '/products', label: 'Products', icon: Package, roles: ['admin', 'manager', 'cashier'] },
		{ href: '/inventory', label: 'Inventory', icon: Warehouse, roles: ['admin', 'manager'] },
		{ href: '/customers', label: 'Customers', icon: Users, roles: ['admin', 'manager', 'cashier'] },
		{ href: '/suppliers', label: 'Suppliers', icon: Truck, roles: ['admin', 'manager'] },
		{ href: '/purchases', label: 'Purchases', icon: ShoppingBag, roles: ['admin', 'manager'] },
		{ href: '/reports', label: 'Reports', icon: BarChart2, roles: ['admin', 'manager'] },
		{ href: '/users', label: 'Users', icon: UserCog, roles: ['admin'] },
		{ href: '/settings', label: 'Settings', icon: Settings, roles: ['admin', 'manager'] }
	];

	const role = $derived(authStore.role ?? 'cashier');

	function logout() {
		authStore.clear();
		goto('/login');
	}

	function nav(href: string) {
		goto(href);
		onclose?.();
	}
</script>

<aside class="flex h-full w-64 shrink-0 flex-col" style="background-color: #005A9C;">
	<!-- Logo -->
	<div class="flex items-center gap-3 px-5 py-5" style="border-bottom: 1px solid rgba(255,255,255,0.15);">
		<div class="flex h-9 w-9 items-center justify-center rounded-xl bg-white/20 shadow-lg">
			<Store size={18} class="text-white" />
		</div>
		<div class="flex-1 min-w-0">
			<p class="font-bold text-white text-sm leading-tight">POS System</p>
			<p class="text-xs text-blue-200 mt-0.5">Point of Sale</p>
		</div>
		<button onclick={onclose} class="text-white/60 hover:text-white transition-colors lg:hidden">
			<X size={18} />
		</button>
	</div>

	<!-- Navigation -->
	<nav class="flex-1 overflow-y-auto px-3 py-4">
		<ul class="space-y-0.5">
			{#each links as link}
				{#if link.roles.includes(role)}
					{@const active = $page.url.pathname === link.href}
					<li>
						<a
							href={link.href}
							onclick={() => onclose?.()}
							class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-all {link.indent ? 'ml-4' : ''} {active ? 'bg-white text-[#005A9C] font-semibold shadow-sm' : 'text-white/70 hover:bg-white/10 hover:text-white'}"
						>
							<link.icon size={15} />
							{link.label}
						</a>
					</li>
				{/if}
			{/each}
		</ul>
	</nav>

	<!-- User section -->
	<div class="p-3" style="border-top: 1px solid rgba(255,255,255,0.15);">
		{#if authStore.user}
			<div class="flex items-center gap-3 px-3 py-2.5 mb-1 rounded-lg bg-white/10">
				<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-white text-xs font-bold" style="color: #005A9C;">
					{authStore.user.name.charAt(0).toUpperCase()}
				</div>
				<div class="min-w-0 flex-1">
					<p class="text-sm font-medium text-white truncate">{authStore.user.name}</p>
					<p class="text-xs text-blue-200 capitalize">{authStore.user.role}</p>
				</div>
				<button
					onclick={themeStore.toggle}
					title={themeStore.dark ? 'Switch to light mode' : 'Switch to dark mode'}
					class="flex h-7 w-7 items-center justify-center rounded-lg text-white/60 hover:bg-white/10 hover:text-white transition-colors"
				>
					{#if themeStore.dark}
						<Sun size={14} />
					{:else}
						<Moon size={14} />
					{/if}
				</button>
			</div>
		{/if}
		<button
			onclick={logout}
			class="flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium text-white/60 hover:bg-white/10 hover:text-white transition-colors"
		>
			<LogOut size={15} />
			Sign out
		</button>
	</div>
</aside>
