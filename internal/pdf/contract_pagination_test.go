package pdf

import (
	"reflect"
	"testing"
)

func TestSplitCounts(t *testing.T) {
	if got := splitCounts(8, 2); !reflect.DeepEqual(got, []int{4, 4}) {
		t.Fatalf("splitCounts(8,2) = %v, want [4 4]", got)
	}
	if got := splitCounts(9, 2); !reflect.DeepEqual(got, []int{5, 4}) {
		t.Fatalf("splitCounts(9,2) = %v, want [5 4]", got)
	}
}

func TestPlanClausePagesBalancesAndAvoidsLonelySignature(t *testing.T) {
	heights := make([]float64, 8)
	for i := range heights {
		heights[i] = 30
	}
	// Tudo caberia em altura, mas a reserva da assinatura na última página
	// impede uma página só: deve dividir 4/4.
	got := planClausePages(heights, 200, 200, 44)
	if !reflect.DeepEqual(got, []int{4, 4}) {
		t.Fatalf("planClausePages = %v, want [4 4]", got)
	}
}

func TestPlanClausePagesSinglePageWhenItFits(t *testing.T) {
	heights := make([]float64, 4)
	for i := range heights {
		heights[i] = 10
	}
	got := planClausePages(heights, 250, 250, 44)
	if !reflect.DeepEqual(got, []int{4}) {
		t.Fatalf("planClausePages = %v, want [4]", got)
	}
}
