package quote

import (
	"time"

	"emissor/internal/core"
)

// LineItem é um item da tabela do orçamento. A quantidade é armazenada em
// milésimos (QuantityMilli) para suportar frações (ex.: 1,5h) sem usar float.
type LineItem struct {
	Description   string     `json:"description"`
	QuantityMilli int64      `json:"quantity_milli"`
	Unit          string     `json:"unit,omitempty"`
	UnitPrice     core.Money `json:"unit_price"`
	Total         core.Money `json:"total"`
}

// Quote é o documento de orçamento (proposta comercial).
type Quote struct {
	ID           string           `json:"id"`
	Number       string           `json:"number"`
	CreatedAt    time.Time        `json:"created_at"`
	IssuedAt     time.Time        `json:"issued_at"`
	ValidUntil   time.Time        `json:"valid_until"`
	PDFFile      string           `json:"pdf_file,omitempty"`
	PDFPath      string           `json:"pdf_path,omitempty"`
	Issuer       core.Party       `json:"issuer"` // prestador
	Client       core.Party       `json:"client"`
	Items        []LineItem       `json:"items"`
	Subtotal     core.Money       `json:"subtotal"`
	Discount     core.Money       `json:"discount"`
	Surcharge    core.Money       `json:"surcharge"`
	Total        core.Money       `json:"total"`
	PaymentTerms string           `json:"payment_terms,omitempty"`
	Deadline     string           `json:"deadline,omitempty"`
	Notes        string           `json:"notes,omitempty"`
	Locale       string           `json:"locale,omitempty"`
	Pix          *core.PixPayment `json:"pix,omitempty"`
}

// DocType implementa a identificação do documento na suíte.
func (Quote) DocType() core.DocType { return core.DocQuote }
