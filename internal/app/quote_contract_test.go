package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"emissor/internal/config"
	"emissor/internal/core"
)

func TestCreateQuoteWritesPDFAndJSON(t *testing.T) {
	temp := t.TempDir()
	cfg := config.Default()
	cfg.Issuer = core.Party{Name: "Minha Empresa", City: "Fortaleza", State: "CE"}
	cfg.Output.Directory = filepath.Join(temp, "Emissor")
	configPath := filepath.Join(temp, "config.json")
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatalf("config.Save() error: %v", err)
	}

	now := time.Date(2026, 7, 3, 16, 0, 0, 0, time.UTC)
	result, err := Service{ConfigPath: configPath, Now: func() time.Time { return now }}.CreateQuote(CreateQuoteInput{
		ClientName: "Empresa XPTO",
		Items: []QuoteItemInput{
			{Description: "Landing page", Quantity: "1", Unit: "un", UnitPrice: "1800,00"},
			{Description: "Hospedagem", Quantity: "1", Unit: "un", UnitPrice: "600,00"},
		},
		Validity: "15 dias",
	})
	if err != nil {
		t.Fatalf("CreateQuote() error: %v", err)
	}
	for _, p := range []string{result.PDFPath, result.JSONPath} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected file %s: %v", p, err)
		}
	}
	if result.Quote.Total.Amount != 240000 {
		t.Fatalf("Total = %d, want 240000", result.Quote.Total.Amount)
	}
	if result.Quote.ValidUntil.Format("2006-01-02") != "2026-07-18" {
		t.Fatalf("ValidUntil = %s, want 2026-07-18", result.Quote.ValidUntil.Format("2006-01-02"))
	}
}

func TestCreateContractWritesPDFAndJSON(t *testing.T) {
	temp := t.TempDir()
	cfg := config.Default()
	cfg.Issuer = core.Party{Name: "Minha Empresa", Document: "11.111.111/0001-11", City: "Fortaleza", State: "CE"}
	cfg.Output.Directory = filepath.Join(temp, "Emissor")
	configPath := filepath.Join(temp, "config.json")
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatalf("config.Save() error: %v", err)
	}

	now := time.Date(2026, 7, 3, 17, 0, 0, 0, time.UTC)
	result, err := Service{ConfigPath: configPath, Now: func() time.Time { return now }}.CreateContract(CreateContractInput{
		ContractorName:     "Empresa XPTO",
		ContractorDocument: "00.000.000/0001-00",
		Object:             "Desenvolvimento de site institucional",
		Value:              "2400,00",
		PaymentTerms:       "50% na assinatura, 50% na entrega",
		Term:               "30 dias",
		SignedDate:         "03/07/2026",
	})
	if err != nil {
		t.Fatalf("CreateContract() error: %v", err)
	}
	for _, p := range []string{result.PDFPath, result.JSONPath} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected file %s: %v", p, err)
		}
	}
	if len(result.Contract.Clauses) == 0 {
		t.Fatal("expected default clauses on contract")
	}
}
