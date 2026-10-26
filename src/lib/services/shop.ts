import { api } from './api';
import { db } from './db';
import { enqueueSync } from './sync.svelte';
import type { Shop, ShopSettings } from '$lib/types';

export interface ShopInfo {
	shop: Shop;
	settings: ShopSettings | null;
}

export const shopService = {
	async getInfo(): Promise<ShopInfo> {
		try {
			const res = await api.get<{ success: boolean; data: ShopInfo }>('/shop/info');
			await db.meta.put({ key: 'shop_info', value: JSON.stringify(res.data) });
			return res.data;
		} catch {
			const cached = await db.meta.get('shop_info');
			if (cached) return JSON.parse(cached.value) as ShopInfo;
			throw new Error('Shop information is not available offline yet');
		}
	},
	async updateName(name: string): Promise<Shop> {
		try {
			const res = await api.put<{ success: boolean; data: Shop }>('/shop/name', { name });
			api.invalidate('/shop/info');
			return res.data;
		} catch {
			const cached = await db.meta.get('shop_info');
			if (!cached) throw new Error('Shop information is not available offline yet');
			const info = JSON.parse(cached.value) as ShopInfo;
			info.shop.name = name;
			await db.meta.put({ key: 'shop_info', value: JSON.stringify(info) });
			await enqueueSync('shop', 'update', info.shop.id, { name });
			return info.shop;
		}
	}
};
