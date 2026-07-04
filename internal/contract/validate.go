package contract

import (
	"errors"
	"strings"
)

func Validate(c Contract) error {
	switch {
	case strings.TrimSpace(c.Contractor.Name) == "":
		return errors.New(`o campo "contratante" é obrigatório`)
	case strings.TrimSpace(c.Contracted.Name) == "":
		return errors.New(`o campo "contratado" é obrigatório`)
	case strings.TrimSpace(c.Object) == "":
		return errors.New(`o campo "objeto" é obrigatório`)
	case strings.TrimSpace(c.Term) == "":
		return errors.New(`o campo "prazo/vigência" é obrigatório`)
	case strings.TrimSpace(c.Place) == "":
		return errors.New(`o campo "local" é obrigatório`)
	case c.Amount.Amount <= 0 && strings.TrimSpace(c.PaymentTerms) == "":
		return errors.New(`informe o valor ou as condições de pagamento`)
	default:
		return nil
	}
}
