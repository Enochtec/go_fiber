import type { User } from '$lib/types';

function createAuthStore() {
	let user = $state<User | null>(null);
	let token = $state<string | null>(null);

	if (typeof localStorage !== 'undefined') {
		token = localStorage.getItem('pos_token');
		const cachedUser = localStorage.getItem('pos_user');
		if (cachedUser) {
			try { user = JSON.parse(cachedUser) as User; } catch { localStorage.removeItem('pos_user'); }
		}
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
			localStorage.setItem('pos_user', JSON.stringify(newUser));
		},

		setUser(u: User) {
			user = u;
			localStorage.setItem('pos_user', JSON.stringify(u));
		},

		clear() {
			user = null;
			token = null;
			localStorage.removeItem('pos_token');
			localStorage.removeItem('pos_user');
		},

		can(roles: string[]) {
			return user ? roles.includes(user.role) : false;
		}
	};
}

export const authStore = createAuthStore();
