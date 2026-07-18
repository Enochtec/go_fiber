<script lang="ts">
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/stores/auth.svelte';
	import { authService } from '$lib/services/auth';
	import { notify } from '$lib/stores/notification.svelte';
	import Notification from '$lib/components/Notification.svelte';
	import { Store, Check, ChevronRight, ChevronLeft, Building2, User, ShieldCheck, Eye, EyeOff, LoaderCircle } from '@lucide/svelte';

	type Step = 1 | 2 | 3;

	const BUSINESS_TYPES = [
		'Retail Shop', 'Supermarket', 'Pharmacy', 'Hardware',
		'Restaurant', 'Boutique', 'Electronics', 'Other'
	] as const;

	const CURRENCIES = [
		{ code: 'KES', label: 'KES — Kenyan Shilling' },
		{ code: 'UGX', label: 'UGX — Ugandan Shilling' },
		{ code: 'TZS', label: 'TZS — Tanzanian Shilling' },
		{ code: 'RWF', label: 'RWF — Rwandan Franc' },
		{ code: 'USD', label: 'USD — US Dollar' },
		{ code: 'GBP', label: 'GBP — British Pound' },
		{ code: 'EUR', label: 'EUR — Euro' },
		{ code: 'NGN', label: 'NGN — Nigerian Naira' },
		{ code: 'ZAR', label: 'ZAR — South African Rand' },
		{ code: 'GHS', label: 'GHS — Ghanaian Cedi' },
	] as const;

	const TIMEZONES = [
		'Africa/Nairobi', 'Africa/Dar_es_Salaam', 'Africa/Kampala',
		'Africa/Kigali', 'Africa/Lagos', 'Africa/Accra',
		'Africa/Johannesburg', 'Africa/Cairo', 'Africa/Casablanca',
		'Europe/London', 'America/New_York', 'America/Chicago',
		'America/Denver', 'America/Los_Angeles', 'Asia/Dubai',
	] as const;

	const COUNTRIES = [
		'Kenya', 'Uganda', 'Tanzania', 'Rwanda', 'Nigeria',
		'Ghana', 'South Africa', 'United States', 'United Kingdom',
		'Other'
	] as const;

	let step = $state<Step>(1);
	let loading = $state(false);
	let showPassword = $state(false);
	let showConfirm = $state(false);

	let form = $state({
		shop_name: '', business_type: 'Retail Shop', business_email: '', business_phone: '',
		country: 'Kenya', county: '', town: '', address: '', currency: 'KES', timezone: 'Africa/Nairobi',
		owner_name: '', owner_email: '', owner_phone: '',
		username: '', password: '', confirm_password: ''
	});

	let errors = $state<Record<string, string>>({});
	let touched = $state<Record<string, boolean>>({});

	function touch(field: string) {
		touched[field] = true;
	}

	function getError(field: string): string | undefined {
		return errors[field];
	}

	const businessFields = ['shop_name', 'business_type', 'country', 'county', 'town'] as const;
	const ownerFields = ['owner_name', 'owner_email', 'owner_phone'] as const;
	const accountFields = ['username', 'password', 'confirm_password'] as const;

	$effect(() => {
		const e: Record<string, string> = {};

		if (touched.shop_name && !form.shop_name.trim()) e.shop_name = 'Shop name is required';
		if (touched.business_email && form.business_email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.business_email)) e.business_email = 'Invalid email';
		if (touched.country && !form.country) e.country = 'Country is required';
		if (touched.county && !form.county.trim()) e.county = 'County is required';
		if (touched.town && !form.town.trim()) e.town = 'Town is required';

		if (touched.owner_name && !form.owner_name.trim()) e.owner_name = 'Full name is required';
		if (touched.owner_email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.owner_email)) e.owner_email = 'Valid email is required';
		if (touched.owner_phone && !form.owner_phone.trim()) e.owner_phone = 'Phone number is required';

		if (touched.username && !form.username.trim()) e.username = 'Username is required';
		if (touched.username && form.username.trim().length < 3) e.username = 'At least 3 characters';
		if (touched.password && !form.password) e.password = 'Password is required';
		if (touched.password && form.password.length < 8) e.password = 'At least 8 characters';
		if (touched.confirm_password && form.password !== form.confirm_password) e.confirm_password = 'Passwords do not match';

		errors = e;
	});

	function canAdvance(stepNum: Step): boolean {
		if (stepNum === 1) {
			return !!form.shop_name.trim() && !!form.country && !!form.county.trim() && !!form.town.trim();
		}
		if (stepNum === 2) {
			return !!form.owner_name.trim() && !!form.owner_email.trim() && !!form.owner_phone.trim()
				&& /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.owner_email);
		}
		if (stepNum === 3) {
			return !!form.username.trim() && form.username.trim().length >= 3
				&& !!form.password && form.password.length >= 8
				&& form.password === form.confirm_password;
		}
		return false;
	}

	const passStrength = $derived.by(() => {
		const p = form.password;
		if (!p) return { label: '', color: 'bg-slate-200', width: '0%', text: '' };
		let score = 0;
		if (p.length >= 8) score += 25;
		if (p.length >= 12) score += 10;
		if (/[a-z]/.test(p)) score += 15;
		if (/[A-Z]/.test(p)) score += 15;
		if (/[0-9]/.test(p)) score += 15;
		if (/[^a-zA-Z0-9]/.test(p)) score += 20;
		if (score < 30) return { label: 'Weak', color: 'bg-red-500', width: `${score}%`, text: 'text-red-600' };
		if (score < 60) return { label: 'Fair', color: 'bg-orange-500', width: `${score}%`, text: 'text-orange-600' };
		if (score < 80) return { label: 'Good', color: 'bg-yellow-500', width: `${score}%`, text: 'text-yellow-600' };
		return { label: 'Strong', color: 'bg-emerald-500', width: '100%', text: 'text-emerald-600' };
	});

	function nextStep() {
		touchAll(step);
		if (!canAdvance(step)) return;
		if (step < 3) step = (step + 1) as Step;
	}

	function prevStep() {
		if (step > 1) step = (step - 1) as Step;
	}

	function touchAll(s: Step) {
		const fields = s === 1 ? businessFields : s === 2 ? ownerFields : accountFields;
		for (const f of fields) touch(f);
	}

	async function submit() {
		touchAll(3);
		if (!canAdvance(3)) return;
		loading = true;
		try {
			const input = {
				shop_name: form.shop_name.trim(),
				business_type: form.business_type,
				business_email: form.business_email.trim(),
				business_phone: form.business_phone.trim(),
				country: form.country,
				county: form.county.trim(),
				town: form.town.trim(),
				address: form.address.trim(),
				currency: form.currency,
				timezone: form.timezone,
				owner_name: form.owner_name.trim(),
				owner_email: form.owner_email.trim(),
				owner_phone: form.owner_phone.trim(),
				username: form.username.trim(),
				password: form.password,
				confirm_password: form.confirm_password
			};
			const res = await authService.register(input);
			if (res.data) {
				authStore.set(res.data.user, res.data.token);
				goto('/register/success', { state: { shop: res.data.shop, user: res.data.user } });
			}
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Registration failed');
		} finally {
			loading = false;
		}
	}

	const steps = [
		{ num: 1, label: 'Business', icon: Building2 },
		{ num: 2, label: 'Owner', icon: User },
		{ num: 3, label: 'Account', icon: ShieldCheck },
	];
