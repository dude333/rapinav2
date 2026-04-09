# googlesheets

Pacote Go que fornece uma API de alto nível para criação e manipulação de documentos do Google Sheets. A autenticação é feita via OAuth 2.0 com uma credencial do tipo "Aplicativo de desktop", de forma que **somente você** tem acesso às planilhas criadas pelo programa. O escopo OAuth é restrito a `drive.file`, o que significa que a aplicação só enxerga os arquivos que ela mesma criou — nenhum outro arquivo do seu Google Drive fica visível.

---

## Interface

```go
type Spreadsheet interface {
    NewSheet(sheetName string) error
    SetZoom(zoomScale float64) error
    SetColWidth(widths []float64)
    FreezePane(cell string) error
    SetFont(size float64, bold, wrap bool) (int, error)
    SetNumber(size float64, bold bool, format string) (int, error)
    PrintCell(row, col, style int, value any)
    RemoveRow(row int) error
    RemoveCol(col int) error
    SaveAs(name string) error
    Close() error
}
```

> **Observação sobre `SetZoom`:** A API REST do Google Sheets v4 não expõe o campo de nível de zoom via `batchUpdate`. O método existe para compatibilidade futura e atualmente não faz nada.

---

## Passo a passo: configurando o acesso ao Google Cloud

Siga estes passos uma única vez para obter o arquivo `credentials.json` que o pacote usa para se autenticar em seu nome.

### Passo 1 — Criar um projeto no Google Cloud

