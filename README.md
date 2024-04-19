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

`rapinav2 relatorio [-d <DIRETORIO>]  [--crescente|-c]`

As empresas serão listadas em ordem alfabética. Basta navegar com as setas, ou use a tecla <kbd>/</kbd> para procurar uma empresa.

Exemplos:

- `rapinav2 relatorio`: cria o relatório no diretório corrente.
- `rapinav2 relatorio -d ./relats`: cria o relatório no diretório `relats`.
- `rapinav2 relatorio -d ./relats -c`: cria o relatório no diretório `relats`, com os trimestres listados na ordem crescente.

Os relatório será gravado com o nome da empresa. Exemplos:

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

Para criar os relatório em uma interface web, execute o comando `servidor`:

```sh
$ ./rapinav2 servidor

[>] Iniciando servidor em :8080

```
E acesse a página através do endereço http://localhost:8080, onde é possível criar relatórios das empresas selecionadas e atualizar os dados (do ano corrente e do ano anterior).

[![webserver.png](https://i.postimg.cc/Y9h1vNqD/webserver.png)](https://postimg.cc/PpnLcwD1)

## Configuração

### `rapina.yaml`

Personalize os parâmetros criando o arquivo `rapina.yaml` no mesmo diretório do executável (`rapinav2` ou `rapinav2.exe`) usando os seguintes parâmetros:

| Parâmetro   | Descrição                                                                        |
| ----------- | -------------------------------------------------------------------------------- |
| `dataSrc`   | Arquivo onde serão gravados os dados coletados <br> Default: ./.dados            |
| `tempDir`   | Diretório onde os arquivos temporários serão armazernados <br> Default: ./.dados |
| `reportDir` | Diretório onde os relatórios serão salvos <br> Default: ./                       |

Exemplo:

```yaml
dataSrc: "/home/user1/dados/rapinav2.db"
tempDir: "/home/user1/dados"
reportDir: "/home/user1/relatorios"
```

## Build

Para compilar o código fonte, siga estas instruções:

1. Instale o Go: https://go.dev/doc/install
2. Instale o Git: https://git-scm.com/book/pt-br/v2/Come%C3%A7ando-Instalando-o-Git
3. Abra o terminal (ou [CMD](https://superuser.com/a/340051/61616) no Windows) e execute os seguintes comandos:

```bash
git clone github.com/dude333/rapinav2
cd rapinav2
go build -o rapinav2 cmd/*
```

O arquivo `rapinav2`, ou `rapinav2.exe` no Windows, será criado.

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