</script>

<svelte:head><title>Register — Maestro POS</title></svelte:head>
<Notification />

<div class="flex min-h-screen bg-slate-50">
	<!-- Left branding -->
	<div class="hidden lg:flex lg:w-5/12 xl:w-2/5 flex-col bg-slate-900 p-10">
		<div class="flex items-center gap-3 mb-auto">
			<div class="flex h-9 w-9 items-center justify-center rounded-lg bg-blue-600">
				<Store size={18} class="text-white" />
			</div>
			<span class="text-white font-bold text-base">Maestro POS</span>
		</div>
		<div class="py-12">
			<h1 class="text-3xl font-bold text-white leading-snug mb-3">
				Get started<br />in minutes.
			</h1>
			<p class="text-slate-400 text-sm leading-relaxed mb-10">
				Create your shop and start selling. No credit card required.
			</p>
			<div class="space-y-3">
				{#each [
					['Step 1: Shop Details', 'Tell us about your business'],
					['Step 2: Your Info', 'Who will manage the system'],
					['Step 3: Secure Access', 'Create your login credentials'],
				] as [title, desc]}
					<div class="flex items-start gap-3 p-3 rounded-lg bg-slate-800 border border-slate-700">
						<div class="h-1.5 w-1.5 rounded-full bg-blue-500 mt-2 shrink-0"></div>
						<div>
							<p class="text-sm font-semibold text-slate-200">{title}</p>
							<p class="text-xs text-slate-500 mt-0.5">{desc}</p>
						</div>
					</div>
				{/each}
			</div>
		</div>
		<p class="text-xs text-slate-600">© {new Date().getFullYear()} Maestro POS</p>
	</div>

	<!-- Right form panel -->
	<div class="flex-1 flex flex-col items-center justify-center p-4 md:p-8">
		<div class="w-full max-w-lg">
			<!-- Mobile logo -->
			<div class="mb-6 text-center lg:hidden">
				<div class="mx-auto mb-3 flex h-10 w-10 items-center justify-center rounded-xl bg-blue-600">
					<Store size={20} class="text-white" />
				</div>
				<h1 class="text-lg font-bold text-slate-900">Create Your Account</h1>
				<p class="text-sm text-slate-500 mt-0.5">Set up your shop in 3 simple steps</p>
			</div>

			<!-- Progress indicator -->
			<div class="mb-8">
				<div class="flex items-center justify-between">
					{#each steps as s, i}
						<button onclick={() => { if (s.num < step) step = s.num as Step; }} class="flex flex-col items-center gap-1.5 group" type="button">
							<div class="flex h-9 w-9 items-center justify-center rounded-full text-xs font-bold transition-all duration-300 {s.num < step ? 'bg-blue-600 text-white' : s.num === step ? 'bg-blue-600 text-white ring-4 ring-blue-200' : 'bg-slate-200 text-slate-400'}">
								{#if s.num < step}
									<Check size={15} />
								{:else}
									{s.num}
								{/if}
							</div>
							<span class="text-[10px] font-semibold uppercase tracking-wide {s.num <= step ? 'text-blue-600' : 'text-slate-400'}">{s.label}</span>
						</button>
						{#if i < steps.length - 1}
							<div class="flex-1 h-px mx-2 {s.num < step ? 'bg-blue-500' : 'bg-slate-200'} transition-colors"></div>
						{/if}
					{/each}
				</div>
				<p class="text-center text-xs text-slate-400 mt-2">Step {step} of 3</p>
			</div>

			<!-- Step 1: Business Information -->
			{#if step === 1}
				<div class="space-y-4 transition-opacity duration-300">
					<div class="mb-4">
						<h2 class="text-lg font-bold text-slate-900">Business Information</h2>
						<p class="text-sm text-slate-500">Tell us about your shop</p>
					</div>

					<div>
						<label for="shop_name" class="block text-xs font-semibold text-slate-600 uppercase tracking-wide mb-1.5">Shop Name <span class="text-red-400">*</span></label>
						<input id="shop_name" type="text" bind:value={form.shop_name} oninput={() => touch('shop_name')} required autocomplete="organization"
							class="w-full rounded-lg border px-3.5 py-2.5 text-sm outline-none transition {getError('shop_name') ? 'border-red-300 bg-red-50 focus:border-red-500 focus:ring-red-500/15' : 'border-slate-200 bg-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/15'}"
							placeholder="e.g. My Corner Shop" />
						{#if getError('shop_name')}<p class="text-xs text-red-500 mt-1">{getError('shop_name')}</p>{/if}
					</div>

					<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
						<div>
							<label for="business_type" class="block text-xs font-semibold text-slate-600 uppercase tracking-wide mb-1.5">Business Type</label>
							<select id="business_type" bind:value={form.business_type}
								class="w-full rounded-lg border border-slate-200 bg-white px-3.5 py-2.5 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-500/15">
								{#each BUSINESS_TYPES as t}
									<option value={t}>{t}</option>
								{/each}
							</select>
						</div>
						<div>
							<label for="country" class="block text-xs font-semibold text-slate-600 uppercase tracking-wide mb-1.5">Country <span class="text-red-400">*</span></label>
							<select id="country" bind:value={form.country} onchange={() => touch('country')}
								class="w-full rounded-lg border px-3.5 py-2.5 text-sm outline-none transition {getError('country') ? 'border-red-300 bg-red-50 focus:border-red-500 focus:ring-red-500/15' : 'border-slate-200 bg-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/15'}">
								{#each COUNTRIES as c}
									<option value={c}>{c}</option>
								{/each}
							</select>
							{#if getError('country')}<p class="text-xs text-red-500 mt-1">{getError('country')}</p>{/if}
						</div>
					</div>

					<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
						<div>
							<label for="county" class="block text-xs font-semibold text-slate-600 uppercase tracking-wide mb-1.5">County / State <span class="text-red-400">*</span></label>
							<input id="county" type="text" bind:value={form.county} oninput={() => touch('county')} required
								class="w-full rounded-lg border px-3.5 py-2.5 text-sm outline-none transition {getError('county') ? 'border-red-300 bg-red-50 focus:border-red-500 focus:ring-red-500/15' : 'border-slate-200 bg-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/15'}"
								placeholder="e.g. Nairobi" />
							{#if getError('county')}<p class="text-xs text-red-500 mt-1">{getError('county')}</p>{/if}
						</div>
						<div>
							<label for="town" class="block text-xs font-semibold text-slate-600 uppercase tracking-wide mb-1.5">Town / City <span class="text-red-400">*</span></label>
							<input id="town" type="text" bind:value={form.town} oninput={() => touch('town')} required
								class="w-full rounded-lg border px-3.5 py-2.5 text-sm outline-none transition {getError('town') ? 'border-red-300 bg-red-50 focus:border-red-500 focus:ring-red-500/15' : 'border-slate-200 bg-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/15'}"
								placeholder="e.g. Westlands" />
							{#if getError('town')}<p class="text-xs text-red-500 mt-1">{getError('town')}</p>{/if}
						</div>
					</div>

					<div>
						<label for="address" class="block text-xs font-semibold text-slate-600 uppercase tracking-wide mb-1.5">Business Address</label>
						<input id="address" type="text" bind:value={form.address}
							class="w-full rounded-lg border border-slate-200 bg-white px-3.5 py-2.5 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-500/15"
							placeholder="Street, building, landmark" />
					</div>

					<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
						<div>
							<label for="business_email" class="block text-xs font-semibold text-slate-600 uppercase tracking-wide mb-1.5">Business Email</label>
							<input id="business_email" type="email" bind:value={form.business_email} oninput={() => touch('business_email')}
								class="w-full rounded-lg border px-3.5 py-2.5 text-sm outline-none transition {getError('business_email') ? 'border-red-300 bg-red-50 focus:border-red-500 focus:ring-red-500/15' : 'border-slate-200 bg-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/15'}"
								placeholder="shop@example.com" />
							{#if getError('business_email')}<p class="text-xs text-red-500 mt-1">{getError('business_email')}</p>{/if}
						</div>
						<div>
							<label for="business_phone" class="block text-xs font-semibold text-slate-600 uppercase tracking-wide mb-1.5">Business Phone</label>
							<input id="business_phone" type="tel" bind:value={form.business_phone}
								class="w-full rounded-lg border border-slate-200 bg-white px-3.5 py-2.5 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-500/15"
								placeholder="+254 712 345 678" />
						</div>
					</div>

					<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
						<div>
							<label for="currency" class="block text-xs font-semibold text-slate-600 uppercase tracking-wide mb-1.5">Currency</label>
							<select id="currency" bind:value={form.currency}
								class="w-full rounded-lg border border-slate-200 bg-white px-3.5 py-2.5 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-500/15">
								{#each CURRENCIES as c}
									<option value={c.code}>{c.label}</option>
								{/each}
							</select>
						</div>
						<div>
							<label for="timezone" class="block text-xs font-semibold text-slate-600 uppercase tracking-wide mb-1.5">Timezone</label>
							<select id="timezone" bind:value={form.timezone}
								class="w-full rounded-lg border border-slate-200 bg-white px-3.5 py-2.5 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-500/15">
								{#each TIMEZONES as tz}
									<option value={tz}>{tz}</option>
								{/each}
							</select>
						</div>
					</div>
				</div>
			{/if}

			<!-- Step 2: Owner Information -->
			{#if step === 2}
				<div class="space-y-4 transition-opacity duration-300">
					<div class="mb-4">
						<h2 class="text-lg font-bold text-slate-900">Owner Information</h2>
						<p class="text-sm text-slate-500">Who will manage this account</p>
					</div>

					<div>
						<label for="owner_name" class="block text-xs font-semibold text-slate-600 uppercase tracking-wide mb-1.5">Full Name <span class="text-red-400">*</span></label>
						<input id="owner_name" type="text" bind:value={form.owner_name} oninput={() => touch('owner_name')} required autocomplete="name"
							class="w-full rounded-lg border px-3.5 py-2.5 text-sm outline-none transition {getError('owner_name') ? 'border-red-300 bg-red-50 focus:border-red-500 focus:ring-red-500/15' : 'border-slate-200 bg-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/15'}"
							placeholder="e.g. Jane Doe" />
						{#if getError('owner_name')}<p class="text-xs text-red-500 mt-1">{getError('owner_name')}</p>{/if}
					</div>

					<div>
						<label for="owner_email" class="block text-xs font-semibold text-slate-600 uppercase tracking-wide mb-1.5">Email Address <span class="text-red-400">*</span></label>
						<input id="owner_email" type="email" bind:value={form.owner_email} oninput={() => touch('owner_email')} required autocomplete="email"
							class="w-full rounded-lg border px-3.5 py-2.5 text-sm outline-none transition {getError('owner_email') ? 'border-red-300 bg-red-50 focus:border-red-500 focus:ring-red-500/15' : 'border-slate-200 bg-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/15'}"
							placeholder="you@example.com" />
						{#if getError('owner_email')}<p class="text-xs text-red-500 mt-1">{getError('owner_email')}</p>{/if}
					</div>

					<div>
						<label for="owner_phone" class="block text-xs font-semibold text-slate-600 uppercase tracking-wide mb-1.5">Phone Number <span class="text-red-400">*</span></label>
						<input id="owner_phone" type="tel" bind:value={form.owner_phone} oninput={() => touch('owner_phone')} required autocomplete="tel"
							class="w-full rounded-lg border px-3.5 py-2.5 text-sm outline-none transition {getError('owner_phone') ? 'border-red-300 bg-red-50 focus:border-red-500 focus:ring-red-500/15' : 'border-slate-200 bg-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/15'}"
							placeholder="+254 712 345 678" />
						{#if getError('owner_phone')}<p class="text-xs text-red-500 mt-1">{getError('owner_phone')}</p>{/if}
					</div>
				</div>
			{/if}

			<!-- Step 3: Account Information -->
			{#if step === 3}
				<div class="space-y-4 transition-opacity duration-300">
					<div class="mb-4">
						<h2 class="text-lg font-bold text-slate-900">Account Security</h2>
						<p class="text-sm text-slate-500">Create your login credentials</p>
					</div>

					<div>
						<label for="username" class="block text-xs font-semibold text-slate-600 uppercase tracking-wide mb-1.5">Username <span class="text-red-400">*</span></label>
						<input id="username" type="text" bind:value={form.username} oninput={() => touch('username')} required autocomplete="username"
							class="w-full rounded-lg border px-3.5 py-2.5 text-sm outline-none transition {getError('username') ? 'border-red-300 bg-red-50 focus:border-red-500 focus:ring-red-500/15' : 'border-slate-200 bg-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/15'}"
							placeholder="Choose a unique username" />
						{#if getError('username')}<p class="text-xs text-red-500 mt-1">{getError('username')}</p>{/if}
					</div>

					<div>
						<label for="password" class="block text-xs font-semibold text-slate-600 uppercase tracking-wide mb-1.5">Password <span class="text-red-400">*</span></label>
						<div class="relative">
							<input id="password" type={showPassword ? 'text' : 'password'} bind:value={form.password} oninput={() => { touch('password'); touch('confirm_password'); }} required autocomplete="new-password"
								class="w-full rounded-lg border px-3.5 py-2.5 pr-10 text-sm outline-none transition {getError('password') ? 'border-red-300 bg-red-50 focus:border-red-500 focus:ring-red-500/15' : 'border-slate-200 bg-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/15'}"
								placeholder="Min. 8 characters" />
							<button type="button" onclick={() => showPassword = !showPassword} class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600">
								{#if showPassword}<EyeOff size={16} />{:else}<Eye size={16} />{/if}
							</button>
						</div>
						{#if form.password}
							<div class="mt-2">
								<div class="h-1.5 rounded-full bg-slate-200 overflow-hidden">
									<div class="h-full rounded-full transition-all duration-300 {passStrength.color}" style="width:{passStrength.width}"></div>
								</div>
								<p class="text-xs mt-0.5 {passStrength.text}">{passStrength.label}</p>
							</div>
						{/if}
						{#if getError('password')}<p class="text-xs text-red-500 mt-1">{getError('password')}</p>{/if}
					</div>

					<div>
						<label for="confirm_password" class="block text-xs font-semibold text-slate-600 uppercase tracking-wide mb-1.5">Confirm Password <span class="text-red-400">*</span></label>
						<div class="relative">
							<input id="confirm_password" type={showConfirm ? 'text' : 'password'} bind:value={form.confirm_password} oninput={() => touch('confirm_password')} required autocomplete="new-password"
								class="w-full rounded-lg border px-3.5 py-2.5 pr-10 text-sm outline-none transition {getError('confirm_password') ? 'border-red-300 bg-red-50 focus:border-red-500 focus:ring-red-500/15' : 'border-slate-200 bg-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/15'}"
								placeholder="Repeat your password" />
							<button type="button" onclick={() => showConfirm = !showConfirm} class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600">
								{#if showConfirm}<EyeOff size={16} />{:else}<Eye size={16} />{/if}
							</button>
						</div>
						{#if form.confirm_password && form.password !== form.confirm_password}
							<p class="text-xs text-red-500 mt-1">Passwords do not match</p>
						{/if}
						{#if getError('confirm_password')}<p class="text-xs text-red-500 mt-1">{getError('confirm_password')}</p>{/if}
					</div>
				</div>
			{/if}

			<!-- Navigation buttons -->
			<div class="flex items-center justify-between mt-8 pt-6 border-t border-slate-200">
				{#if step > 1}
					<button onclick={prevStep} type="button"
						class="flex items-center gap-1.5 rounded-lg border border-slate-200 bg-white px-4 py-2.5 text-sm font-medium text-slate-600 hover:bg-slate-50 transition-colors">
						<ChevronLeft size={15} /> Back
					</button>
				{:else}
					<div></div>
				{/if}

				{#if step < 3}
					<button onclick={nextStep} type="button"
						class="flex items-center gap-1.5 rounded-lg bg-blue-600 px-5 py-2.5 text-sm font-semibold text-white hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors active:scale-[0.98]"
						disabled={!canAdvance(step)}>
						Continue <ChevronRight size={15} />
					</button>
				{:else}
					<button onclick={submit} disabled={loading || !canAdvance(3)}
						class="flex items-center gap-2 rounded-lg bg-blue-600 px-5 py-2.5 text-sm font-semibold text-white hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors active:scale-[0.98]">
						{#if loading}
							<LoaderCircle size={16} class="animate-spin" /> Creating Account…
						{:else}
							<Check size={16} /> Create Account
						{/if}
					</button>
				{/if}
			</div>

			<p class="mt-6 text-center text-xs text-slate-400">
				Already have an account?
				<a href="/login" class="text-blue-600 font-semibold hover:text-blue-700">Sign in</a>
			</p>
		</div>
	</div>
</div>
