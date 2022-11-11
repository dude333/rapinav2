# Rapina v2
## Estrutura

```
internal/
├── config.go
├── dinheiro.go
├── empresa.go
├── contabil/
│   ├── api/
│   │   ├── api_dfp.go                 => api para apresentação das demonstrações financeiras
│   │   └── api_dfp_test.go
│   ├── contabil.go => entidades
│   ├── contabil_test.go
│   ├── repositorio/
│   │   ├── repositorio_config.go
│   │   ├── repositorio_cvm.go         => baixa dados de DPF e ITR da CVM
│   │   ├── repositorio_cvm_test.go
│   │   ├── repositorio_erros.go
│   │   ├── repositorio_infra.go       => interface para serviços de download e compressão de arquivos
│   │   ├── repositorio_sqlite.go      => armazena dados no sqlite
│   │   └── repositorio_sqlite_test.go
│   └── servico/
│       ├── servico_contabil.go        => lógica do processamento dos dados
│       └── servico_contabil_test.go
└── cotacao/
    ├── ativo.go
    ├── repositorio/
    │   ├── repositorio_b3.go          => baixa arquivo cotações, armazena no BD e retorna cotação de um ativo
    │   ├── repositorio_b3_test.go
    │   ├── repositorio_consts.go
    │   └── repositorio_infra.go       => interface para serviços de download e compressão de arquivos
    └── servico/
        ├── erros.go
        ├── servico.go                 => retorna a cotação de um ativo com dados dos repositórios (B3, banco de dados...)
        └── servico_test.go


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
    
    2.1. Relatório CSV
        * Dados (na horizontal):
            Ano
            Pat. Líq.
            Receita Líq.
            EBITDA
            Res. Fin.
            Lucro Líq.
            Mrg. Líq.
            ROE
            Caixa
            Dívida
            D. L. / EBITDA
            FCO
            CAPEX
            FCF
            FCL CAPEX
            Prov.
            Payout


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

* https://martinfowler.com/bliki/PresentationDomainDataLayering.html