function createInvalidationStore() {
	let productVersion = $state(0);

	return {
		get productVersion() { return productVersion; },
		invalidateProducts() { productVersion++; }
	};
}

export const invalidation = createInvalidationStore();
