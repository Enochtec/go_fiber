function createTheme() {
	let dark = $state(false);

	return {
		get dark() {
			return dark;
		},
		init() {
			if (typeof localStorage === 'undefined') return;
			const saved = localStorage.getItem('theme');
			dark = saved === 'dark';
			document.documentElement.classList.toggle('dark', dark);
		},
		toggle() {
			dark = !dark;
			document.documentElement.classList.toggle('dark', dark);
			localStorage.setItem('theme', dark ? 'dark' : 'light');
		}
	};
}

export const themeStore = createTheme();
