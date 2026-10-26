const DEFAULT_SHOP_NAME = 'One Point POS';
let name = $state(DEFAULT_SHOP_NAME);

export const shopStore = {
	get name() {
		return name;
	},
	setName(value: string) {
		name = value.trim() || DEFAULT_SHOP_NAME;
	}
};