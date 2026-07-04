package contract

import (
	"time"

	"emissor/internal/core"
)

// Clause é uma cláusula do contrato (título + corpo).
type Clause struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// Contract é o documento de contrato de prestação de serviço.
type Contract struct {
	ID           string       `json:"id"`
	Number       string       `json:"number"`
	CreatedAt    time.Time    `json:"created_at"`
	SignedDate   time.Time    `json:"signed_date"`
	PDFFile      string       `json:"pdf_file,omitempty"`
	PDFPath      string       `json:"pdf_path,omitempty"`
	Contractor   core.Party   `json:"contractor"` // contratante (cliente)
	Contracted   core.Party   `json:"contracted"` // contratado (emitente)
	Object       string       `json:"object"`
	Amount       core.Money   `json:"amount"`
	PaymentTerms string       `json:"payment_terms,omitempty"`
	Term         string       `json:"term"`
	Clauses      []Clause     `json:"clauses"`
	Place        string       `json:"place"`
	Witnesses    []core.Party `json:"witnesses,omitempty"`
	Locale       string       `json:"locale,omitempty"`
}

// DocType implementa a identificação do documento na suíte.
func (Contract) DocType() core.DocType { return core.DocContract }
