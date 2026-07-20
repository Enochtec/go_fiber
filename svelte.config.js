import adapter from '@sveltejs/adapter-static';

const config = {
	compilerOptions: {
		runes: true
	},
	kit: {
		adapter: adapter({ fallback: 'index.html' })
	}
};

export default config;