1. Acesse o [Google Cloud Console](https://console.cloud.google.com/).
2. Clique no seletor de projetos no topo da página e escolha **Novo projeto**.
3. Dê um nome ao projeto (ex.: `my-sheets-app`) e clique em **Criar**.
4. Certifique-se de que o novo projeto está selecionado no seletor antes de continuar.

### Passo 2 — Habilitar as APIs necessárias

É preciso habilitar duas APIs: **Google Sheets API** e **Google Drive API**.

1. No menu lateral, acesse **APIs e serviços → Biblioteca**.
2. Pesquise por **Google Sheets API**, clique sobre ela e depois em **Ativar**.
3. Volte à Biblioteca, pesquise por **Google Drive API**, clique sobre ela e depois em **Ativar**.

### Passo 3 — Configurar a tela de consentimento OAuth

> Esta etapa define o que o usuário vê quando a aplicação solicita permissão. Como somente você usará esta aplicação, é possível mantê-la no modo _Teste_ indefinidamente.

1. Acesse **APIs e serviços → Tela de consentimento OAuth**.
2. Escolha **Externo** como tipo de usuário e clique em **Criar**.
3. Preencha os campos obrigatórios:
   - **Nome do app** — qualquer nome, ex.: `my-sheets-app`
   - **E-mail de suporte ao usuário** — seu endereço de e-mail
   - **Informações de contato do desenvolvedor** — seu endereço de e-mail
4. Clique em **Salvar e continuar** em todas as telas seguintes até chegar ao **Resumo**, depois clique em **Voltar ao painel**.
5. No painel da tela de consentimento, clique em **Publicar app** — ou mantenha no modo _Teste_ e adicione sua conta como usuário de teste na sub-etapa abaixo.

   **Se permanecer no modo Teste:** clique em **Adicionar usuários** em _Usuários de teste_ e adicione a conta Google com a qual você irá se autenticar. Apenas os usuários listados conseguem concluir o fluxo OAuth enquanto o app estiver nesse modo.

### Passo 4 — Criar uma credencial OAuth 2.0

1. Acesse **APIs e serviços → Credenciais**.
2. Clique em **Criar credenciais → ID do cliente OAuth**.
3. Em _Tipo de aplicativo_, escolha **Aplicativo de desktop**.
4. Dê um nome (ex.: `googlesheets-cli`) e clique em **Criar**.
5. Na caixa de diálogo que aparecer, clique em **Fazer download do JSON**.
6. Renomeie o arquivo baixado para `credentials.json` e coloque-o no diretório de trabalho a partir do qual você executa o programa (ou informe o caminho via `Config.CredentialsFile`).

> **Mantenha o `credentials.json` em segredo.** Adicione-o ao `.gitignore` e nunca o envie para o controle de versão.

### Passo 5 — Primeira execução (login interativo)

Na primeira execução, o pacote exibirá uma URL como:

```
Authorise this application by visiting:

  https://accounts.google.com/o/oauth2/auth?...

Paste the authorisation code here and press Enter:
```

1. Abra a URL no seu navegador.
2. Escolha a conta Google que será a proprietária das planilhas.
3. Revise as permissões ("Ver, editar, criar e excluir apenas os arquivos específicos do Google Drive usados com este app") e clique em **Permitir**.
4. Copie o código de autorização exibido no navegador.
5. Cole-o no terminal e pressione Enter.

O pacote armazenará o token resultante em `token.json` (ou em `Config.TokenFile`). Execuções posteriores carregam o token automaticamente, sem necessidade de interação.

> **Mantenha o `token.json` em segredo.** Ele concede acesso em seu nome. Adicione-o ao `.gitignore`.

### Entradas recomendadas no `.gitignore`

```
credentials.json
token.json
```

---

## Configuração do Nginx para OAuth

Se você executar a aplicação atrás de um proxy reverso (nginx), é necessário configurar o servidor para que o redirect URI do OAuth funcione corretamente. O pacote cria um servidor HTTP local para receber o código de autorização do Google.

### Configuração do Nginx

Adicione um bloco `server` ao seu arquivo de configuração do nginx (ex.: `/etc/nginx/sites-available/Default`):

```nginx
server {
    listen 80;
    server_name seu-dominio.com;

    # Proxy reverso para a aplicação
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $server_name;
        proxy_set_header X-Forwarded-Port $server_port;
    }

    # Rota específica para autenticação OAuth
    location ~* /oauth2 {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Registrando o Redirect URI no Google Cloud

Após configurar o nginx, você precisa registrar o redirect URI correto nas credenciais OAuth:

1. Acesse [Google Cloud Console](https://console.cloud.google.com/) → **APIs e serviços → Credenciais**.
2. Clique em seu **ID do cliente OAuth** (tipo: Aplicativo de desktop).
3. Na seção **URIs autorizados para redirecionamento**, adicione:
   - Se usar porta automática: `http://seu-dominio.com`
   - Se usar porta específica (ex: 8080): `http://seu-dominio.com:8080`

### Executando com Port Específico

Se preferir usar uma porta específica para o redirect URI, execute com a flag `--tokenport`:

```bash
relatorio relat --googlesheets --tokenport 8080
```

### Verificação do Redirect URI

Durante a primeira execução, o programa exibirá algo como:

```
Abra esta URL no seu navegador para autorizar o acesso:

  https://accounts.google.com/o/oauth2/auth?...&redirect_uri=http://localhost:8080

Aguardando autorização...
```

Se você estiver usando nginx:

- A URL exibida continuará mostrando `localhost:8080` (ou a porta configurada)
- **Cuidado:** Se acessar o programa através do domínio público (ex: `seu-dominio.com`), você precisará atualizar o redirect URI no Google Cloud para corresponder

### Certificado SSL/TLS

Para produção, configure SSL/TLS no nginx:

```nginx
server {
    listen 443 ssl http2;
    server_name seu-dominio.com;

    ssl_certificate /caminho/para/certificado.crt;
    ssl_certificate_key /caminho/para/chave.key;

    # ... resto da configuração acima ...
}

# Redirecionar HTTP para HTTPS
server {
    listen 80;
    server_name seu-dominio.com;
    return 301 https://$server_name$request_uri;
}
```

Nesse caso, registre o redirect URI como:

- `https://seu-dominio.com` ou
- `https://seu-dominio.com:porta` (se usar porta não-padrão)

````

---

## Instalação

```bash
go get github.com/yourusername/googlesheets
````

> Substitua `github.com/yourusername/googlesheets` pelo caminho de módulo real definido no seu `go.mod`.

---

## Início rápido

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/yourusername/googlesheets"
)

func main() {
    ctx := context.Background()

    // Cria uma nova planilha no Google Drive do usuário.
    // New retorna *Sheet, não uma interface — você obtém o tipo concreto completo.
    ss, err := googlesheets.New(ctx, googlesheets.Config{
        CredentialsFile: "credentials.json",
        TokenFile:       "token.json",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer ss.Close()

    // Registra estilos.
    cabecalho, _ := ss.SetFont(12, true, false)
    corpo, _     := ss.SetFont(11, false, false)
    dinheiro, _  := ss.SetNumber(11, false, `"R$"#,##0.00`)

    // Layout.
    ss.SetColWidth([]float64{200, 120, 120, 160})
    ss.FreezePane("A2") // congela a linha de cabeçalho

    // Cabeçalho.
    ss.PrintCell(0, 0, cabecalho, "Descrição")
    ss.PrintCell(0, 1, cabecalho, "Data")
    ss.PrintCell(0, 2, cabecalho, "Categoria")
    ss.PrintCell(0, 3, cabecalho, "Valor")

    // Dados.
    ss.PrintCell(1, 0, corpo, "Material de escritório")
    ss.PrintCell(1, 1, corpo, "2025-01-10")
    ss.PrintCell(1, 2, corpo, "Administrativo")
    ss.PrintCell(1, 3, dinheiro, 45.99)

    // Renomeia a planilha e envia as escritas pendentes.
    if err := ss.SaveAs("Relatório de Despesas"); err != nil {
        log.Fatal(err)
    }

    // SpreadsheetURL() é um método de *Sheet — sem necessidade de type assertion.
    fmt.Println("Concluído:", ss.SpreadsheetURL())
}
```

---

## Referência da API

### `New(ctx, cfg) (*Sheet, error)`

Cria uma nova planilha Google Sheets em branco e retorna um `*Sheet`.

### `Open(ctx, cfg, spreadsheetID) (*Sheet, error)`

Encapsula uma planilha existente. O `spreadsheetID` é a longa sequência alfanumérica presente na URL da planilha:
`https://docs.google.com/spreadsheets/d/<SPREADSHEET_ID>/edit`.

