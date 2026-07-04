package quote

import (
	"fmt"
	"strconv"
	"strings"

	"emissor/internal/core"
)

// ParseQuantity aceita "1", "2", "1,5", "2.5". Vazio vira 1. Retorna a
// quantidade em milésimos (ex.: "1,5" -> 1500).
func ParseQuantity(input string) (int64, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return 1000, nil
	}
	s = strings.ReplaceAll(s, ",", ".")
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("quantidade inválida: %s", input)
	}
	if f < 0 {
		return 0, fmt.Errorf("quantidade não pode ser negativa: %s", input)
	}
	return int64(f*1000 + 0.5), nil
}

// FormatQuantity formata a quantidade em milésimos para exibição ("1", "1,5").
func FormatQuantity(qtyMilli int64) string {
	whole := qtyMilli / 1000
	frac := qtyMilli % 1000
	if frac == 0 {
		return strconv.FormatInt(whole, 10)
	}
	s := fmt.Sprintf("%d.%03d", whole, frac)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return strings.ReplaceAll(s, ".", ",")
}

// NewLineItem cria um item calculando seu total (quantidade × valor unitário).
func NewLineItem(description string, qtyMilli int64, unit string, unitPrice core.Money) LineItem {
	return LineItem{
		Description:   strings.TrimSpace(description),
		QuantityMilli: qtyMilli,
		Unit:          strings.TrimSpace(unit),
		UnitPrice:     unitPrice,
		Total:         unitPrice.MulQuantityMilli(qtyMilli),
	}
}

// Recalculate recalcula o total de cada item, o subtotal e o total do
// orçamento (total = subtotal − desconto + acréscimo). Nunca usa float.
func Recalculate(q *Quote) {
	subtotal := core.NewMoney(0, "BRL")
	for i := range q.Items {
		item := &q.Items[i]
		item.Total = item.UnitPrice.MulQuantityMilli(item.QuantityMilli)
		subtotal = subtotal.Add(item.Total)
	}
	q.Subtotal = subtotal
	total := subtotal.Sub(q.Discount).Add(q.Surcharge)
	if total.Amount < 0 {
		total = core.NewMoney(0, "BRL")
	}
	q.Total = total
}
