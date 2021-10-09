package infra

import (
	"context"

	domínio "github.com/dude333/rapinav2/cotacao/dominio"
)

// RepositórioAtivosMock implementa a interface RepositórioAtivos
type RepositórioAtivosMock struct{}

func (r *RepositórioAtivosMock) Cotação(ctx context.Context, código string, data domínio.Data) (*domínio.Ativo, error) {
	return &domínio.Ativo{}, ErrFalhaDownload
}
