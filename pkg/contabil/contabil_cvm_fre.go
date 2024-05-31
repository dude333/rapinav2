package contabil

import (
	"context"

	"github.com/dude333/rapinav2/pkg/contabil/dominio"
	"github.com/dude333/rapinav2/pkg/progress"
)

type cvmFRE struct {
	infra infra
	cfg   *cfg
}

func NovaFRE(configs ...ConfigFn) (*cvmFRE, error) {
	var cvm cvmFRE
	cvm.cfg.loadConfigs(configs...)
	cvm.infra = &localInfra{dirDados: cvm.cfg.dirDados}

	return &cvm, nil
}

// Importar baixa o arquivo de FREs de todas as empresas de um determinado
// ano do site da CVM.
func (c *cvmFRE) Importar(ctx context.Context, ano int, trimestral bool) <-chan dominio.Resultado {
	results := make(chan dominio.Resultado)

	go func() {
		defer close(results)

		url := urlArquivo(ano, trimestral)

		arquivos, zipHash, err := c.infra.DownloadAndUnzip(url, filtros())
		if err != nil {
			results <- dominio.Resultado{Error: err}
			return
		}

		defer c.infra.Cleanup(arquivos)

		// if c.existe(zipHash) {
		// 	progress.Warning("Este arquivo 'dfp/itr' já foi processado anteriormente")
		// 	return
		// }

		for _, arquivo := range arquivos {
			progress.Running(arquivo.path)

			// Processa o arquivo e envia o resultado para o canal 'results'
			err = processarArquivoDFP(ctx, arquivo, results)
			if err != nil {
				results <- dominio.Resultado{Hash: arquivo.hash}
			}

			progress.RunOK()
		}

		// Grava o hash do zip no banco de dados
		results <- dominio.Resultado{Hash: zipHash}
	}()

	return results
}
