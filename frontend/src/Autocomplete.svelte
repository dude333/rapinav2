<script>
  let name = "world";

  let results = [];
  let timer;

  let inputFind;
  let divDropdown;
  let ulDropdown;

  async function query(str) {
    const res = await fetch(
      "https://autocomplete.clearbit.com/v1/companies/suggest?query=" + str
    );
    const json = await res.json();
    const list = json.map((obj) => obj.name);
    return list;
  }

  function navigate(ev) {
    console.log("navigate", ev.keyCode);
    switch (ev.keyCode) {
      case 13: // enter
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
    inputFind.value = ev.target.textContent;
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
      await showResults2(val);
      ev.target.value = val;
    }, 150);
  }

  async function showResults2(val) {
    results = await query(val);
  }
</script>

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

<form autocomplete="off">
  <label>Find:
    <input bind:this={inputFind} on:keyup={(ev) => debounce(ev)} />
    <div bind:this={divDropdown} class="autocomplete dropdown">
      <ul bind:this={ulDropdown}>
        {#each results as result}
          <li
            tabindex="0"
            on:click={(ev) => select(ev)}
            on:keyup={(ev) => navigate(ev)}>
            {result}
          </li>
        {/each}
      </ul>
    </div>
  </label>
</form>
