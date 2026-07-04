package config

import (
	"os"
	"path/filepath"
	"runtime"

	"emissor/internal/core"
)

// ConfigVersion é a versão atual do arquivo de configuração da suíte.
const ConfigVersion = 2

type Config struct {
	Version   int             `json:"version"`
	Issuer    core.Party      `json:"issuer"`
	Pix       PixConfig       `json:"pix"`
	Output    OutputConfig    `json:"output"`
	PDF       PDFConfig       `json:"pdf"`
	Numbering NumberingConfig `json:"numbering"` // numeração do recibo
	Defaults  DefaultsConfig  `json:"defaults"`  // padrões do recibo
	Documents DocumentsConfig `json:"documents"` // orçamento e contrato
}

type PixConfig struct {
	Enabled                 bool   `json:"enabled"`
	GenerateQRCodeByDefault bool   `json:"generate_qrcode_by_default"`
	ShowCopyPaste           bool   `json:"show_copy_paste"`
	ShowKeyText             bool   `json:"show_key_text"`
	ReceiverName            string `json:"receiver_name"`
	ReceiverCity            string `json:"receiver_city"`
	DefaultDescription      string `json:"default_description"`
	TxIDPrefix              string `json:"txid_prefix"`
}

type OutputConfig struct {
	Directory      string `json:"directory"`
	OrganizeByType bool   `json:"organize_by_type"`
	OrganizeByDate bool   `json:"organize_by_date"`
	DateLayout     string `json:"date_layout"`
}

type PDFConfig struct {
	Template      string `json:"template"`
	AccentColor   string `json:"accent_color"`
	ShowLogo      bool   `json:"show_logo"`
	LogoPath      string `json:"logo_path"`
	ShowSignature bool   `json:"show_signature"`
	SignaturePath string `json:"signature_path"`
}

type NumberingConfig struct {
	Enabled    bool   `json:"enabled"`
	NextNumber int    `json:"next_number"`
	Padding    int    `json:"padding"`
	Prefix     string `json:"prefix"`
}

type DefaultsConfig struct {
	Currency      string `json:"currency"`
	PaymentMethod string `json:"payment_method"`
	City          string `json:"city"`
	State         string `json:"state"`
	Notes         string `json:"notes"`
}

// DocumentsConfig agrupa a configuração específica de orçamento e contrato.
// O recibo continua usando os campos de topo (Numbering/Defaults).
type DocumentsConfig struct {
	Quote    QuoteConfig    `json:"quote"`
	Contract ContractConfig `json:"contract"`
}

type QuoteConfig struct {
	Template  string          `json:"template"`
	Numbering NumberingConfig `json:"numbering"`
	Defaults  QuoteDefaults   `json:"defaults"`
}

type QuoteDefaults struct {
	Currency     string `json:"currency"`
	ValidityDays int    `json:"validity_days"`
	PaymentTerms string `json:"payment_terms"`
	Notes        string `json:"notes"`
}

type ContractConfig struct {
	Template  string           `json:"template"`
	Numbering NumberingConfig  `json:"numbering"`
	Defaults  ContractDefaults `json:"defaults"`
}

type ContractDefaults struct {
	Place   string   `json:"place"`
	Clauses []string `json:"clauses"`
}

// DefaultClauses são os títulos de cláusula incluídos por padrão no contrato.
func DefaultClauses() []string {
	return []string{"OBJETO", "PAGAMENTO", "PRAZO", "OBRIGAÇÕES", "RESCISÃO", "FORO"}
}

func Default() Config {
	return Config{
		Version: ConfigVersion,
		Pix: PixConfig{
			Enabled:                 true,
			GenerateQRCodeByDefault: true,
			ShowCopyPaste:           true,
			ShowKeyText:             true,
			DefaultDescription:      "Pagamento",
			TxIDPrefix:              "REC",
		},
		Output: OutputConfig{
			Directory:      DefaultOutputDir(),
			OrganizeByType: true,
			OrganizeByDate: true,
			DateLayout:     "year/month/day",
		},
		PDF: PDFConfig{
			Template:    "professional",
			AccentColor: "#1F2937",
		},
		Numbering: NumberingConfig{
			Enabled:    true,
			NextNumber: 1,
			Padding:    4,
		},
		Defaults: DefaultsConfig{
			Currency:      "BRL",
			PaymentMethod: "Pix",
			Notes:         "Recebi o valor descrito acima referente ao serviço prestado.",
		},
		Documents: defaultDocuments(),
	}
}

func defaultDocuments() DocumentsConfig {
	return DocumentsConfig{
		Quote: QuoteConfig{
			Template:  "professional",
			Numbering: NumberingConfig{Enabled: true, NextNumber: 1, Padding: 4, Prefix: "ORC-"},
			Defaults: QuoteDefaults{
				Currency:     "BRL",
				ValidityDays: 15,
				PaymentTerms: "50% na aprovação, 50% na entrega",
			},
		},
		Contract: ContractConfig{
			Template:  "professional",
			Numbering: NumberingConfig{Enabled: true, NextNumber: 1, Padding: 4, Prefix: "CTR-"},
			Defaults: ContractDefaults{
				Place:   "Fortaleza/CE",
				Clauses: DefaultClauses(),
			},
		},
	}
}

// DefaultPath é o caminho do config da suíte Emissor.
func DefaultPath() string {
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "Emissor", "config.json")
		}
	}
	base, err := os.UserConfigDir()
	if err != nil || base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "emissor", "config.json")
}

func MetadataDir(configPath string) string {
	if configPath == "" {
		configPath = DefaultPath()
	}
	return filepath.Join(filepath.Dir(configPath), "metadata")
}

// MetadataDirFor devolve a pasta de metadados de um tipo de documento.
func MetadataDirFor(configPath string, docType core.DocType) string {
	return filepath.Join(MetadataDir(configPath), docType.MetadataFolder())
}

func ReceiptsMetadataDir(configPath string) string {
	return MetadataDirFor(configPath, core.DocReceipt)
}

// DefaultOutputDir é a pasta base de saída da suíte (~/Documentos/Emissor).
func DefaultOutputDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "Emissor"
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(home, "Documents", "Emissor")
	}
	return filepath.Join(home, "Documentos", "Emissor")
}
