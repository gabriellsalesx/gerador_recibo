package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"emissor/internal/config"
	"emissor/internal/core"
	"emissor/internal/pdf"
	"emissor/internal/receipt"
	"emissor/internal/storage"
)

type Service struct {
	ConfigPath string
	Now        func() time.Time
}

type CreateReceiptInput struct {
	IssuerName      string
	IssuerDocument  string
	IssuerCity      string
	IssuerState     string
	PayerName       string
	PayerDocument   string
	Value           string
	Description     string
	PaymentMethod   string
	PaidAt          string
	Notes           string
	OutputDir       string
	PixKey          string
	PixKeyType      string
	PixReceiverName string
	PixReceiverCity string
	DisablePix      bool
}

type CreateReceiptResult struct {
	Receipt    receipt.Receipt
	PDFPath    string
	JSONPath   string
	PixPNGPath string
	Warnings   []string
}

func (s Service) CreateReceipt(input CreateReceiptInput) (CreateReceiptResult, error) {
	now := time.Now
	if s.Now != nil {
		now = s.Now
	}
	createdAt := now()
	cfg, cfgPath, configExists, err := config.LoadOrDefault(s.ConfigPath)
	if err != nil {
		return CreateReceiptResult{}, err
	}

	applyInputToConfig(&cfg, input)

	amount, err := core.ParseMoney(input.Value)
	if err != nil {
		return CreateReceiptResult{}, fmt.Errorf("o campo \"valor\" é inválido: %w", err)
	}
	paidAt, err := core.ParseDate(input.PaidAt, createdAt)
	if err != nil {
		return CreateReceiptResult{}, err
	}

	number := storage.FormatNumber(cfg.Numbering.NextNumber, cfg.Numbering.Padding, cfg.Numbering.Prefix)
	if !cfg.Numbering.Enabled {
		number = createdAt.Format("20060102150405")
	}

	paymentMethod := strings.TrimSpace(input.PaymentMethod)
	if paymentMethod == "" {
		paymentMethod = cfg.Defaults.PaymentMethod
	}
	if paymentMethod == "" {
		paymentMethod = "Pix"
	}
	notes := input.Notes
	if notes == "" {
		notes = cfg.Defaults.Notes
	}

	r := receipt.Receipt{
		ID:            number,
		Number:        number,
		CreatedAt:     createdAt,
		IssuedAt:      createdAt,
		PaidAt:        paidAt,
		Issuer:        cfg.Issuer,
		Payer:         core.Party{Name: strings.TrimSpace(input.PayerName), Document: strings.TrimSpace(input.PayerDocument)},
		Amount:        amount,
		Description:   strings.TrimSpace(input.Description),
		PaymentMethod: paymentMethod,
		Notes:         notes,
		Locale:        "pt-BR",
	}
	if err := receipt.Validate(r); err != nil {
		return CreateReceiptResult{}, err
	}

	outputDir := input.OutputDir
	if outputDir == "" {
		outputDir = cfg.Output.Directory
	}
	files, err := storage.BuildFileSet(outputDir, core.DocReceipt, createdAt, number)
	if err != nil {
		return CreateReceiptResult{}, err
	}
	metadataFiles, err := storage.BuildFileSet(config.ReceiptsMetadataDir(cfgPath), core.DocReceipt, createdAt, number)
	if err != nil {
		return CreateReceiptResult{}, err
	}
	r.PDFFile = filepath.Base(files.PDFPath)
	r.PDFPath = files.PDFPath

	if shouldGeneratePix(cfg, input, paymentMethod) {
		pixPayment, err := buildPixPayment(cfg, r, metadataFiles)
		if err != nil {
			return CreateReceiptResult{}, fmt.Errorf("não consegui gerar o Pix: %w", err)
		}
		r.Pix = pixPayment
	}

	pdfOpts := pdf.Options{}
	var warnings []string
	if cfg.PDF.ShowLogo && strings.TrimSpace(cfg.PDF.LogoPath) != "" {
		img, err := loadPDFImage(cfg.PDF.LogoPath)
		if err != nil {
			warnings = append(warnings, "Não consegui carregar o logo: "+err.Error())
		} else {
			pdfOpts.LogoImage = img
		}
	}
	if cfg.PDF.ShowSignature && strings.TrimSpace(cfg.PDF.SignaturePath) != "" {
		img, err := loadPDFImage(cfg.PDF.SignaturePath)
		if err != nil {
			warnings = append(warnings, "Não consegui carregar a assinatura: "+err.Error())
		} else {
			pdfOpts.SignatureImage = img
		}
	}

	pdfBytes, err := pdf.Generate(r, pdfOpts)
	if err != nil {
		return CreateReceiptResult{}, fmt.Errorf("não consegui gerar o PDF: %w", err)
	}
	if err := storage.WriteNewFile(files.PDFPath, pdfBytes, 0o644); err != nil {
		return CreateReceiptResult{}, fmt.Errorf("não consegui salvar o PDF: %w", err)
	}
	if r.Pix != nil && len(r.Pix.QRCodePNG) > 0 {
		if err := storage.WriteNewFile(metadataFiles.PixPNGPath, r.Pix.QRCodePNG, 0o644); err != nil {
			return CreateReceiptResult{}, fmt.Errorf("não consegui salvar o QR Code Pix: %w", err)
		}
	}
	if err := storage.WriteNewJSON(metadataFiles.JSONPath, r); err != nil {
		return CreateReceiptResult{}, fmt.Errorf("não consegui salvar os metadados JSON: %w", err)
	}

	if configExists && cfg.Numbering.Enabled {
		cfg.Numbering.NextNumber++
		if err := config.Save(cfgPath, cfg); err != nil {
			warnings = append(warnings, "Recibo gerado, mas não consegui atualizar a numeração da configuração.")
		}
	}

	return CreateReceiptResult{
		Receipt:    r,
		PDFPath:    files.PDFPath,
		JSONPath:   metadataFiles.JSONPath,
		PixPNGPath: metadataFiles.PixPNGPath,
		Warnings:   warnings,
	}, nil
}

