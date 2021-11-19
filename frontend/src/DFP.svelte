<!--
SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>

SPDX-License-Identifier: MIT
-->
<script>
  import Autocomplete from "./Autocomplete.svelte";

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
  let empresa;

  async function load(str) {
    if (!str || str.length < 2) return;

    [dfp, err] = await apiDFP(str);
    console.log("load()", dfp, err);
  }

  $: load(empresa);

  function format(n) {
    if (!n) return "-";
    return Math.round(n / 10e6).toLocaleString("pt-BR");
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

  function toggle(i) {
    const base = dfp.contas[i].codigo;
    let hasChild = false;
    dfp.contas.forEach((el, idx) => {
      if (idx != i && el.codigo.startsWith(base)) {
        hasChild = true;
        dfp.contas[idx].hide = !dfp.contas[i].collapse;
      }
    });
    if (hasChild) dfp.contas[i].collapse = !dfp.contas[i].collapse;
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

      {#each dfp.contas as conta, i}
        {#if !conta.hide}
          <tr style="font-weight: {fontWeight(conta.codigo)}">
            <td on:click={() => toggle(i)}>{conta.collapse ? '+' : '-'}&nbsp;{conta.codigo}</td>
            <td>{conta.descr}</td>
            {#each conta.totais as total}
              <td style="text-align:right;">{format(total)}</td>
            {/each}
          </tr>
        {/if}
      {/each}
    </table>
  </small>
{/if}
