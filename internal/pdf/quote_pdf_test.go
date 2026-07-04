package pdf

import (
	"bytes"
	"testing"
	"time"

	"emissor/internal/core"
	"emissor/internal/quote"
)

func TestGenerateQuotePDF(t *testing.T) {
	q := quote.Quote{
		ID:         "0001",
		Number:     "ORC-0001",
		CreatedAt:  time.Date(2026, 7, 3, 16, 0, 0, 0, time.UTC),
		IssuedAt:   time.Date(2026, 7, 3, 16, 0, 0, 0, time.UTC),
		ValidUntil: time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC),
		Issuer:     core.Party{Name: "Minha Empresa", City: "Fortaleza", State: "CE"},
		Client:     core.Party{Name: "Empresa XPTO"},
		Items: []quote.LineItem{
			quote.NewLineItem("Landing page", 1000, "un", core.NewMoney(180000, "BRL")),
			quote.NewLineItem("Hospedagem 12 meses", 1000, "un", core.NewMoney(60000, "BRL")),
		},
		PaymentTerms: "50% na aprovação, 50% na entrega",
	}
	quote.Recalculate(&q)
	data, err := GenerateQuote(q, Options{})
	if err != nil {
		t.Fatalf("GenerateQuote() returned error: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF")) {
		t.Fatal("GenerateQuote() did not return PDF bytes")
	}
}
