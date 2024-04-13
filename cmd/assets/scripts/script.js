const _fromList = document.getElementById("state_container");
const _toList = document.getElementById("selected");
const _allOptions = document.getElementById("allOptions");
const _terminalDiv = document.getElementById("terminal");
const _filesDiv = document.getElementById("files");

const _customEvent = new CustomEvent("reloadFiles");
_filesDiv.dispatchEvent(_customEvent);

// Mover opções de um select para o outro -------------------------------------

let _options = [];

// Detecting a mobile browser
// https://stackoverflow.com/a/11381730/276311
function isMobile() {
  let check = false;
  (function (a) {
    if (
      /(android|bb\d+|meego).+mobile|avantgo|bada\/|blackberry|blazer|compal|elaine|fennec|hiptop|iemobile|ip(hone|od)|iris|kindle|lge |maemo|midp|mmp|mobile.+firefox|netfront|opera m(ob|in)i|palm( os)?|phone|p(ixi|re)\/|plucker|pocket|psp|series(4|6)0|symbian|treo|up\.(browser|link)|vodafone|wap|windows ce|xda|xiino/i.test(
        a,
      ) ||
      /1207|6310|6590|3gso|4thp|50[1-6]i|770s|802s|a wa|abac|ac(er|oo|s\-)|ai(ko|rn)|al(av|ca|co)|amoi|an(ex|ny|yw)|aptu|ar(ch|go)|as(te|us)|attw|au(di|\-m|r |s )|avan|be(ck|ll|nq)|bi(lb|rd)|bl(ac|az)|br(e|v)w|bumb|bw\-(n|u)|c55\/|capi|ccwa|cdm\-|cell|chtm|cldc|cmd\-|co(mp|nd)|craw|da(it|ll|ng)|dbte|dc\-s|devi|dica|dmob|do(c|p)o|ds(12|\-d)|el(49|ai)|em(l2|ul)|er(ic|k0)|esl8|ez([4-7]0|os|wa|ze)|fetc|fly(\-|_)|g1 u|g560|gene|gf\-5|g\-mo|go(\.w|od)|gr(ad|un)|haie|hcit|hd\-(m|p|t)|hei\-|hi(pt|ta)|hp( i|ip)|hs\-c|ht(c(\-| |_|a|g|p|s|t)|tp)|hu(aw|tc)|i\-(20|go|ma)|i230|iac( |\-|\/)|ibro|idea|ig01|ikom|im1k|inno|ipaq|iris|ja(t|v)a|jbro|jemu|jigs|kddi|keji|kgt( |\/)|klon|kpt |kwc\-|kyo(c|k)|le(no|xi)|lg( g|\/(k|l|u)|50|54|\-[a-w])|libw|lynx|m1\-w|m3ga|m50\/|ma(te|ui|xo)|mc(01|21|ca)|m\-cr|me(rc|ri)|mi(o8|oa|ts)|mmef|mo(01|02|bi|de|do|t(\-| |o|v)|zz)|mt(50|p1|v )|mwbp|mywa|n10[0-2]|n20[2-3]|n30(0|2)|n50(0|2|5)|n7(0(0|1)|10)|ne((c|m)\-|on|tf|wf|wg|wt)|nok(6|i)|nzph|o2im|op(ti|wv)|oran|owg1|p800|pan(a|d|t)|pdxg|pg(13|\-([1-8]|c))|phil|pire|pl(ay|uc)|pn\-2|po(ck|rt|se)|prox|psio|pt\-g|qa\-a|qc(07|12|21|32|60|\-[2-7]|i\-)|qtek|r380|r600|raks|rim9|ro(ve|zo)|s55\/|sa(ge|ma|mm|ms|ny|va)|sc(01|h\-|oo|p\-)|sdk\/|se(c(\-|0|1)|47|mc|nd|ri)|sgh\-|shar|sie(\-|m)|sk\-0|sl(45|id)|sm(al|ar|b3|it|t5)|so(ft|ny)|sp(01|h\-|v\-|v )|sy(01|mb)|t2(18|50)|t6(00|10|18)|ta(gt|lk)|tcl\-|tdg\-|tel(i|m)|tim\-|t\-mo|to(pl|sh)|ts(70|m\-|m3|m5)|tx\-9|up(\.b|g1|si)|utst|v400|v750|veri|vi(rg|te)|vk(40|5[0-3]|\-v)|vm40|voda|vulc|vx(52|53|60|61|70|80|81|83|85|98)|w3c(\-| )|webc|whit|wi(g |nc|nw)|wmlb|wonu|x700|yas\-|your|zeto|zte\-/i.test(
        a.substr(0, 4),
      )
    )
      check = true;
  })(navigator.userAgent || navigator.vendor || window.opera);
  return check;
}

function handleKeyPress(event) {
  if (event.key === "Enter") {
    event.preventDefault();
    moveOption(event.target);
  }
}

function mobileMoveOption(from) {
  if (!isMobile()) return;
  moveOption(from);
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

// Filtros ---------------------------------------------------------------------

const filterInput = document.getElementById("filterInput");
const selectElement = document.getElementById("state_container");
const titleEl = document.getElementById("title");

function mobileFilterOptions(filterInput) {
  if (!isMobile()) return;
  filterOptions(filterInput);
}

function filterOptions(filterInput) {
  const filterText = filterInput.value.toLowerCase();
  const options = selectElement.querySelectorAll("option");

  titleEl.innerText = filterText;
  const optionsCopy = [...options];

  // for (const option of options) {
  //   const optionText = option.textContent.toLowerCase();
  //   if (optionText.indexOf(filterText) !== -1) {
  //     selectElement.removeChild(option);
  //   }
  // }

  for (let i = selectElement.options.length - 1; i >= 0; i--) {
    if (
      selectElement.options[i].text.toLowerCase().indexOf(filterText) === -1
    ) {
      selectElement.remove(i);
    }
  }
}

