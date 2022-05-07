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
</script>

<style>
    tr:nth-child(2n) {
        background-color: #f9fcf9;
    }
    td {
        font-size: 0.8rem;
        font-weight: 400;
        text-align: right;
        white-space: nowrap;
        padding-left: 0.5em !important;
        padding-right: 0.5em !important;
    }
    .code {
        font-weight: 300;
        text-align: left;
    }
    .descr{
        text-align: left;
    }
    .av {
        font-size: 0.56rem;
        font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
    }
</style>

<tr transition:fade="{{ duration: 100 }}">
    <td class="code" on:click="{toggle}">
        <span style="font-weight: {fontWeight(codigo)}">{codigo}</span>
        {subcontas ? (opened ? "(-)" : "(+)") : ""}
    </td>
    <td class="descr"><span style="font-weight: {fontWeight(codigo)}">{descr}</span></td>
    {#each totais as total, i}
        <td>
            <Numbers n="{total}" p="0"/>
        </td>
        <td class="av">
            <Numbers
                n="{i < totais.length - 1 ? 100 * (total / totais[i + 1] - 1) : 0}"
                p="1"
                sufix="%"
            />
        </td>
    {/each}
</tr>

{#if opened && subcontas}
    {#each subcontas as conta (conta)}
        <svelte:self {...conta} />
    {/each}
{/if}
