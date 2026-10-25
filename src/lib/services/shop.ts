import { api } from './api';
import type { Shop, ShopSettings } from '$lib/types';

export interface ShopInfo {
	shop: Shop;
	settings: ShopSettings | null;
}

export const shopService = {
	async getInfo(): Promise<ShopInfo> {
		const res = await api.get<{ success: boolean; data: ShopInfo }>('/shop/info');
		return res.data;
	},
	async updateName(name: string): Promise<Shop> {
		const res = await api.put<{ success: boolean; data: Shop }>('/shop/name', { name });
		api.invalidate('/shop/info');
		return res.data;
	}
};
