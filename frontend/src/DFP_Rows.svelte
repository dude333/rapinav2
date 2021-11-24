<script>
    import { fade } from "svelte/transition";

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

<tr
    transition:fade={{ duration: 100 }}
    style="font-weight: {fontWeight(codigo)}"
>
    <td on:click={toggle}>
        {codigo}
        <span style="font-weight: 300"
            >{subcontas ? (opened ? "(-)" : "(+)") : ""}</span
        >
    </td>
    <td>{descr}</td>
    {#each totais as total}
        <td style="text-align:right;">{format(total)}</td>
    {/each}
</tr>

{#if opened && subcontas}
    {#each subcontas as conta}
        <svelte:self {...conta} />
    {/each}
{/if}
