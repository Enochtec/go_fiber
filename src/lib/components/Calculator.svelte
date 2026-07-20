<script lang="ts">
	let { onclose }: { onclose: () => void } = $props();

	let display = $state('0');
	let prev = $state<number | null>(null);
	let op = $state<string | null>(null);
	let reset = $state(false);

	function input(n: string) {
		if (reset) { display = n; reset = false; return; }
		if (display === '0' && n !== '.') { display = n; return; }
		if (n === '.' && display.includes('.')) return;
		display += n;
	}

	function setOp(next: string) {
		if (op && !reset) calc();
		prev = parseFloat(display);
		op = next;
		reset = true;
	}

	function calc() {
		if (prev === null || !op) return;
		const curr = parseFloat(display);
		let result: number;
		switch (op) {
			case '+': result = prev + curr; break;
			case '-': result = prev - curr; break;
			case '×': result = prev * curr; break;
			case '÷': result = curr === 0 ? 0 : prev / curr; break;
			default: return;
		}
		display = String(round(result));
		prev = null;
		op = null;
		reset = true;
	}

	function round(n: number) { return Math.round(n * 1e10) / 1e10; }

	function clear() { display = '0'; prev = null; op = null; reset = false; }

	function del() {
		if (display.length > 1) display = display.slice(0, -1);
		else display = '0';
	}

	function percent() {
		display = String(parseFloat(display) / 100);
	}

	function negate() {
		display = String(parseFloat(display) * -1);
	}
</script>

<div class="fixed inset-0 z-[9999] flex items-center justify-center bg-black/50 backdrop-blur-sm" onclick={() => onclose()}>
	<div
		class="w-72 rounded-[1px] bg-slate-900 shadow-2xl overflow-hidden"
		onclick={(e) => e.stopPropagation()}
		role="presentation"
	>
		<div class="px-5 pt-8 pb-4 text-right">
			<div class="text-3xl font-light tracking-wider text-white font-mono truncate">{display}</div>
		</div>

		<div class="grid grid-cols-4 gap-1.5 px-3 pb-4">
			<button onclick={clear} class="rounded-xl py-3 text-sm font-semibold bg-slate-200 text-slate-800 hover:bg-slate-300 active:scale-95 transition-all">C</button>
			<button onclick={del} class="rounded-xl py-3 text-sm font-semibold bg-slate-200 text-slate-800 hover:bg-slate-300 active:scale-95 transition-all">⌫</button>
			<button onclick={percent} class="rounded-xl py-3 text-sm font-semibold bg-slate-200 text-slate-800 hover:bg-slate-300 active:scale-95 transition-all">%</button>
			<button onclick={() => setOp('÷')} class="rounded-xl py-3 text-lg font-bold bg-orange-500 text-white hover:bg-orange-400 active:scale-95 transition-all">÷</button>

			<button onclick={() => input('7')} class="rounded-xl py-3 text-lg font-semibold bg-slate-700 text-white hover:bg-slate-600 active:scale-95 transition-all">7</button>
			<button onclick={() => input('8')} class="rounded-xl py-3 text-lg font-semibold bg-slate-700 text-white hover:bg-slate-600 active:scale-95 transition-all">8</button>
			<button onclick={() => input('9')} class="rounded-xl py-3 text-lg font-semibold bg-slate-700 text-white hover:bg-slate-600 active:scale-95 transition-all">9</button>
			<button onclick={() => setOp('×')} class="rounded-xl py-3 text-lg font-bold bg-orange-500 text-white hover:bg-orange-400 active:scale-95 transition-all">×</button>

			<button onclick={() => input('4')} class="rounded-xl py-3 text-lg font-semibold bg-slate-700 text-white hover:bg-slate-600 active:scale-95 transition-all">4</button>
			<button onclick={() => input('5')} class="rounded-xl py-3 text-lg font-semibold bg-slate-700 text-white hover:bg-slate-600 active:scale-95 transition-all">5</button>
			<button onclick={() => input('6')} class="rounded-xl py-3 text-lg font-semibold bg-slate-700 text-white hover:bg-slate-600 active:scale-95 transition-all">6</button>
			<button onclick={() => setOp('-')} class="rounded-xl py-3 text-lg font-bold bg-orange-500 text-white hover:bg-orange-400 active:scale-95 transition-all">-</button>

			<button onclick={() => input('1')} class="rounded-xl py-3 text-lg font-semibold bg-slate-700 text-white hover:bg-slate-600 active:scale-95 transition-all">1</button>
			<button onclick={() => input('2')} class="rounded-xl py-3 text-lg font-semibold bg-slate-700 text-white hover:bg-slate-600 active:scale-95 transition-all">2</button>
			<button onclick={() => input('3')} class="rounded-xl py-3 text-lg font-semibold bg-slate-700 text-white hover:bg-slate-600 active:scale-95 transition-all">3</button>
			<button onclick={() => setOp('+')} class="rounded-xl py-3 text-lg font-bold bg-orange-500 text-white hover:bg-orange-400 active:scale-95 transition-all">+</button>

			<button onclick={negate} class="rounded-xl py-3 text-lg font-semibold bg-slate-700 text-white hover:bg-slate-600 active:scale-95 transition-all">±</button>
			<button onclick={() => input('0')} class="rounded-xl py-3 text-lg font-semibold bg-slate-700 text-white hover:bg-slate-600 active:scale-95 transition-all">0</button>
			<button onclick={() => input('.')} class="rounded-xl py-3 text-lg font-semibold bg-slate-700 text-white hover:bg-slate-600 active:scale-95 transition-all">.</button>
			<button onclick={calc} class="rounded-xl py-3 text-lg font-bold bg-blue-600 text-white hover:bg-blue-500 active:scale-95 transition-all">=</button>
		</div>
	</div>
</div>