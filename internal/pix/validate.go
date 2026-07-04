package pix

import (
	"errors"
	"strings"
	"unicode"
)

func Validate(p Payment) error {
	if strings.TrimSpace(p.Key) == "" {
		return errors.New("chave Pix não configurada")
	}
	if NormalizeName(p.ReceiverName) == "" {
		return errors.New("nome do recebedor Pix não informado")
	}
	if NormalizeCity(p.ReceiverCity) == "" {
		return errors.New("cidade do recebedor Pix não informada")
	}
	if p.AmountCents <= 0 {
		return errors.New("valor Pix deve ser maior que zero")
	}
	return validateKey(p.Key, p.KeyType)
}

func validateKey(key, keyType string) error {
	key = strings.TrimSpace(key)
	switch strings.ToLower(strings.TrimSpace(keyType)) {
	case "", "random", "aleatoria", "chave_aleatoria":
		if len(key) < 8 {
			return errors.New("chave Pix parece curta demais")
		}
	case "email":
		if !strings.Contains(key, "@") || !strings.Contains(key, ".") {
			return errors.New("chave Pix de e-mail inválida")
		}
	case "cpf":
		if countDigits(key) != 11 {
			return errors.New("chave Pix CPF deve ter 11 dígitos")
		}
	case "cnpj":
		if countDigits(key) != 14 {
			return errors.New("chave Pix CNPJ deve ter 14 dígitos")
		}
	case "phone", "telefone":
		if !strings.HasPrefix(key, "+") || countDigits(key) < 12 {
			return errors.New("chave Pix telefone deve usar formato internacional, exemplo +5585999999999")
		}
	default:
		return errors.New("tipo de chave Pix inválido")
	}
	return nil
}

func countDigits(s string) int {
	total := 0
	for _, r := range s {
		if unicode.IsDigit(r) {
			total++
		}
	}
	return total
}
