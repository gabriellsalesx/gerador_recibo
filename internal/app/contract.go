package app

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"emissor/internal/config"
	"emissor/internal/contract"
	"emissor/internal/core"
	"emissor/internal/pdf"
	"emissor/internal/storage"
)

type CreateContractInput struct {
	// Emitente (contratado) — sobrescritas opcionais da config.
	IssuerName     string
	IssuerDocument string
	IssuerCity     string
	IssuerState    string
	// Contratante (cliente).
	ContractorName     string
	ContractorDocument string
	ContractorAddress  string
	Object             string
	Value              string
	PaymentTerms       string
	Term               string
	Place              string
	SignedDate         string
	Template           string   // chave do modelo (contract.Templates)
	Clauses            []string // títulos; usado se não houver modelo
	OutputDir          string
}

type CreateContractResult struct {
	Contract contract.Contract
	PDFPath  string
	JSONPath string
	Warnings []string
}

func (s Service) CreateContract(input CreateContractInput) (CreateContractResult, error) {
	now := time.Now
	if s.Now != nil {
		now = s.Now
	}
	createdAt := now()
	cfg, cfgPath, configExists, err := config.LoadOrDefault(s.ConfigPath)
	if err != nil {
		return CreateContractResult{}, err
	}
	applyIssuerOverrides(&cfg, input.IssuerName, input.IssuerDocument, input.IssuerCity, input.IssuerState)

	ccfg := cfg.Documents.Contract

	amount := core.NewMoney(0, "BRL")
	if strings.TrimSpace(input.Value) != "" {
		amount, err = core.ParseMoney(input.Value)
		if err != nil {
			return CreateContractResult{}, fmt.Errorf("valor inválido: %w", err)
		}
	}

	signedDate, err := core.ParseDate(input.SignedDate, createdAt)
	if err != nil {
		return CreateContractResult{}, err
	}

	place := strings.TrimSpace(input.Place)
	if place == "" {
		place = ccfg.Defaults.Place
	}

	number := storage.FormatNumber(ccfg.Numbering.NextNumber, ccfg.Numbering.Padding, ccfg.Numbering.Prefix)
	if !ccfg.Numbering.Enabled {
		number = createdAt.Format("20060102150405")
	}

	contractor := core.Party{
		Name:     strings.TrimSpace(input.ContractorName),
		Document: strings.TrimSpace(input.ContractorDocument),
		Address:  strings.TrimSpace(input.ContractorAddress),
	}

	clauseData := contract.ClauseData{
		Object:         strings.TrimSpace(input.Object),
		Amount:         amount,
		PaymentTerms:   strings.TrimSpace(input.PaymentTerms),
		Term:           strings.TrimSpace(input.Term),
		Place:          place,
		ContractorName: contractor.Name,
		ContractedName: cfg.Issuer.Name,
	}

	var clauses []contract.Clause
	if tmpl, ok := contract.TemplateByKey(input.Template); ok {
		clauses = tmpl.Build(clauseData)
	} else {
		titles := input.Clauses
		if len(titles) == 0 {
			titles = ccfg.Defaults.Clauses
		}
		if len(titles) == 0 {
			titles = config.DefaultClauses()
		}
		clauses = contract.BuildClauses(titles, clauseData)
	}

	c := contract.Contract{
		ID:           number,
		Number:       number,
		CreatedAt:    createdAt,
		SignedDate:   signedDate,
		Contractor:   contractor,
		Contracted:   cfg.Issuer,
		Object:       strings.TrimSpace(input.Object),
		Amount:       amount,
		PaymentTerms: strings.TrimSpace(input.PaymentTerms),
		Term:         strings.TrimSpace(input.Term),
		Place:        place,
		Clauses:      clauses,
		Locale:       "pt-BR",
	}
	if err := contract.Validate(c); err != nil {
		return CreateContractResult{}, err
	}

	outputDir := input.OutputDir
	if outputDir == "" {
		outputDir = cfg.Output.Directory
	}
	files, err := storage.BuildFileSet(outputDir, core.DocContract, createdAt, number)
	if err != nil {
		return CreateContractResult{}, err
	}
	metadataFiles, err := storage.BuildFileSet(config.MetadataDirFor(cfgPath, core.DocContract), core.DocContract, createdAt, number)
	if err != nil {
		return CreateContractResult{}, err
	}
	c.PDFFile = filepath.Base(files.PDFPath)
	c.PDFPath = files.PDFPath

	pdfBytes, err := pdf.GenerateContract(c, pdf.Options{})
	if err != nil {
		return CreateContractResult{}, fmt.Errorf("não consegui gerar o PDF: %w", err)
	}
	if err := storage.WriteNewFile(files.PDFPath, pdfBytes, 0o644); err != nil {
		return CreateContractResult{}, fmt.Errorf("não consegui salvar o PDF: %w", err)
	}
	if err := storage.WriteNewJSON(metadataFiles.JSONPath, c); err != nil {
		return CreateContractResult{}, fmt.Errorf("não consegui salvar os metadados JSON: %w", err)
	}

	var warnings []string
	if configExists && ccfg.Numbering.Enabled {
		cfg.Documents.Contract.Numbering.NextNumber++
		if err := config.Save(cfgPath, cfg); err != nil {
			warnings = append(warnings, "Contrato gerado, mas não consegui atualizar a numeração da configuração.")
		}
	}

	return CreateContractResult{
		Contract: c,
		PDFPath:  files.PDFPath,
		JSONPath: metadataFiles.JSONPath,
		Warnings: warnings,
	}, nil
}
