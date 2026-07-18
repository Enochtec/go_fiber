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
	}
};
