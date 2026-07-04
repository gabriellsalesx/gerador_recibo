package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"emissor/internal/config"
	"emissor/internal/core"
)

func TestCreateReceiptWritesPDFJSONAndPix(t *testing.T) {
	temp := t.TempDir()
	cfg := config.Default()
	cfg.Issuer = core.Party{
		Name:       "Minha Empresa",
		City:       "Fortaleza",
		State:      "CE",
		PixKey:     "contato@example.com",
		PixKeyType: "email",
	}
	cfg.Pix.ReceiverName = "Minha Empresa"
	cfg.Pix.ReceiverCity = "Fortaleza"
	cfg.Output.Directory = filepath.Join(temp, "Recibos")
	configPath := filepath.Join(temp, "config.json")
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatalf("config.Save() returned error: %v", err)
	}

	now := time.Date(2026, 7, 3, 15, 30, 12, 0, time.UTC)
	result, err := Service{
		ConfigPath: configPath,
		Now:        func() time.Time { return now },
	}.CreateReceipt(CreateReceiptInput{
		PayerName:     "Joao Silva",
		Value:         "250,00",
		Description:   "Desenvolvimento de site",
		PaymentMethod: "Pix",
		PaidAt:        "03/07/2026",
	})
	if err != nil {
		t.Fatalf("CreateReceipt() returned error: %v", err)
	}

	for _, path := range []string{result.PDFPath, result.JSONPath, result.PixPNGPath} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected file %s: %v", path, err)
		}
	}
	if filepath.Dir(result.JSONPath) == filepath.Dir(result.PDFPath) {
		t.Fatal("metadata JSON should not be saved next to the visible PDF")
	}
	entries, err := os.ReadDir(filepath.Dir(result.PDFPath))
	if err != nil {
		t.Fatalf("os.ReadDir() returned error: %v", err)
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".pdf" {
			t.Fatalf("visible receipt directory should contain only PDFs, found %s", entry.Name())
		}
	}
	if result.Receipt.Pix == nil || result.Receipt.Pix.CopyPaste == "" {
		t.Fatal("expected Pix data on receipt")
	}
}
