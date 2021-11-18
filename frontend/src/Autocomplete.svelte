<script>
  export let empresa = "";

  let results = [];
  let timer;

  let inputFind;
  let ulDropdown;

  async function query2(str) {
    const res = await fetch(
      "https://autocomplete.clearbit.com/v1/companies/suggest?query=" + str
    );
    const json = await res.json();
    const list = json.map((obj) => obj.name);
    return list;
  }
  async function query(str) {
    const res = await fetch("/api/dfp/empresas/" + str);
    const json = await res.json();
    return json.empresas;
  }

  function navigate(ev) {
    console.log("navigate", ev.keyCode);
    switch (ev.keyCode) {
      case 9: // tab
      case 13: // enter
        results = [];
        select(ev);
        break;
      case 38: // up
        if (ev.target.previousSibling) {
          ev.target.previousSibling.focus();
        } else {
          inputFind.focus();
        }
        break;
      case 40: // down
        ev.target.nextSibling && ev.target.nextSibling.focus();
        break;
    }
  }

  function select(ev) {
    empresa = ev.target.textContent && ev.target.textContent.trim();
    inputFind.value = empresa;
    inputFind.focus();
  }

  async function debounce(ev) {
    if (ev.keyCode == 40) {
      ulDropdown.firstChild && ulDropdown.firstChild.focus();
      return;
    }
    clearTimeout(timer);
    timer = setTimeout(async () => {
      const val = ev.target.value;
      console.log("val", val);
      await showResults(val);
      ev.target.value = val;
    }, 150);
  }

  async function showResults(val) {
    const r = await query(val);
    if (r && r.length > 1) {
      results = r;
    }
  }
</script>

<form
  autocomplete="off"
  on:submit|preventDefault={() => console.log("===>", inputFind.value)}
>
  <label
    >Find:
    <input bind:this={inputFind} on:keyup={(ev) => debounce(ev)} />
    {#if results && results.length > 0}
      <div class="autocomplete dropdown">
        <ul bind:this={ulDropdown}>
          {#each results as result}
            <li
              tabindex="0"
              on:click={(ev) => select(ev)}
              on:keydown|preventDefault={(ev) => navigate(ev)}
            >
              {result}
            </li>
          {/each}
        </ul>
      </div>
    {/if}
  </label>
</form>

<style>
  .autocomplete.dropdown {
    inset: 141px auto auto 81px;
    min-width: 194px;
    background-color: #f8f8f8;
    position: absolute;
    box-shadow: 0 1px 3px 0px;
    border-radius: 3px;
    border: 1px solid #fafafa;
    z-index: 100;
    max-height: 250px;
    overflow-y: auto;
  }

  .autocomplete.dropdown ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .autocomplete.dropdown ul li {
    padding: 4px 12px;
  }

  .autocomplete.dropdown ul li:focus,
  .autocomplete.dropdown ul li:hover {
    background-color: #f0f0f0;
    cursor: pointer;
  }
</style>
