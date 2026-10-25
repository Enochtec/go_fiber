const DEFAULT_SHOP_NAME = 'One Stop Point of Sale';
let name = $state(DEFAULT_SHOP_NAME);

export const shopStore = {
	get name() {
		return name;
	},
	setName(value: string) {
		name = value.trim() || DEFAULT_SHOP_NAME;
	}
};