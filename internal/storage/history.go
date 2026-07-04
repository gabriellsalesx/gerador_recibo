package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"emissor/internal/core"
)

// DocItem é um item de histórico genérico, válido para recibo, orçamento e
// contrato. É montado a partir dos metadados JSON de cada documento.
type DocItem struct {
	DocType      core.DocType
	ID           string
	Number       string
	CreatedAt    time.Time
	Counterparty string // pagador / cliente / contratante
	Amount       core.Money
	JSONPath     string
	PDFPath      string
}

// rawDoc captura os campos comuns e específicos de cada tipo para montar o
// histórico sem precisar conhecer a struct concreta.
type rawDoc struct {
	ID         string     `json:"id"`
	Number     string     `json:"number"`
	CreatedAt  time.Time  `json:"created_at"`
	PDFFile    string     `json:"pdf_file"`
	PDFPath    string     `json:"pdf_path"`
	Payer      core.Party `json:"payer"`      // recibo
	Client     core.Party `json:"client"`     // orçamento
	Contractor core.Party `json:"contractor"` // contrato
	Amount     core.Money `json:"amount"`     // recibo / contrato
	Total      core.Money `json:"total"`      // orçamento
}

func (r rawDoc) counterparty() string {
	for _, name := range []string{r.Payer.Name, r.Client.Name, r.Contractor.Name} {
		if strings.TrimSpace(name) != "" {
			return name
		}
	}
	return ""
}

func (r rawDoc) amount() core.Money {
	if r.Amount.Amount != 0 {
		return r.Amount
	}
	return r.Total
}

// ListDocItems varre as pastas informadas em busca dos metadados JSON de um
// tipo de documento e devolve os itens mais recentes primeiro.
func ListDocItems(docType core.DocType, baseDirs []string, limit int) ([]DocItem, error) {
	seen := map[string]bool{}
	var items []DocItem
	for _, baseDir := range baseDirs {
		dirItems, err := listDocItemsInDir(docType, baseDir)
		if err != nil {
			return nil, err
		}
		for _, item := range dirItems {
			key := string(item.DocType) + "|" + item.ID + "|" + item.PDFPath
			if item.ID == "" || seen[key] {
				continue
			}
			seen[key] = true
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	if limit > 0 && len(items) > limit {
		return items[:limit], nil
	}
	return items, nil
}

func listDocItemsInDir(docType core.DocType, baseDir string) ([]DocItem, error) {
	baseDir, err := ExpandPath(baseDir)
	if err != nil {
		return nil, err
	}
	var items []DocItem
	err = filepath.WalkDir(baseDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		var r rawDoc
		if err := json.Unmarshal(data, &r); err != nil {
			return nil
		}
		if r.ID == "" && r.Number == "" {
			return nil
		}
		items = append(items, DocItem{
			DocType:      docType,
			ID:           r.ID,
			Number:       r.Number,
			CreatedAt:    r.CreatedAt,
			Counterparty: r.counterparty(),
			Amount:       r.amount(),
			JSONPath:     path,
			PDFPath:      docPDFPath(path, r),
		})
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return items, nil
}

func docPDFPath(jsonPath string, r rawDoc) string {
	if r.PDFPath != "" {
		if filepath.IsAbs(r.PDFPath) {
			return r.PDFPath
		}
		return filepath.Join(filepath.Dir(jsonPath), r.PDFPath)
	}
	if r.PDFFile != "" {
		return filepath.Join(filepath.Dir(jsonPath), r.PDFFile)
	}
	ext := filepath.Ext(jsonPath)
	return strings.TrimSuffix(jsonPath, ext) + ".pdf"
}

// ReadJSON lê e decodifica um arquivo de metadados em value (usado para
// reemitir/duplicar documentos a partir do JSON salvo).
func ReadJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}
