package core

import (
	"fmt"
	"strings"
	"time"
)

// ParseDate aceita "hoje", "ontem", "DD/MM/AAAA" e "AAAA-MM-DD". Vazio vira hoje.
func ParseDate(input string, now time.Time) (time.Time, error) {
	s := strings.TrimSpace(strings.ToLower(input))
	if s == "" || s == "hoje" {
		return DateOnly(now), nil
	}
	if s == "ontem" {
		return DateOnly(now.AddDate(0, 0, -1)), nil
	}

	for _, layout := range []string{"02/01/2006", "2006-01-02"} {
		t, err := time.ParseInLocation(layout, s, now.Location())
		if err == nil {
			return DateOnly(t), nil
		}
	}
	return time.Time{}, fmt.Errorf("data inválida: %s", input)
}

// ParseValidity aceita, além dos formatos de ParseDate, "N dias" calculado a
// partir da data de emissão. Usado pelo orçamento.
func ParseValidity(input string, issuedAt time.Time) (time.Time, error) {
	s := strings.TrimSpace(strings.ToLower(input))
	if s == "" {
		return time.Time{}, nil
	}
	if strings.HasSuffix(s, "dias") || strings.HasSuffix(s, "dia") {
		numStr := strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(s, "dias"), "dia"))
		var n int
		if _, err := fmt.Sscanf(numStr, "%d", &n); err == nil && n >= 0 {
			return DateOnly(issuedAt.AddDate(0, 0, n)), nil
		}
	}
	return ParseDate(input, issuedAt)
}

func DateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
