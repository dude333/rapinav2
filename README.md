<p align="center" style="text-align: center">
  <img src="https://i.postimg.cc/1zbBSYy5/Rapina-logo-rounded.png" width="40%">
</p>
<p align="center">
  Crie Relatórios Financeiros de Empresas Listadas na B3
  <br/>
  <br/>
  <a href="https://github.com/dude333/rapinav2/releases">
    <img alt="GitHub release" src="https://img.shields.io/github/tag/dude333/rapinav2.svg?label=download"/>
  </a>
  <a href="https://github.com/dude333/rapinav2/blob/v2/LICENSE">
    <img alt="License" src="https://img.shields.io/github/license/dude333/rapinav2?label=license"/>
  </a>
<p>

# Rapina

Rapina é um programa que cria relatórios financeiros de empresas listadas na B3, processando arquivos de demonstrações financeiras trimestrais (ITR) e anuais (DFP) do site da CVM e armazenando-os em um banco de dados local (sqlite). As informações são extraídas do balanço patrimonial, fluxo de caixa, DRE (demonstração de resultado) e DVA (demonstração de valor adicionado).

O programa coleta arquivos desde 2010. Como estes arquivos contém os dados do ano corrente e do ano anterior, foi possível também a extração de dados de 2009.

Com base nestes dados, são gerados relatórios das demonstrações financeiras por empresa.

## Instalação

