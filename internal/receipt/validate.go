package receipt

import (
	"errors"
	"strings"
)

func Validate(r Receipt) error {
	switch {
	case strings.TrimSpace(r.Issuer.Name) == "":
		return errors.New(`o campo "emitente" é obrigatório`)
	case strings.TrimSpace(r.Payer.Name) == "":
		return errors.New(`o campo "pagador" é obrigatório`)
	case r.Amount.Amount <= 0:
		return errors.New(`o campo "valor" é obrigatório`)
	case strings.TrimSpace(r.Description) == "":
		return errors.New(`o campo "referente" é obrigatório`)
	case r.CreatedAt.IsZero():
		return errors.New(`o campo "data de emissão" é obrigatório`)
	case r.PaidAt.IsZero():
		return errors.New(`o campo "data de pagamento" é obrigatório`)
	case strings.TrimSpace(r.PaymentMethod) == "":
		return errors.New(`o campo "forma de pagamento" é obrigatório`)
	default:
		return nil
	}
}
