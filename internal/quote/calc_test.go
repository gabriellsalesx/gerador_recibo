package quote

import (
	"testing"

	"emissor/internal/core"
)

func TestRecalculate(t *testing.T) {
	q := Quote{
		Items: []LineItem{
			NewLineItem("Landing page", 1000, "un", core.NewMoney(180000, "BRL")),
			NewLineItem("Hospedagem 12 meses", 1000, "un", core.NewMoney(60000, "BRL")),
		},
		Discount: core.NewMoney(0, "BRL"),
	}
	Recalculate(&q)
	if q.Subtotal.Amount != 240000 {
		t.Fatalf("Subtotal = %d, want 240000", q.Subtotal.Amount)
	}
	if q.Total.Amount != 240000 {
		t.Fatalf("Total = %d, want 240000", q.Total.Amount)
	}
}

func TestRecalculateFractionalQuantityAndDiscount(t *testing.T) {
	q := Quote{
		Items: []LineItem{
			NewLineItem("Consultoria", 1500, "h", core.NewMoney(20000, "BRL")), // 1,5h * 200,00 = 300,00
		},
		Discount: core.NewMoney(5000, "BRL"), // 50,00
	}
	Recalculate(&q)
	if q.Items[0].Total.Amount != 30000 {
		t.Fatalf("item total = %d, want 30000", q.Items[0].Total.Amount)
	}
	if q.Total.Amount != 25000 {
		t.Fatalf("Total = %d, want 25000", q.Total.Amount)
	}
}

func TestParseQuantity(t *testing.T) {
	cases := map[string]int64{"": 1000, "2": 2000, "1,5": 1500, "2.5": 2500}
	for in, want := range cases {
		got, err := ParseQuantity(in)
		if err != nil {
			t.Fatalf("ParseQuantity(%q) error: %v", in, err)
		}
		if got != want {
			t.Fatalf("ParseQuantity(%q) = %d, want %d", in, got, want)
		}
	}
}
