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


## Banco de Dados

1. Empresas
    * id
    * Nome
    * CNPJ
    * Ano
    * MesesAcumulados

2. Contas
    * id_empresa
    * etc.

=======

1. Empresas
    * id
    * Nome
    * CNPJ
    * Ano

2. DPF (necessário carregar primeiro para apagar todos os 
dados dessa empresa/ano nas tabelas Empresas, DFP e ITR antes de iniciar a inserção)
    * id_empresa
    * etc.

3. ITR
    * id_empresa
    * meses_acumulados
    * etc.

## Arquitetura

* Value object: https://levelup.gitconnected.com/practical-ddd-in-golang-value-object-4fc97bcad70
* Entity: https://levelup.gitconnected.com/practical-ddd-in-golang-entity-40d32bdad2a3
* Domain service: https://levelup.gitconnected.com/practical-ddd-in-golang-domain-service-4418a1650274
