package app

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"emissor/internal/config"
	"emissor/internal/core"
	"emissor/internal/pdf"
	"emissor/internal/quote"
	"emissor/internal/storage"
)

type QuoteItemInput struct {
	Description string
	Quantity    string
	Unit        string
	UnitPrice   string
}

type CreateQuoteInput struct {
	IssuerName     string
	IssuerDocument string
	IssuerCity     string
	IssuerState    string
	ClientName     string
	ClientDocument string
	Items          []QuoteItemInput
	Discount       string
	Surcharge      string
	Validity       string
	PaymentTerms   string
	Deadline       string
	Notes          string
	OutputDir      string
	EnablePix      bool
}

type CreateQuoteResult struct {
	Quote      quote.Quote
	PDFPath    string
	JSONPath   string
	PixPNGPath string
	Warnings   []string
}

func (s Service) CreateQuote(input CreateQuoteInput) (CreateQuoteResult, error) {
	now := time.Now
	if s.Now != nil {
		now = s.Now
	}
	createdAt := now()
	cfg, cfgPath, configExists, err := config.LoadOrDefault(s.ConfigPath)
	if err != nil {
		return CreateQuoteResult{}, err
	}
	applyIssuerOverrides(&cfg, input.IssuerName, input.IssuerDocument, input.IssuerCity, input.IssuerState)

	qcfg := cfg.Documents.Quote

	items := make([]quote.LineItem, 0, len(input.Items))
	for _, it := range input.Items {
		if strings.TrimSpace(it.Description) == "" {
			continue
		}
		qty, err := quote.ParseQuantity(it.Quantity)
		if err != nil {
			return CreateQuoteResult{}, err
		}
		price, err := core.ParseMoney(it.UnitPrice)
		if err != nil {
			return CreateQuoteResult{}, fmt.Errorf("valor unitário inválido para %q: %w", it.Description, err)
		}
		items = append(items, quote.NewLineItem(it.Description, qty, it.Unit, price))
	}

	discount := core.NewMoney(0, "BRL")
	if strings.TrimSpace(input.Discount) != "" {
		discount, err = core.ParseMoney(input.Discount)
		if err != nil {
			return CreateQuoteResult{}, fmt.Errorf("desconto inválido: %w", err)
		}
	}
	surcharge := core.NewMoney(0, "BRL")
	if strings.TrimSpace(input.Surcharge) != "" {
		surcharge, err = core.ParseMoney(input.Surcharge)
		if err != nil {
			return CreateQuoteResult{}, fmt.Errorf("acréscimo inválido: %w", err)
		}
	}

	validity := strings.TrimSpace(input.Validity)
	if validity == "" && qcfg.Defaults.ValidityDays > 0 {
		validity = fmt.Sprintf("%d dias", qcfg.Defaults.ValidityDays)
	}
	validUntil, err := core.ParseValidity(validity, createdAt)
	if err != nil {
		return CreateQuoteResult{}, err
	}

	number := storage.FormatNumber(qcfg.Numbering.NextNumber, qcfg.Numbering.Padding, qcfg.Numbering.Prefix)
	if !qcfg.Numbering.Enabled {
		number = createdAt.Format("20060102150405")
	}

	paymentTerms := strings.TrimSpace(input.PaymentTerms)
	if paymentTerms == "" {
		paymentTerms = qcfg.Defaults.PaymentTerms
	}

	q := quote.Quote{
		ID:           number,
		Number:       number,
		CreatedAt:    createdAt,
		IssuedAt:     createdAt,
		ValidUntil:   validUntil,
		Issuer:       cfg.Issuer,
		Client:       core.Party{Name: strings.TrimSpace(input.ClientName), Document: strings.TrimSpace(input.ClientDocument)},
		Items:        items,
		Discount:     discount,
		Surcharge:    surcharge,
		PaymentTerms: paymentTerms,
		Deadline:     strings.TrimSpace(input.Deadline),
		Notes:        strings.TrimSpace(input.Notes),
		Locale:       "pt-BR",
	}
	quote.Recalculate(&q)
	if err := quote.Validate(q); err != nil {
		return CreateQuoteResult{}, err
	}

	outputDir := input.OutputDir
	if outputDir == "" {
		outputDir = cfg.Output.Directory
	}
	files, err := storage.BuildFileSet(outputDir, core.DocQuote, createdAt, number)
	if err != nil {
		return CreateQuoteResult{}, err
	}
	metadataFiles, err := storage.BuildFileSet(config.MetadataDirFor(cfgPath, core.DocQuote), core.DocQuote, createdAt, number)
	if err != nil {
		return CreateQuoteResult{}, err
	}
	q.PDFFile = filepath.Base(files.PDFPath)
	q.PDFPath = files.PDFPath

	if input.EnablePix && cfg.Pix.Enabled && strings.TrimSpace(cfg.Issuer.PixKey) != "" && q.Total.Amount > 0 {
		pixPayment, err := buildPix(cfg, q.Total, q.Number, "Orçamento "+q.Number, filepath.Base(metadataFiles.PixPNGPath))
		if err != nil {
			return CreateQuoteResult{}, fmt.Errorf("não consegui gerar o Pix: %w", err)
		}
		q.Pix = pixPayment
	}

	pdfBytes, err := pdf.GenerateQuote(q, pdf.Options{})
	if err != nil {
		return CreateQuoteResult{}, fmt.Errorf("não consegui gerar o PDF: %w", err)
	}
	if err := storage.WriteNewFile(files.PDFPath, pdfBytes, 0o644); err != nil {
		return CreateQuoteResult{}, fmt.Errorf("não consegui salvar o PDF: %w", err)
	}
	if q.Pix != nil && len(q.Pix.QRCodePNG) > 0 {
		if err := storage.WriteNewFile(metadataFiles.PixPNGPath, q.Pix.QRCodePNG, 0o644); err != nil {
			return CreateQuoteResult{}, fmt.Errorf("não consegui salvar o QR Code Pix: %w", err)
		}
	}
	if err := storage.WriteNewJSON(metadataFiles.JSONPath, q); err != nil {
		return CreateQuoteResult{}, fmt.Errorf("não consegui salvar os metadados JSON: %w", err)
	}

	var warnings []string
	if configExists && qcfg.Numbering.Enabled {
		cfg.Documents.Quote.Numbering.NextNumber++
		if err := config.Save(cfgPath, cfg); err != nil {
			warnings = append(warnings, "Orçamento gerado, mas não consegui atualizar a numeração da configuração.")
		}
	}

	return CreateQuoteResult{
		Quote:      q,
		PDFPath:    files.PDFPath,
		JSONPath:   metadataFiles.JSONPath,
		PixPNGPath: metadataFiles.PixPNGPath,
		Warnings:   warnings,
	}, nil
}
