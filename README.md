# Rapina v2
## Estrutura

```
|--relatório
   |--repositorio
      |--sqlite.go
   |--apresentacao
      |--excel.go
      |--web.go
   |--servico

|--contabil
   |--repositorio
      |--b3.go
      |--infra.go
   |--apresentacao
      |--...
   |--servico
      |--empresa.go
      |--fii.go
   |--dominio
      |--empresa.go
      |--fii.go

|--dividendos
   |--repositorio
      |--b3.go
      |--infra.go
   |--apresentacao
      |--...
   |--servico
      |--empresa.go
      |--fii.go
   |--dominio
      |--empresa.go
      |--fii.go

|--cotacao
   |--repositorio
      |--b3.go          => baixa arquivo cotações, armazena no BD e retorna cotação de um ativo
      |--infra.go       => rotinas de download e unzip de arquivos
   |--apresentacao
      |--...
   |--servico
      |--ativo.go       => retorna a cotação de um ativo com dados dos repositórios (B3, banco de dados...)
   |--dominio
      |--ativo.go

```

## Roadmap

1. Cotação de um ativo
    * Pasta: `cotacao`
    * Modelo: `Ativo`
    * Dados obtidos de várias fontes:
        * Busca nos servidores: B3, Yahoo Finance, Alpha Vantage...
    * O resultado é salvo no banco de dados

2. Relatórios contábeis
    * Pasta: `contabil`
    * Modelo: `Empresa` e `FII`
    * Dados obtidos da CVM
        * BPA, BPP, DFC
        * FII
    * O resultado é salvo no banco de dados

3. Dividendos
    * Pasta: `dividendos`
    * Modelos: `Ativo`
    * Dados obtidos da B3
    * O resultado é salvo no banco de dados
