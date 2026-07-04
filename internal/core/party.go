package core

// DocType identifica o tipo de documento da suíte. É usado por storage e pelo
// gerador de PDF para despachar o comportamento correto.
type DocType string

const (
	DocReceipt  DocType = "receipt"
	DocQuote    DocType = "quote"
	DocContract DocType = "contract"
)

// Folder devolve o nome da pasta de saída para o tipo de documento.
func (d DocType) Folder() string {
	switch d {
	case DocQuote:
		return "Orcamentos"
	case DocContract:
		return "Contratos"
	default:
		return "Recibos"
	}
}

// FilePrefix devolve o prefixo do nome de arquivo para o tipo de documento.
func (d DocType) FilePrefix() string {
	switch d {
	case DocQuote:
		return "orcamento"
	case DocContract:
		return "contrato"
	default:
		return "recibo"
	}
}

// MetadataFolder devolve o nome da subpasta de metadados para o tipo.
func (d DocType) MetadataFolder() string {
	switch d {
	case DocQuote:
		return "orcamentos"
	case DocContract:
		return "contratos"
	default:
		return "receipts"
	}
}

// Party representa uma parte do documento (emitente, pagador, cliente,
// contratante ou contratado). É compartilhado pelos três tipos de documento.
type Party struct {
	Type         string `json:"type,omitempty"`
	Name         string `json:"name"`
	TradeName    string `json:"trade_name,omitempty"`
	DocumentType string `json:"document_type,omitempty"`
	Document     string `json:"document,omitempty"`
	Email        string `json:"email,omitempty"`
	Phone        string `json:"phone,omitempty"`
	Address      string `json:"address,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
	Country      string `json:"country,omitempty"`
	PixKey       string `json:"pix_key,omitempty"`
	PixKeyType   string `json:"pix_key_type,omitempty"`
}

// PixPayment carrega os dados do Pix estático offline, compartilhado por recibo
// e (opcionalmente) orçamento.
type PixPayment struct {
	Enabled       bool   `json:"enabled"`
	Key           string `json:"key"`
	KeyType       string `json:"key_type,omitempty"`
	ReceiverName  string `json:"receiver_name"`
	ReceiverCity  string `json:"receiver_city"`
	Amount        Money  `json:"amount"`
	Description   string `json:"description,omitempty"`
	TxID          string `json:"txid"`
	CopyPaste     string `json:"copy_paste"`
	QRCodeFile    string `json:"qr_code_file,omitempty"`
	ShowCopyPaste bool   `json:"show_copy_paste"`
	ShowKeyText   bool   `json:"show_key_text"`
	QRCodePNG     []byte `json:"-"`
}
