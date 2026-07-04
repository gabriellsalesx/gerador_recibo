package storage

import (
	"path/filepath"
	"testing"
	"time"

	"emissor/internal/core"
)

func TestBuildFileSet(t *testing.T) {
	base := t.TempDir()
	createdAt := time.Date(2026, 7, 3, 15, 30, 12, 0, time.UTC)
	files, err := BuildFileSet(base, core.DocReceipt, createdAt, "0001")
	if err != nil {
		t.Fatalf("BuildFileSet() returned error: %v", err)
	}
	wantPDF := filepath.Join(base, "Recibos", "2026", "07", "03", "recibo-20260703-153012-0001.pdf")
	if files.PDFPath != wantPDF {
		t.Fatalf("PDFPath = %q, want %q", files.PDFPath, wantPDF)
	}
}

func TestBuildFileSetQuote(t *testing.T) {
	base := t.TempDir()
	createdAt := time.Date(2026, 7, 3, 16, 0, 0, 0, time.UTC)
	files, err := BuildFileSet(base, core.DocQuote, createdAt, "0001")
	if err != nil {
		t.Fatalf("BuildFileSet() returned error: %v", err)
	}
	wantPDF := filepath.Join(base, "Orcamentos", "2026", "07", "03", "orcamento-20260703-160000-0001.pdf")
	if files.PDFPath != wantPDF {
		t.Fatalf("PDFPath = %q, want %q", files.PDFPath, wantPDF)
	}
}

func TestFormatNumber(t *testing.T) {
	if got := FormatNumber(7, 4, ""); got != "0007" {
		t.Fatalf("FormatNumber() = %q", got)
	}
	if got := FormatNumber(3, 4, "ORC-"); got != "ORC-0003" {
		t.Fatalf("FormatNumber() = %q", got)
	}
}

func TestListDocItemsIncludesPDFPath(t *testing.T) {
	metadataBase := t.TempDir()
	outputBase := t.TempDir()
	createdAt := time.Date(2026, 7, 3, 15, 30, 12, 0, time.UTC)
	metadataFiles, err := BuildFileSet(metadataBase, core.DocReceipt, createdAt, "0001")
	if err != nil {
		t.Fatalf("BuildFileSet() returned error: %v", err)
	}
	outputFiles, err := BuildFileSet(outputBase, core.DocReceipt, createdAt, "0001")
	if err != nil {
		t.Fatalf("BuildFileSet() returned error: %v", err)
	}
	meta := map[string]any{
		"id":         "0001",
		"number":     "0001",
		"created_at": createdAt,
		"pdf_file":   filepath.Base(outputFiles.PDFPath),
		"pdf_path":   outputFiles.PDFPath,
		"payer":      core.Party{Name: "Joao Silva"},
		"amount":     core.NewMoney(25000, "BRL"),
	}
	if err := WriteNewJSON(metadataFiles.JSONPath, meta); err != nil {
		t.Fatalf("WriteNewJSON() returned error: %v", err)
	}

	items, err := ListDocItems(core.DocReceipt, []string{metadataBase}, 10)
	if err != nil {
		t.Fatalf("ListDocItems() returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ListDocItems() returned %d items, want 1", len(items))
	}
	if items[0].PDFPath != outputFiles.PDFPath {
		t.Fatalf("PDFPath = %q, want %q", items[0].PDFPath, outputFiles.PDFPath)
	}
	if items[0].Counterparty != "Joao Silva" {
		t.Fatalf("Counterparty = %q, want %q", items[0].Counterparty, "Joao Silva")
	}
}
