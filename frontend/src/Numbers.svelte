<script>
	// export let ah = 0;
	export let n = 0;
	export let p = 0;
	export let prefix = "";
	export let sufix = "";

	function format(n, p, prefix, sufix) {
		if (!n) return "-";
		if (Math.abs(n) >= 1e6) {
			n = Math.round(n / 1e6);
			sufix = "M " + sufix;
		} else if (Math.abs(n) >= 1e3) {
			n = Math.round(n / 1e3);
			sufix = "K " + sufix;
		}
		n = new Intl.NumberFormat('pt-BR', { minimumFractionDigits: p, maximumFractionDigits: p }).format(n)
		return prefix + n + sufix;
	}
</script>

<style>
	/* .numbers {
        font-size: 0.8rem;
		font-weight: 400;
		min-width: 5rem;
		white-space: nowrap;
		display: flex;
		flex-flow: row nowrap;
		justify-content: space-between;
	}
	.n {
		padding: 0 0.3rem;
		min-width: 4rem;
        text-align: center;
	} */
	/* .ah,  */
	/* .av {
		font-size: 0.7em;
		vertical-align: middle;
		position: relative;
		align-self: center;
	} */
	/* .ah:before {
		content: '↕ ';
	} */
	/* .av:after {
		content: ' ↔';
	} */
	.red {
		color: red;
	}
</style>

<div class="numbers">
	<!-- <span class="ah {ah < 0 ? 'red' : ''}">{format(ah*100, 1)}%</span> -->
	<span class="n  {n < 0 ? 'red' : ''}">{format(n, p, prefix, sufix)}</span>
</div>
