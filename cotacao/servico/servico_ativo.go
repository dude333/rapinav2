package serviço

import (
	"context"
	domínio "github.com/dude333/rapinav2/cotacao/dominio"
	"github.com/pkg/errors"
)

// ativo é um serviço de busca de informações de um ativo em vários
// repositório, como banco de dados ou via API.
type ativo struct {
	repos []domínio.RepositórioAtivo
}

func Novo(repos []domínio.RepositórioAtivo) domínio.ServiçoAtivo {
	return &ativo{
		repos: repos,
	}
}

// Cotação busca a cotação de um ativo em vários repositórios com base
// no "código" de um ativo de um determinado "dia", retornando o primeiro
// valor encontado ou o erro de todos os repositórios.
func (s *ativo) Cotação(código string, dia domínio.Data) (*domínio.Ativo, error) {
	var errs error
	for i := range s.repos {
		ativo, err := s.repos[i].Cotação(context.Background(), código, dia)
		if err == nil {
			return ativo, nil
		}
		if errs == nil {
			errs = err
		} else {
			errs = errors.Wrap(err, errs.Error()+" :&")
		}
	}
	return &domínio.Ativo{}, errs
}
