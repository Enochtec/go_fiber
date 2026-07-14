import { api } from './api';
import type { ApiResponse, User } from '$lib/types';

export const authService = {
	login: (email: string, password: string) =>
		api.post<ApiResponse<{ user: User; token: string }>>('/auth/login', { email, password }),

	me: () => api.get<ApiResponse<User>>('/auth/me')
};
