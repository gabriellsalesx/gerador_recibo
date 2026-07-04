package pdf

import (
	"bytes"
	"testing"
	"time"

	"emissor/internal/core"
	"emissor/internal/receipt"
)

func TestGeneratePDF(t *testing.T) {
	r := receipt.Receipt{
		ID:            "0001",
		Number:        "0001",
		CreatedAt:     time.Date(2026, 7, 3, 15, 30, 12, 0, time.UTC),
		IssuedAt:      time.Date(2026, 7, 3, 15, 30, 12, 0, time.UTC),
		PaidAt:        time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC),
		Issuer:        core.Party{Name: "Minha Empresa", City: "Fortaleza", State: "CE"},
		Payer:         core.Party{Name: "Joao Silva"},
		Amount:        core.NewMoney(25000, "BRL"),
		Description:   "Desenvolvimento de site",
		PaymentMethod: "Pix",
	}
	data, err := Generate(r, Options{})
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF")) {
		t.Fatal("Generate() did not return PDF bytes")
	}
	if got := countPDFPages(data); got != 1 {
		t.Fatalf("Generate() created %d pages, want 1", got)
	}
}

func TestGeneratePDFWithPixStaysOnOnePage(t *testing.T) {
	r := receipt.Receipt{
		ID:            "0001",
		Number:        "0001",
		CreatedAt:     time.Date(2026, 7, 3, 15, 30, 12, 0, time.UTC),
		IssuedAt:      time.Date(2026, 7, 3, 15, 30, 12, 0, time.UTC),
		PaidAt:        time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC),
		Issuer:        core.Party{Name: "Minha Empresa", Document: "00.000.000/0001-00", Phone: "(85) 99999-9999", Email: "contato@example.com", Address: "Rua Exemplo, 123", City: "Fortaleza", State: "CE"},
		Payer:         core.Party{Name: "Joao Silva", Document: "000.000.000-00"},
		Amount:        core.NewMoney(25000, "BRL"),
		Description:   "Desenvolvimento de site",
		PaymentMethod: "Pix",
		Notes:         "Recebi o valor descrito acima referente ao servico prestado.",
		Pix: &core.PixPayment{
			Enabled:       true,
			Key:           "contato@example.com",
			Amount:        core.NewMoney(25000, "BRL"),
			CopyPaste:     "00020126640014br.gov.bcb.pix0119contato@example.com0219PAGAMENTO DE RECIBO5204000053039865406250.005802BR5913MINHA EMPRESA6009FORTALEZA62110507REC000163044CCC",
			ShowCopyPaste: true,
			ShowKeyText:   true,
		},
	}
	data, err := Generate(r, Options{})
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}
	if got := countPDFPages(data); got != 1 {
		t.Fatalf("Generate() created %d pages, want 1", got)
	}
}

func countPDFPages(data []byte) int {
	return bytes.Count(data, []byte("/Type /Page")) - bytes.Count(data, []byte("/Type /Pages"))
}
