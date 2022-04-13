<script>
	// export let ah = 0;
	export let n = 0;
	export let av = 0;
	
	function format(n, p) {
		return Number(parseFloat(n).toFixed(p)).toLocaleString('pt-BR');
	}
</script>

<style>
	.numbers {
        font-size: 1rem;
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
	}
	/* .ah,  */
	.av {
		font-size: 0.7em;
		vertical-align: middle;
		position: relative;
		align-self: center;
	}
	/* .ah:before {
		content: '↕ ';
	} */
	.av:after {
		content: ' ↔';
	}
	.red {
		color: red;
	}
</style>

<div class="numbers">
	<!-- <span class="ah {ah < 0 ? 'red' : ''}">{format(ah*100, 1)}%</span> -->
	<span class="n  {n  < 0 ? 'red' : ''}">{n}</span>
	<span class="av {av < 0 ? 'red' : ''}">{format(av*100, 1)}%</span>
</div>
