import type { User } from '$lib/types';

function createAuthStore() {
	let user = $state<User | null>(null);
	let token = $state<string | null>(null);

	if (typeof localStorage !== 'undefined') {
		token = localStorage.getItem('pos_token');
	}

	return {
		get user() { return user; },
		get token() { return token; },
		get isAuthenticated() { return !!token; },
		get role() { return user?.role ?? null; },

		set(newUser: User, newToken: string) {
			user = newUser;
			token = newToken;
			localStorage.setItem('pos_token', newToken);
		},

		setUser(u: User) {
			user = u;
		},

		clear() {
			user = null;
			token = null;
			localStorage.removeItem('pos_token');
		},

		can(roles: string[]) {
			return user ? roles.includes(user.role) : false;
		}
	};
}

export const authStore = createAuthStore();
