const slim = new SlimSelect({
  select: '#select-empresas',
  settings: { searchPlaceholder: 'Buscar empresa...', placeholderText: 'Selecione as empresas...' }
});

let eventSource = null;
let busy = false;

function connectSSE() {
  if (eventSource) eventSource.close();
  eventSource = new EventSource('/api/stream');

  eventSource.onopen = () => setSSEStatus('connected', 'Conectado');

  eventSource.onmessage = (e) => {
    const msg = e.data;
    let cls = '';
    if (/erro|error|falha/i.test(msg)) cls = 'err';
    else if (/aviso|warn/i.test(msg))  cls = 'warn';
    logLine(msg, cls);
  };

  // Evento "status" para sinalizar início/fim de processos específicos (ex: update-db)
  eventSource.addEventListener("status", (e) => {
    const st = JSON.parse(e.data);
    if (st.updateDbRunning) {
      setBusy(true);
      const elapsed = Math.floor((Date.now() - st.updateDbStartedUnix * 1000) / 1000);
      logLine(`Atualização em execução (jobId=${st.jobId}, tempo decorrido: ${elapsed}s)`);
    } else {
      setBusy(false);
    }
  });


  // Evento "close" sinaliza fim do processamento (broadcast de [done])
  eventSource.addEventListener('close', (e) => {
    setBusy(false);
    logLine('✔ ' + e.data, 'info');
    loadFiles();
  });

  eventSource.onerror = () => {
    setSSEStatus('', 'Reconectando...');
    setTimeout(connectSSE, 3000);
  };
}

function setSSEStatus(cls, text) {
  document.getElementById('sse-dot').className = 'dot ' + cls;
  document.getElementById('sse-status').textContent = text;
}

function logLine(text, cls) {
  const el = document.getElementById('console');
  const line = document.createElement('div');
  if (cls) line.className = cls;
  line.textContent = text.replace(/\x1b\[([0-9;]*)m/g, ''); // remove ESC codes
  el.appendChild(line);
  el.scrollTop = el.scrollHeight;
}

document.getElementById('btn-clear-console').addEventListener('click', () => {
  document.getElementById('console').innerHTML = '';
});

function setBusy(state) {
  busy = state;
  document.getElementById('btn-relatorio').disabled = state;
  document.getElementById('btn-update-db').disabled = state;
  if (state) setSSEStatus('busy', 'Processando...');
  else setSSEStatus('connected', 'Conectado');
}

async function loadEmpresas() {
  try {
    const res = await fetch('/api/empresas');
    if (!res.ok) throw new Error('HTTP ' + res.status);
    const empresas = await res.json();
    slim.setData(empresas.map(e => ({ text: e.nome, value: e.cnpj })));
  } catch(err) {
    logLine('Erro ao carregar empresas: ' + err.message, 'err');
  }
}

async function loadFiles() {
  const container = document.getElementById('file-list');
  try {
    const res = await fetch('/api/files');
    if (res.status === 204) {
      container.innerHTML = '<p class="empty">Nenhum arquivo encontrado.</p>';
      return;
    }
    if (!res.ok) throw new Error('HTTP ' + res.status);
    const files = await res.json();

    let html = '<div class="file-table">'  
      + '<div class="file-row header">'
      + '<div class="cell" data-title="Nome">Nome</div><div class="cell" data-title="Tamanho">Tamanho</div><div class="cell" data-title="Modificado">Modificado</div>'
      + '</div>';

    for (const f of files) {
      const icon = f.link ? '<div class="icon-googlesheets"></div> ' : '<div class="icon-xlsx"></div> ';
      const link  = '<a href="' + (f.link || ('/relatorios/' + f.name)) + '" target="_blank" rel="noopener">' + escHtml(f.name) + '</a>';
      html += '<div class="file-row">' 
        + '<div class="cell" data-title="Nome">' + icon + link + '</div>'
        + '<div class="cell" data-title="Tamanho">' + escHtml(f.size) + '</div>'
        + '<div class="cell" data-title="Modificado">' + escHtml(f.modTime) + '</div>'
        + '</div>';
    }
    html += '</div>';

   container.innerHTML = html;
  } catch(err) {
    container.innerHTML = '<p class="empty">Erro ao carregar arquivos.</p>';
    logLine('Erro ao carregar arquivos: ' + err.message, 'err');
  }
}

function escHtml(s) {
  return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}

document.getElementById('btn-relatorio').addEventListener('click', async () => {
  const selecionados = slim.getSelected();
  if (!selecionados || selecionados.length === 0) {
    logLine('⚠ Selecione ao menos uma empresa.', 'warn');
    return;
  }
  // Reconstrói objetos {cnpj, nome} a partir dos itens selecionados no SlimSelect
  const allOptions = slim.getData().filter(o => selecionados.includes(o.value));
  const empresas = allOptions.map(o => ({ cnpj: o.value, nome: o.text }));
  const output = document.querySelector('input[name=output]:checked').value;

  setBusy(true);
  logLine('Gerando relatórios para ' + empresas.length + ' empresa(s)...', 'info');

  try {
    const res = await fetch('/api/relatorios', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({ empresas, output }),
    });
    if (!res.ok) {
      const body = await res.json().catch(() => ({}));
      throw new Error(body.error || 'HTTP ' + res.status);
    }
  } catch(err) {
    logLine('Erro: ' + err.message, 'err');
    setBusy(false);
  }
});

document.getElementById('btn-update-db').addEventListener('click', async () => {
  setBusy(true);
  logLine('Atualizando base de dados (DFP)...', 'info');
  try {
    const res = await fetch('/api/update-db', { method: 'POST' });
    if (!res.ok) {
      const body = await res.json().catch(() => ({}));
      throw new Error(body.error || 'HTTP ' + res.status);
    }
  } catch(err) {
    logLine('Erro: ' + err.message, 'err');
    setBusy(false);
  }
});

document.getElementById('btn-refresh-files').addEventListener('click', loadFiles);

connectSSE();
loadEmpresas();
loadFiles();