<script lang="ts">
	import { page } from '$app/stores';
	import { authStore } from '$lib/stores/auth.svelte';
	import {
		LayoutDashboard, ShoppingCart, Package, Warehouse,
		Users, Truck, ShoppingBag, BarChart2, UserCog, Settings,
		LogOut, Store, X, Sun, Moon, History
	} from '@lucide/svelte';
	import { goto } from '$app/navigation';
	import { themeStore } from '$lib/stores/theme.svelte';

	interface Props { onclose?: () => void; }
	let { onclose }: Props = $props();

	const groups = [
		{
			label: 'Main',
			links: [
				{ href: '/dashboard', label: 'Dashboard', icon: LayoutDashboard, roles: ['admin', 'manager', 'cashier'] },
				{ href: '/sales',     label: 'Point of Sale', icon: ShoppingCart, roles: ['admin', 'manager', 'cashier'] },
				{ href: '/sales/history', label: 'Sales History', icon: History, roles: ['admin', 'manager', 'cashier'] },
			]
		},
		{
			label: 'Catalogue',
			links: [
				{ href: '/products',  label: 'Products',   icon: Package,   roles: ['admin', 'manager', 'cashier'] },
				{ href: '/inventory', label: 'Inventory',  icon: Warehouse,  roles: ['admin', 'manager'] },
				{ href: '/customers', label: 'Customers',  icon: Users,      roles: ['admin', 'manager', 'cashier'] },
			]
		},
		{
			label: 'Procurement',
			links: [
				{ href: '/suppliers', label: 'Suppliers',  icon: Truck,      roles: ['admin', 'manager'] },
				{ href: '/purchases', label: 'Purchases',  icon: ShoppingBag, roles: ['admin', 'manager'] },
			]
		},
		{
			label: 'Analytics',
			links: [
				{ href: '/reports',   label: 'Reports',    icon: BarChart2,  roles: ['admin', 'manager'] },
			]
		},
		{
			label: 'Admin',
			links: [
				{ href: '/users',     label: 'Users',      icon: UserCog,    roles: ['admin'] },
				{ href: '/settings',  label: 'Settings',   icon: Settings,   roles: ['admin', 'manager'] },
			]
		},
	];

	const role = $derived(authStore.role ?? 'cashier');
	const path = $derived($page.url.pathname);

	function logout() { authStore.clear(); goto('/login'); }
</script>

<aside class="flex h-full w-64 shrink-0 flex-col bg-slate-900 select-none">
	<!-- Logo -->
	<div class="flex items-center gap-3 px-4 py-4 border-b border-slate-800">
		<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-blue-600">
			<Store size={16} class="text-white" />
		</div>
		<div class="flex-1 min-w-0">
			<p class="text-sm font-bold text-white leading-none">Maestro POS</p>
			<p class="text-xs text-slate-500 mt-0.5">Point of Sale</p>
		</div>
		<button onclick={onclose} class="h-7 w-7 flex items-center justify-center rounded-md text-slate-500 hover:text-slate-300 hover:bg-slate-800 transition-colors lg:hidden">
			<X size={15} />
		</button>
	</div>

	<!-- Navigation -->
	<nav class="flex-1 overflow-y-auto py-3">
		{#each groups as group}
			{@const visible = group.links.filter(l => l.roles.includes(role))}
			{#if visible.length > 0}
				<div class="mb-1">
					<p class="px-4 py-1.5 text-[10px] font-semibold uppercase tracking-widest text-slate-500">{group.label}</p>
					{#each visible as link}
						{@const active = path === link.href || (link.href !== '/sales' && path.startsWith(link.href + '/') && link.href !== '/dashboard')}
						<a
							href={link.href}
							data-sveltekit-preload-code
							onclick={() => onclose?.()}
							class="flex items-center gap-2.5 mx-2 px-3 py-2 rounded-md text-sm font-medium transition-colors
								{active
									? 'bg-blue-600 text-white'
									: 'text-slate-400 hover:text-slate-200 hover:bg-slate-800'}"
						>
							<link.icon size={15} class="shrink-0" />
							<span>{link.label}</span>
						</a>
					{/each}
				</div>
			{/if}
		{/each}
	</nav>

	<!-- User section -->
	<div class="border-t border-slate-800 p-3 space-y-1">
		{#if authStore.user}
			<div class="flex items-center gap-2.5 px-3 py-2 rounded-md bg-slate-800">
				<div class="h-7 w-7 shrink-0 flex items-center justify-center rounded-full bg-blue-600 text-xs font-bold text-white">
					{authStore.user.name.charAt(0).toUpperCase()}
				</div>
				<div class="flex-1 min-w-0">
					<p class="text-xs font-semibold text-slate-200 truncate">{authStore.user.name}</p>
					<p class="text-[10px] text-slate-500 capitalize">{authStore.user.role}</p>
				</div>
				<button
					onclick={themeStore.toggle}
					title={themeStore.dark ? 'Light mode' : 'Dark mode'}
					class="h-6 w-6 flex items-center justify-center rounded text-slate-500 hover:text-slate-300 transition-colors"
				>
					{#if themeStore.dark}<Sun size={13} />{:else}<Moon size={13} />{/if}
				</button>
			</div>
		{/if}
		<button
			onclick={logout}
			class="flex w-full items-center gap-2.5 px-3 py-2 rounded-md text-sm text-slate-500 hover:text-slate-200 hover:bg-slate-800 transition-colors"
		>
			<LogOut size={14} />
			Sign out
		</button>
	</div>
</aside>
