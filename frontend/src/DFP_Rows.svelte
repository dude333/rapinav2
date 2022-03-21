<script>
    import { fade } from "svelte/transition";
    import Numbers from "./Numbers.svelte";

    export let subcontas = [];
    export let codigo = "";
    export let descr = "";
    export let totais = [];

    let opened = false;
    const toggle = (_) => (opened = !opened);

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

    function format(n) {
        if (!n) return "-";
        return Math.round(n / 10e6).toLocaleString("pt-BR");
    }
</script>

<style>
    .code {
        white-space: nowrap;
    }
</style>

<tr
    transition:fade={{ duration: 100 }}
    style="font-weight: {fontWeight(codigo)}"
>
    <td class="code" on:click={toggle}>
        {codigo}
        <span style="font-weight: 300"
            >{subcontas ? (opened ? "(-)" : "(+)") : ""}</span
        >
    </td>
    <td>{descr}</td>
    {#each totais as total, i}
        <td><Numbers av={i < (totais.length-1) ? (total/totais[i+1])-1 : 0} n={format(total)} /></td>
    {/each}
</tr>

{#if opened && subcontas}
    {#each subcontas as conta (conta)}
        <svelte:self {...conta} />
    {/each}
{/if}
