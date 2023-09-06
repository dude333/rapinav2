<p align="center" style="text-align: center">
  <img src="https://i.postimg.cc/htdDDfdD/Rapina-logo.png" width="70%"><br/>
</p>
<p align="center">
  Crie Relatórios Financeiros de Empresas Listadas na B3
  <br/>
  <br/>
  <a href="https://github.com/dude333/rapinav2/releases">
    <img alt="GitHub release" src="https://img.shields.io/github/tag/dude333/rapinav2.svg?label=latest"/>
  </a>
  <a href="https://github.com/dude333/rapinav2/blob/v2/LICENSE">
    <img alt="License" src="https://img.shields.io/github/license/dude333/rapinav2?label=license"/>
  </a>
<p>

# Rapina

### Nota da Versão 2

No momento, esta versão só apresenta dados trimestrais, e os dados de fluxo de caixa e DVA estão incompletos, pois a CVM só disponibiliza dados acumulados ao invés de dados trimestrais nestes casos.

## Introdução

Este programa processa os arquivos de demonstrações financeiras trimestrais (ITR) e anuais (DFP) do site da CVM e os armazena em um banco de dados local (sqlite). A partir desses dados, são extraídas informações do balanço patrimonial, fluxo de caixa, DRE (demonstração de resultado) e DVA (demonstração de valor adicionado).

O programa coleta vários arquivos desde 2010, incluindo informações do ano corrente e do ano anterior, permitindo a extração de dados de 2009.

Com base nestes dados, são gerados relatórios das demonstrações financeiras por empresa.

## Instalação

Baixe o executável da [página de release](https://github.com/dude333/rapinav2/releases) e renomeie o executável para rapinav2.exe (no caso do Windows) ou rapinav2 (para o Linux ou macOS).

## Uso

### Criação/Atualização dos Dados

Antes de se criar um relatório pela primeira vez, é **necessário** baixar os dados do site da CVM. 

Para isso, execute o seguinte comando no terminal:

`./rapinav2 atualizar [ano]`

Exemplos:
* `./rapinav2 atualizar`: baixar todos os dados.
* `./rapinav2 atualizar 2023`: baixar apenar um ano específico.

### Criação do Relatório

Para criar uma planilha com os dados financeiros trimestrais de um empresa, execute o seguinte comando:

`./rapinav2 relatorio [-d <DIRETORIO>]  [--crescente|-c]`

As empresas serão listadas em ordem alfabética. Basta navegar com as setas, ou use a tecla <kbd>/</kbd> para procurar uma empresa.

Exemplos:
* `./rapinav2 relatorio`: cria o relatório no diretório corrente.
* `./rapinav2 relatorio -d ./relats`: cria o relatório no diretório `relats`.
* `./rapinav2 relatorio -d ./relats -c`: cria o relatório no diretório `relats`, com os trimestres listados na ordem crescente.

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

## Configuração

### `rapina.yaml`

Caso deseje mudar o local de gravação do banco de dados e dos arquivos temporários, criar o arquivo `rapina.yaml` no mesmo diretório do executável (`rapinav2` ou `rapinav2.exe`) com os seguintes dados:

```yaml
dataSrc: "/home/user1/dados/rapinav2.db?cache=shared&mode=rwc&_journal_mode=WAL&_busy_timeout=5000"
tempDir: "/home/user1/dados"
```

* `dataSrc`: arquivo do banco de dados.
* `tempDir`: diretório para arquivos temporários.

## Build

Para compilar o código fonte, basta seguir as instruções deste link: https://go.dev/doc/install

Também é necessário instalar o Git, que você pode encontrar aqui: https://git-scm.com/book/pt-br/v2/Come%C3%A7ando-Instalando-o-Git

Depois de instalados, abra o terminal ([CMD](https://superuser.com/a/340051/61616) no Windows) e execute os seguintes comandos:

1. `git clone github.com/dude333/rapinav2`
2. `cd rapinav2`
4. `go build -o rapinav2 cmd/*`

O arquivo `rapinav2`, ou `rapinav2.exe` no Windows, será criado.

## Nota Final

Os relatórios tem finalidade apenas informativa e podem conter informações incorretas.

