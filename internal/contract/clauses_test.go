package contract

import (
	"strings"
	"testing"

	"emissor/internal/core"
)

func TestBuildClausesFillsKnownTitles(t *testing.T) {
	data := ClauseData{
		Object:       "Desenvolvimento de site",
		Amount:       core.NewMoney(240000, "BRL"),
		PaymentTerms: "50% na assinatura",
		Term:         "30 dias",
		Place:        "Fortaleza/CE",
	}
	clauses := BuildClauses([]string{"OBJETO", "PAGAMENTO", "FORO"}, data)
	if len(clauses) != 3 {
		t.Fatalf("BuildClauses returned %d clauses, want 3", len(clauses))
	}
	if !strings.Contains(clauses[0].Body, "Desenvolvimento de site") {
		t.Fatalf("OBJETO body missing object: %q", clauses[0].Body)
	}
	if !strings.Contains(clauses[1].Body, "R$ 2.400,00") {
		t.Fatalf("PAGAMENTO body missing amount: %q", clauses[1].Body)
	}
	if !strings.Contains(clauses[2].Body, "Fortaleza/CE") {
		t.Fatalf("FORO body missing place: %q", clauses[2].Body)
	}
}
