package receipt

import (
	"time"

	"emissor/internal/core"
)

// Receipt é o documento de recibo (comprovante de pagamento recebido).
type Receipt struct {
	ID            string           `json:"id"`
	Number        string           `json:"number"`
	CreatedAt     time.Time        `json:"created_at"`
	IssuedAt      time.Time        `json:"issued_at"`
	PaidAt        time.Time        `json:"paid_at"`
	PDFFile       string           `json:"pdf_file,omitempty"`
	PDFPath       string           `json:"pdf_path,omitempty"`
	Issuer        core.Party       `json:"issuer"`
	Payer         core.Party       `json:"payer"`
	Amount        core.Money       `json:"amount"`
	Description   string           `json:"description"`
	PaymentMethod string           `json:"payment_method"`
	TransactionID string           `json:"transaction_id,omitempty"`
	Notes         string           `json:"notes,omitempty"`
	Locale        string           `json:"locale,omitempty"`
	Pix           *core.PixPayment `json:"pix,omitempty"`
}

// DocType implementa a identificação do documento na suíte.
func (Receipt) DocType() core.DocType { return core.DocReceipt }
