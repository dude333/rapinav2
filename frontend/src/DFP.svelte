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

  onMount(async () => {
    // [dfp, err] = await apiDFP("84.429.695%2F0001-11", "2020");
    // console.log("onMount()", dfp, err);
  });

  async function load(str) {
    if (!str || str.length < 2) return;
    
    [dfp, err] = await apiDFP(str, "2020");
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

      {#each dfp.contas as conta}
        <tr
          class={tag(conta.codigo)}
          style="font-weight: {fontWeight(conta.codigo)}"
          on:click={() => toggle(conta.codigo)}
        >
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

<style>
  table * {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen,
      Ubuntu, Cantarell, "Open Sans", "Helvetica Neue", sans-serif;
  }
</style>