Baixe o executável da [página de release](https://github.com/dude333/rapinav2/releases) e renomeie o executável para rapinav2.exe (no caso do Windows) ou rapinav2 (para o Linux ou macOS).

## Uso

### Criação/Atualização dos Dados

Antes de se criar um relatório pela primeira vez, **é necessário** baixar os dados do site da CVM. Para isso, execute o seguinte comando no terminal:

`rapinav2 atualizar <--all | --ano <ano>>`

Exemplos:

- `rapinav2 atualizar --all`: baixar e atualizar dados de todos os anos.
- `rapinav2 atualizar --ano 2023`: baixar e atualizar um ano específico.

### Criação do Relatório

Para criar uma planilha com os dados financeiros trimestrais de um empresa, execute o seguinte comando:

`rapinav2 relatorio [-d <DIRETORIO>]  [--crescente|-c] [--googlesheets|-s]`

**Opcional:** use o parâmetro `-s` para salvar os relatórios no _Google Drive_. Veja o passo a passo abaixo de como configurar o acesso ao _Google Sheets_ e _Google Drive_.

As empresas serão listadas em ordem alfabética. Basta navegar com as setas, ou use a tecla <kbd>/</kbd> para procurar uma empresa.

Exemplos:

- `rapinav2 relatorio`: cria o relatório no diretório corrente.
- `rapinav2 relatorio -d ./relats`: cria o relatório no diretório `relats`.
- `rapinav2 relatorio -d ./relats -c`: cria o relatório no diretório `relats`, com os trimestres listados na ordem crescente.

Os relatórios serão gravados com o nome da empresa. Exemplos:

```
3R_PETROLEUM_ÓLEO_E_GÁS_S.A.xlsx
AES_BRASIL_ENERGIA_S.A.xlsx
CIA_SANEAMENTO_DO_PARANA_-_SANEPAR.xlsx
FLEURY_S.A.xlsx
LOCALIZA_RENT_A_CAR_S.A.xlsx
LOJAS_RENNER_S.A.xlsx
RAIA_DROGASIL_S.A.xlsx
```

### Servidor Web

Para criar os relatórios em uma interface web, execute o comando `servidor`:

```sh
$ ./rapinav2 servidor [--googlesheets|-s]

[>] Iniciando servidor em :8005

```

**Opcional:** use o parâmetro `-s` para salvar os relatórios no _Google Drive_. Veja o passo a passo abaixo de como configurar o acesso ao _Google Sheets_ e _Google Drive_.

E acesse a página através do endereço http://localhost:8005, onde é possível criar relatórios das empresas selecionadas e atualizar os dados (do ano corrente e do ano anterior).

[![webserver.png](https://i.postimg.cc/Y9h1vNqD/webserver.png)](https://postimg.cc/PpnLcwD1)

## Configuração

### `rapina.yaml`

Personalize os parâmetros criando o arquivo `rapina.yaml` no mesmo diretório do executável (`rapinav2` ou `rapinav2.exe`) usando os seguintes parâmetros:

| Parâmetro                   | Descrição                                                                   |
| --------------------------- | --------------------------------------------------------------------------- |
| `dataSrc`                   | Arquivo onde serão gravados os dados coletados <br> Default: ./.dados       |
| `tempDir`                   | Pasta onde os arquivos temporários serão armazenados <br> Default: ./.dados |
| `relatorio.outputDir`       | Pasta onde os relatórios serão salvos <br> Default: ./                      |
| `relatorio.googleSheetsDir` | Pasta do _Google Drive_ onde os relatórios serão salvos                     |

Exemplo do _rapina.yaml_:

```yaml
dataSrc: "dados/rapina/rapina.db?cache=shared&mode=rwc&_journal_mode=WAL"
tempDir: "dados/rapina"
relatorio:
  outputDir: "dados/reports"
  googlesheetsDir: "/rapina"
```

## Build

Para compilar o código fonte, siga estas instruções:

1. Instale o Go: https://go.dev/doc/install
2. Instale o Git: https://git-scm.com/book/pt-br/v2/Come%C3%A7ando-Instalando-o-Git
3. Abra o terminal (ou [CMD](https://superuser.com/a/340051/61616) no Windows) e execute os seguintes comandos:

```bash
git clone github.com/dude333/rapinav2
cd rapinav2
go mod tidy
go build -o rapinav2 cmd/*
```

O arquivo `rapinav2`, ou `rapinav2.exe` no Windows, será criado.

---

## Passo a passo: configurando o acesso ao Google Cloud

Siga estes passos uma única vez para obter o arquivo `credentials.json` que o Rapina usa para se autenticar em seu nome.

### Passo 1 — Criar um projeto no Google Cloud

1. Acesse o [Google Cloud Console](https://console.cloud.google.com/).
2. Clique no seletor de projetos no topo da página e escolha **Novo projeto**.
3. Dê um nome ao projeto (ex.: `rapina`) e clique em **Criar**.
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
   - **Nome do app** — qualquer nome, ex.: `rapina`
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
6. Renomeie o arquivo baixado para `credentials.json` e coloque-o no diretório de trabalho a partir do qual você executa o programa.

> **Mantenha o `credentials.json` em segredo.**

### Passo 5 — Primeira execução (login interativo)

Na primeira execução, o programa exibirá uma URL como:

```
Authorise this application by visiting:

  https://accounts.google.com/o/oauth2/auth?...

```

1. Abra a URL no seu navegador.
2. Escolha a conta Google que será a proprietária das planilhas.
3. Revise as permissões ("Ver, editar, criar e excluir apenas os arquivos específicos do Google Drive usados com este app") e clique em **Permitir**.

O programa armazenará o token resultante em `token.json`. Execuções posteriores carregam o token automaticamente, sem necessidade de interação.

> **Mantenha o `token.json` em segredo.** Ele concede acesso em seu nome.

---

## Dados

- Relação tickers x CNPJ:
  ```
  https://sistemaswebb3-listados.b3.com.br/isinPage
  https://sistemaswebb3-listados.b3.com.br/isinProxy/IsinCall/GetTextDownload/
  => obj.geralPt.id => btoa(JSON.stringify(obj.geralPt.id))
  https://sistemaswebb3-listados.b3.com.br/isinProxy/IsinCall/GetFileDownload/{id}
  ```

## Nota Final

Os relatórios tem finalidade apenas informativa e podem conter informações incorretas.
