package repositório

import (
	"errors"
	"fmt"
)

var (
	ErrArquivoInválido = errors.New("arquivo inválido")
	ErrDFPInválida     = errors.New("DFP inválida")

	ErrAnoInválidoFn = func(ano int) error { return fmt.Errorf("ano inválido: %d", ano) }
)