### `Config`

| Campo             | Padrão             | Descrição                                         |
| ----------------- | ------------------ | ------------------------------------------------- |
| `CredentialsFile` | `credentials.json` | Caminho para o arquivo JSON de credenciais OAuth2 |
| `TokenFile`       | `token.json`       | Caminho para o token OAuth2 em cache              |

### Métodos de `*Sheet`

| Método                              | Descrição                                                   |
| ----------------------------------- | ----------------------------------------------------------- |
| `NewSheet(name)`                    | Cria uma nova aba e a torna ativa                           |
| `SetZoom(scale)`                    | Sem efeito (limitação da API)                               |
| `SetColWidth(widths)`               | Define larguras de coluna em pixels (índice 0 = coluna A)   |
| `FreezePane(cell)`                  | Congela linhas/colunas até a célula informada em notação A1 |
| `SetFont(size, bold, wrap)`         | Registra um estilo de texto; retorna o índice do estilo     |
| `SetNumber(size, bold, fmt)`        | Registra um estilo numérico; retorna o índice do estilo     |
| `PrintCell(row, col, style, value)` | Escreve um valor com estilo (índice base 0)                 |
| `RemoveRow(row)`                    | Remove uma linha (índice base 0)                            |
| `RemoveCol(col)`                    | Remove uma coluna (índice base 0)                           |
| `SaveAs(name)`                      | Renomeia a planilha e envia as escritas pendentes           |
| `Close()`                           | Envia as escritas pendentes                                 |

---

## Segurança

- O escopo OAuth utilizado é `https://www.googleapis.com/auth/drive.file`. Este é o **escopo mais restritivo do Drive** que ainda permite a criação de arquivos. A aplicação não consegue ler, listar ou modificar nenhum arquivo que ela não tenha criado.
- `credentials.json` identifica seu projeto no Cloud; `token.json` concede acesso em tempo de execução. Trate ambos como segredos.
- A expiração do token é tratada automaticamente pela biblioteca OAuth2 (tokens de atualização têm longa duração para credenciais de Aplicativo de desktop).

---

## Executando os testes

```bash
go test ./...
```

Os testes cobrem o parser de notação A1 e os valores padrão de configuração, sem necessidade de acesso à rede.
