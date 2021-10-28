<!--
SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>

SPDX-License-Identifier: MIT
-->
<script>
	import { onMount } from "svelte";
	import { apiDFP } from "./dfp";

	// type Conta = {
	// 	codigo: string;
	// 	descr: string;
	// 	totais: string[];
	// };

	// type DFP = {
	// 	nome: string;
	// 	cnpj: string;
	// 	anos: number[];
	// 	contas: Conta[];
	// };

	let dfp;
	let err = "";
	onMount(async () => {
		[dfp, err] = await apiDFP("84.429.695%2F0001-11", "2020");
		console.log("onMount()", dfp, err);
	});

	function format(n) {
		if (!n) return "-";
		return (n / 1000).toLocaleString("pt-BR");
	}

	function fontWeight(cod) {
		switch (lvl(cod)) {
			case 0:
				return 900;
			case 1:
				return 700;
		}
		return 400;
	}

	function lvl(cod) {
		return cod.split(".").length - 1;
	}

	function tag(cod) {
		const c = cod.split(".");

		if (c.length <= 1) return "";

		let res = "";
		const max = c.length < 3 ? c.length : 3;
		for (let i = 1; i < max; i++) {
			const j = c.slice(0, i);
			res += "lvl-" + j.join("-") + " ";
		}
		return res.trim();
	}

	const toggled = {};
	function toggle(cod) {
		const c = cod.split(".");
		const j = ".lvl-" + c.join("-");
		const elements = document.querySelectorAll(j);
		for (let i = 0; i < elements.length; i++) {
			elements[i].style.display = toggled[cod] ? "" : "none";
		}
		toggled[cod] = !toggled[cod];
	}

	// https://svelte.dev/repl/69efbdcbbb6743e9988f777ef0f906ed?version=3.44.0
	// background = linear-gradient(to top, #d7e7d7 40%, #f8fcf8 40%)
	// t.style.background = "linear-gradient(to right,"+col1+" "+percentage+"%, "+col2+" "+percentage+"%)";
</script>

<div class="container">
	<div class="navbar">
		<a href="/" class="navbar-title">Rapina</a>
		<div class="navbar-nav">
			<a href="fii.html">[FII:Rendimentos]</a>
			<a href="financials.html">[Ações:Finanças]</a>
		</div>
	</div>
</div>
<hr />

<div id="content">
	<div class="container">
		{#if err != ''}
			<p>Erro: {err}</p>
		{/if}
		{#if err == '' && dfp && dfp.cnpj != '' && dfp.contas}
			<p>CNPJ: {dfp.cnpj}</p>
			<p>Nome: {dfp.nome}</p>
			<small>
				<table>
					<tr>
						<th>Cód.</th>
						<th>Descr.</th>
						{#each dfp.anos as ano}
							<th style="text-align:center">{ano}</th>
						{/each}
					</tr>

					{#each dfp.contas as conta}
						<tr
							class={tag(conta.codigo)}
							style="font-weight: {fontWeight(conta.codigo)}"
							on:click={() => toggle(conta.codigo)}>
							<td>{conta.codigo}</td>
							<td>{conta.descr}</td>
							{#each conta.totais as total}
								<td style="text-align:right;">{format(total)}</td>
							{/each}
						</tr>
					{/each}
				</table>
			</small>
		{/if}
	</div>
</div>

<footer>
	<hr />
	<div class="icon baseline">
		<svg
			xmlns="http://www.w3.org/2000/svg"
			width="1em"
			height="1em"
			viewBox="0 0 24 24">
			<path
				d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
		</svg>
	</div>
	<!-- <i class="fab fa-github"></i> -->
	<a href="https://github.com/dude333/rapina">github.com/dude333/rapina</a>
</footer>