func applyInputToConfig(cfg *config.Config, input CreateReceiptInput) {
	if input.IssuerName != "" {
		cfg.Issuer.Name = strings.TrimSpace(input.IssuerName)
	}
	if input.IssuerDocument != "" {
		cfg.Issuer.Document = strings.TrimSpace(input.IssuerDocument)
	}
	if input.IssuerCity != "" {
		cfg.Issuer.City = strings.TrimSpace(input.IssuerCity)
	}
	if input.IssuerState != "" {
		cfg.Issuer.State = strings.TrimSpace(input.IssuerState)
	}
	if input.PixKey != "" {
		cfg.Issuer.PixKey = strings.TrimSpace(input.PixKey)
	}
	if input.PixKeyType != "" {
		cfg.Issuer.PixKeyType = strings.TrimSpace(input.PixKeyType)
	}
	if input.PixReceiverName != "" {
		cfg.Pix.ReceiverName = strings.TrimSpace(input.PixReceiverName)
	}
	if input.PixReceiverCity != "" {
		cfg.Pix.ReceiverCity = strings.TrimSpace(input.PixReceiverCity)
	}
}

func shouldGeneratePix(cfg config.Config, input CreateReceiptInput, paymentMethod string) bool {
	if input.DisablePix || !cfg.Pix.Enabled {
		return false
	}
	if !cfg.Pix.GenerateQRCodeByDefault {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(paymentMethod), "Pix") {
		return false
	}
	return strings.TrimSpace(cfg.Issuer.PixKey) != ""
}

func buildPixPayment(cfg config.Config, r receipt.Receipt, files storage.FileSet) (*core.PixPayment, error) {
	return buildPix(cfg, r.Amount, r.Number, r.Description, filepath.Base(files.PixPNGPath))
}

func loadPDFImage(path string) ([]byte, error) {
	expanded, err := storage.ExpandPath(path)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(expanded)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("arquivo não encontrado: %s", expanded)
		}
		return nil, err
	}
	if _, ok := pdf.DetectImageType(data); !ok {
		return nil, fmt.Errorf("formato não suportado (use PNG, JPG ou GIF): %s", expanded)
	}
	return data, nil
}
