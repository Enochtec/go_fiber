import { api } from './api';
import type { ApiResponse, RegisterInput, RegisterResult, User } from '$lib/types';

export const authService = {
	login: (email: string, password: string) =>
		api.post<ApiResponse<{ user: User; token: string }>>('/auth/login', { email, password }),

	register: (input: RegisterInput) =>
		api.post<ApiResponse<RegisterResult>>('/auth/register', input),

	me: () => api.get<ApiResponse<User>>('/auth/me')
};
