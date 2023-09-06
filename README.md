<p align="center" style="text-align: center">
  <img src="https://i.postimg.cc/htdDDfdD/Rapina-logo.png" width="70%"><br/>
</p>
<p align="center">
Crie Relatórios Financeiros de Empresas Listadas na B3
 <br/>
  <br/>
  <a href="https://github.com/dude333/rapina/blob/master/LICENSE">
    <img alt="GitHub" src="https://img.shields.io/github/license/dude333/rapina"/>
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

Para a versão 2 é necessário instalar o compilador Go no seu computador. Basta seguir as instruções deste link: https://go.dev/doc/install

Também é necessário instalar o Git, que você pode encontrar aqui: https://git-scm.com/book/pt-br/v2/Come%C3%A7ando-Instalando-o-Git

Depois de instalados, abra o terminal ([CMD](https://superuser.com/a/340051/61616) no Windows) e execute os seguintes comandos:

1. `git clone github.com/dude333/rapina`
2. `cd rapina`
3. `git checkout v2`
4. `go build -o rapina cmd/*`

O arquivo `rapina`, ou `rapina.exe` no Windows, será criado.

## Uso

### Criação/Atualização dos Dados

Antes de se criar um relatório pela primeira vez, é **necessário** baixar os dados do site da CVM. 

Para isso, execute o seguinte comando no terminal:

`.\rapina atualizar [ano]`

Exemplos:
* `.\rapina atualizar`: baixar todos os dados.
* `.\rapina atualizar 2023`: baixar apenar um ano específico.

### Criação do Relatório

Para criar uma planilha com os dados financeiros trimestrais de um empresa, execute o seguinte comando:

`.\rapina relatorio [-d <DIRETORIO>]  [--crescente|-c]`

As empresas serão listadas em ordem alfabética. Basta navegar com as setas, ou use a tecla <kbd>/</kbd> para procurar uma empresa.

Exemplos:
* `\.rapina relatorio`: cria o relatório no diretório corrente.
* `\.rapina relatorio -d ./relats`: cria o relatório no diretório `relats`.
* `\.rapina relatorio -d ./relats -c`: cria o relatório no diretório `relats`, com os trimestres listados na ordem crescente.

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

Caso deseje mudar o local de gravação do banco de dados e dos arquivos temporários, criar o arquivo `rapina.yaml` no mesmo diretório do executável (`rapina` ou `rapina.exe`) com os seguintes dados:

```yaml
dataSrc: "/home/user1/dados/rapina.db?cache=shared&mode=rwc&_journal_mode=WAL&_busy_timeout=5000"
tempDir: "/home/user1/dados"
```

* `dataSrc`: arquivo do banco de dados.
* `tempDir`: diretório para arquivos temporários.

## Nota Final

Os relatórios tem finalidade apenas informativa e podem conter informações incorretas.

