package pdf

import (
	"bytes"
	"testing"
	"time"

	"emissor/internal/contract"
	"emissor/internal/core"
)

func TestGenerateContractPDF(t *testing.T) {
	data := contract.ClauseData{
		Object:       "Desenvolvimento de site institucional",
		Amount:       core.NewMoney(240000, "BRL"),
		PaymentTerms: "50% na assinatura, 50% na entrega",
		Term:         "30 dias",
		Place:        "Fortaleza/CE",
	}
	c := contract.Contract{
		ID:           "0001",
		Number:       "CTR-0001",
		CreatedAt:    time.Date(2026, 7, 3, 17, 0, 0, 0, time.UTC),
		SignedDate:   time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC),
		Contractor:   core.Party{Name: "Empresa XPTO", Document: "00.000.000/0001-00"},
		Contracted:   core.Party{Name: "Minha Empresa", Document: "11.111.111/0001-11"},
		Object:       data.Object,
		Amount:       data.Amount,
		PaymentTerms: data.PaymentTerms,
		Term:         data.Term,
		Place:        data.Place,
		Clauses:      contract.BuildClauses([]string{"OBJETO", "PAGAMENTO", "PRAZO", "FORO"}, data),
	}
	out, err := GenerateContract(c, Options{})
	if err != nil {
		t.Fatalf("GenerateContract() returned error: %v", err)
	}
	if !bytes.HasPrefix(out, []byte("%PDF")) {
		t.Fatal("GenerateContract() did not return PDF bytes")
	}
}
