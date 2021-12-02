<!--
SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>

SPDX-License-Identifier: MIT
-->
<script>
  import Autocomplete from "./Autocomplete.svelte";
  import { apiDFP } from "./dfp";
  import Rows from './DFP_Rows.svelte';

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

  let dfp = [];
  let err = "";
  let empresa;

  async function load(str) {
    if (!str || str.length < 2) return;

    [dfp, err] = await apiDFP(str);

    // dfp.contas.forEach((_, i) => toggle(i));
  }

  $: load(empresa);

  const toggle = idx => {
    dfp.contas[idx].collapsed = !dfp.contas[idx].collapsed;
		const hide = !!dfp.contas[idx].collapsed;
		
		for (let i = idx+1; i < dfp.contas.length; i++) {
			if (hide && dfp.contas[i].codigo.startsWith(dfp.contas[idx].codigo)) {
				dfp.contas[i].hide = true; 	
			}
			if (!hide) {
				const j = parent(dfp.contas[i].codigo)
				dfp.contas[i].hide = j >= 0 ? dfp.contas[j].collapsed : false;
			}
		}
	}

  const symbol = (conta, idx) => {
		for (let i = idx+1; i < dfp.contas.length; i++) {
			if (dfp.contas[i].codigo.startsWith(conta.codigo)) {
				return conta.collapsed ? "˃" : "˅";
			}
		}
		return "\xa0";
  }

  const parent = (code) => {
		const arr = code.split(".");
		arr.pop();
		const codigo = arr.join(".");
		if (codigo != "") {
			return dfp.contas.findIndex(x => x.n === codigo);
		}
		return -1;
	}


  // https://svelte.dev/repl/69efbdcbbb6743e9988f777ef0f906ed?version=3.44.0
  // background = linear-gradient(to top, #d7e7d7 40%, #f8fcf8 40%)
  // t.style.background = "linear-gradient(to right,"+col1+" "+percentage+"%, "+col2+" "+percentage+"%)";
</script>

<style>
  table * {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen,
      Ubuntu, Cantarell, "Open Sans", "Helvetica Neue", sans-serif;
  }

  table th {
    position: -webkit-sticky;
    position: sticky;
    top: 0;
    z-index: 1;
    background: #fff;
    box-shadow: 0 1px 1px 0px rgba(0, 0, 0, 0.4);
  }
</style>

<Autocomplete bind:empresa />

{#if err != ""}
  <p>Erro: {err}</p>
{/if}
{#if err == "" && dfp && dfp.cnpj != "" && dfp.contas}
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

      {#each dfp.contas as conta (conta)}
        <Rows {...conta} />
      {/each}

    </table>
  </small>
{/if}
