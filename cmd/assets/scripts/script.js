const _fromList = document.getElementById("state_container");
const _toList = document.getElementById("selected");
const _allOptions = document.getElementById("allOptions");
const _terminalDiv = document.getElementById("terminal");
const _filesDiv = document.getElementById("files");

const _customEvent = new CustomEvent("reloadFiles");
_filesDiv.dispatchEvent(_customEvent);

// Mover opções de um select para o outro -------------------------------------

let _options = [];

function handleKeyPress(event) {
  if (event.key === "Enter") {
    event.preventDefault();
    moveOption(event.target);
  }
}

function moveOption(from) {
  if (!from || !from.value) {
    return;
  }

  const obj = { cnpj: from.value, nome: from.options[from.selectedIndex].text };

  let to = _toList;
  if (from === _toList) {
    to = _fromList;
    // Remove item when moving back to _fromList
    _options = _options.filter((o) => o.cnpj !== obj.cnpj);
  } else {
    // Add item when moving to _toList
    _options.push(obj);
  }
  _allOptions.value = JSON.stringify(_options);

  addOption(from, to);
  removeOptionByIndex(from, from.selectedIndex);

  sortOptions(_fromList);
  sortOptions(_toList);
}

function addOption(from, to) {
  const option = document.createElement("option");
  const origOption = from.options[from.selectedIndex];
  option.setAttribute("data-index", origOption.getAttribute("data-index"));
  option.value = from.value;
  option.text = from.options[from.selectedIndex].text;
  to.appendChild(option);
}

function removeOptionByIndex(parent, index) {
  const options = parent.options;
  if (index >= 0 && index < options.length) {
    parent.removeChild(options[index]);
  }
}

function sortOptions(parent) {
  const options = parent.options;
  const sortedOptions = [...options].sort((a, b) => {
    const indexA = parseInt(a.getAttribute("data-index"));
    const indexB = parseInt(b.getAttribute("data-index"));
    return indexA - indexB;
  });
  parent.innerHTML = "";
  sortedOptions.forEach((option) => parent.appendChild(option));
}

// Terminal -------------------------------------------------------------------

document.body.addEventListener("htmx:afterSwap", function (event) {
  if (event.detail.elt.id === "terminal") {
    let responseData = event.detail.xhr.responseText;
    let processedData = formatData(responseData);
    _terminalDiv.innerHTML = processedData;
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
      _terminalDiv.innerHTML =
        "Criando relatórios...<br />" +
        _options
          .map((e) => "* " + e.cnpj + ": " + e.nome + "<br />")
          .join("\n");
    }
  });
document
  .querySelector("form.container")
  .addEventListener("htmx:afterRequest", (_) => {
    disableButtons(false);
    htmx.trigger(_filesDiv, "reloadFiles");
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
    _terminalDiv.innerHTML += formatData(event.data);
  };

  eventSource.addEventListener("close", function (event) {
    console.log("Server has closed the connection.");
    eventSource.close();
    running = false;
    disableButtons(false);
    _terminalDiv.innerHTML += "<br />[>] Importação concluída";
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
