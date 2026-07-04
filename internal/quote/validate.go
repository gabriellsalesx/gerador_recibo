package quote

import (
	"errors"
	"strings"
)

func Validate(q Quote) error {
	switch {
	case strings.TrimSpace(q.Issuer.Name) == "":
		return errors.New(`o campo "prestador" é obrigatório`)
	case strings.TrimSpace(q.Client.Name) == "":
		return errors.New(`o campo "cliente" é obrigatório`)
	case len(q.Items) == 0:
		return errors.New(`o orçamento precisa de pelo menos um item`)
	case q.CreatedAt.IsZero():
		return errors.New(`o campo "data de emissão" é obrigatório`)
	case q.ValidUntil.IsZero():
		return errors.New(`o campo "validade" é obrigatório`)
	}
	for i, item := range q.Items {
		if strings.TrimSpace(item.Description) == "" {
			return errors.New("todos os itens precisam de descrição")
		}
		if item.QuantityMilli <= 0 {
			return errors.New("todos os itens precisam de quantidade maior que zero")
		}
		if item.UnitPrice.Amount <= 0 {
			return errors.New("todos os itens precisam de valor unitário maior que zero")
		}
		_ = i
	}
	return nil
}
