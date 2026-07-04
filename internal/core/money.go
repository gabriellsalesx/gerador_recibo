package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Money representa um valor monetário em centavos (inteiro) para evitar erros
// de ponto flutuante. É compartilhado por recibo, orçamento e contrato.
type Money struct {
	Amount   int64  `json:"amount_cents"`
	Currency string `json:"currency"`
}

func NewMoney(cents int64, currency string) Money {
	if currency == "" {
		currency = "BRL"
	}
	return Money{Amount: cents, Currency: strings.ToUpper(currency)}
}

// ParseMoney aceita formatos como "250", "250,00", "250.00", "1.250,00" e
// "1250.00", normalizando para centavos.
func ParseMoney(input string) (Money, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return Money{}, errors.New("valor vazio")
	}

	s = strings.NewReplacer("R$", "", "r$", "", "BRL", "", "brl", "", " ", "").Replace(s)
	for _, r := range s {
		if !unicode.IsDigit(r) && r != ',' && r != '.' {
			return Money{}, fmt.Errorf("valor contém caractere inválido: %q", r)
		}
	}

	lastComma := strings.LastIndex(s, ",")
	lastDot := strings.LastIndex(s, ".")
	lastSep := lastComma
	if lastDot > lastSep {
		lastSep = lastDot
	}

	var wholePart, centsPart string
	if lastSep >= 0 {
		after := onlyDigits(s[lastSep+1:])
		if len(after) > 0 && len(after) <= 2 {
			wholePart = onlyDigits(s[:lastSep])
			centsPart = after
		} else {
			wholePart = onlyDigits(s)
		}
	} else {
		wholePart = onlyDigits(s)
	}

	if wholePart == "" {
		wholePart = "0"
	}
	for len(centsPart) < 2 {
		centsPart += "0"
	}
	if centsPart == "" {
		centsPart = "00"
	}

	whole, err := strconv.ParseInt(wholePart, 10, 64)
	if err != nil {
		return Money{}, fmt.Errorf("valor inválido: %w", err)
	}
	cents, err := strconv.ParseInt(centsPart[:2], 10, 64)
	if err != nil {
		return Money{}, fmt.Errorf("centavos inválidos: %w", err)
	}
	return NewMoney(whole*100+cents, "BRL"), nil
}

func (m Money) IsZero() bool {
	return m.Amount == 0
}

// Add e Sub facilitam cálculos de orçamento (subtotal, desconto, acréscimo).
func (m Money) Add(other Money) Money {
	return NewMoney(m.Amount+other.Amount, m.Currency)
}

func (m Money) Sub(other Money) Money {
	return NewMoney(m.Amount-other.Amount, m.Currency)
}

// MulQuantity multiplica um valor unitário por uma quantidade (em milésimos de
// unidade para suportar frações como 1,5h) retornando centavos arredondados.
func (m Money) MulQuantityMilli(qtyMilli int64) Money {
	cents := (m.Amount*qtyMilli + 500) / 1000
	return NewMoney(cents, m.Currency)
}

func (m Money) DecimalString() string {
	sign := ""
	amount := m.Amount
	if amount < 0 {
		sign = "-"
		amount = -amount
	}
	return fmt.Sprintf("%s%d.%02d", sign, amount/100, amount%100)
}

func (m Money) FormatBRL() string {
	sign := ""
	amount := m.Amount
	if amount < 0 {
		sign = "-"
		amount = -amount
	}
	return fmt.Sprintf("%sR$ %s,%02d", sign, formatThousands(amount/100), amount%100)
}

func onlyDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func formatThousands(n int64) string {
	raw := strconv.FormatInt(n, 10)
	if len(raw) <= 3 {
		return raw
	}
	var parts []string
	for len(raw) > 3 {
		parts = append([]string{raw[len(raw)-3:]}, parts...)
		raw = raw[:len(raw)-3]
	}
	if raw != "" {
		parts = append([]string{raw}, parts...)
	}
	return strings.Join(parts, ".")
}
