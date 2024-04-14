const $options = document.getElementById("allOptions");
const $terminalDiv = document.getElementById("terminal");
const $filesDiv = document.getElementById("files");

const _customEvent = new CustomEvent("reloadFiles");
$filesDiv.dispatchEvent(_customEvent);

// Select -----------------------------------------------------------------------

function setOpts(newVal) {
  const obj = slimSelect
    .getData()
    .filter((e) => e.selected)
    .map((e) => {
      return { cnpj: e.value, nome: e.text };
    });
  $options.value = JSON.stringify(obj);
}

const slimSelect = new SlimSelect({
  select: "#empresas",
  settings: {
    placeholderText: "Selecione as empresas",
    searchText: "Nenhum item encontrado",
    searchPlaceholder: "Procurar",
    searchHighlight: false,
  },
  events: {
    afterChange: (newVal) => setOpts(newVal),
  },
});

setTimeout(() => {
  setOpts();
}, 100);

// Terminal -------------------------------------------------------------------

document.body.addEventListener("htmx:afterSwap", function (event) {
  if (event.detail.elt.id === "terminal") {
    let responseData = event.detail.xhr.responseText;
    let processedData = formatData(responseData);
    $terminalDiv.innerHTML = processedData;
  }
});

let currentColor = "0";
function formatData(data) {
  if (data.trim().length === 0) {
    return data;
  }
  // Replace ANSI escape codes with HTML span elements for colors
  const formattedData = data
    .replace(/\x1b\[(\d+)m/g, (_, colorCode) => {
      currentColor = colorCode;
      return "";
    })
    .replace(/\r|\n/g, "<br />");

  if (!formattedData) {
    return "";
  }

  return `<span class="color${currentColor}m">` + formattedData + "</span>";
}

// Submit form (criar relatórios) --------------------------------------------

document
  .querySelector("form.container")
  .addEventListener("htmx:beforeRequest", (event) => {
    if ((event.ta = document.getElementById("submit"))) {
      disableButtons(true);
      $terminalDiv.innerHTML = "Criando relatórios...<br />";
      //   $options
      //     .map((e) => "* " + e.cnpj + ": " + e.nome + "<br />")
      //     .join("\n");
    }
  });
document
  .querySelector("form.container")
  .addEventListener("htmx:afterRequest", (_) => {
    disableButtons(false);
    htmx.trigger($filesDiv, "reloadFiles");
  });

// Atualizar banco de dados --------------------------------------------------

function startEventSource(event) {
  event.preventDefault();

  console.log("Starting EventSource...");

  const eventSource = new EventSource("/update");
  disableButtons(true);

  eventSource.onmessage = function (event) {
    if (event.data.trim().length === 0) {
      return;
    }
    $terminalDiv.innerHTML += formatData(event.data);
  };

  eventSource.addEventListener("close", function (event) {
    console.log("Server has closed the connection.");
    eventSource.close();
    running = false;
    disableButtons(false);
    $terminalDiv.innerHTML += "<br />[>] Importação concluída";
  });

  eventSource.onerror = function (event) {
    console.error("EventSource failed:", event);
    eventSource.close();
    disableButtons(false);
    running = false;
  };
}

// Botões ---------------------------------------------------------------------

function disableButtons(disable) {
  const buttonsInDiv = document.querySelectorAll("#buttons button");
  buttonsInDiv.forEach((button) => {
    button.disabled = disable;
    button.style.cursor = disable ? "not-allowed" : "pointer";
    button.style.opacity = disable ? "0.5" : "1";
  });
}
