package core

import (
	"testing"
	"time"
)

func TestParseMoney(t *testing.T) {
	tests := map[string]int64{
		"250":      25000,
		"250,00":   25000,
		"250.00":   25000,
		"1.250,00": 125000,
		"1250.00":  125000,
		"0,01":     1,
	}
	for input, want := range tests {
		got, err := ParseMoney(input)
		if err != nil {
			t.Fatalf("ParseMoney(%q) returned error: %v", input, err)
		}
		if got.Amount != want {
			t.Fatalf("ParseMoney(%q) = %d, want %d", input, got.Amount, want)
		}
	}
}

func TestFormatBRL(t *testing.T) {
	got := NewMoney(125099, "BRL").FormatBRL()
	if got != "R$ 1.250,99" {
		t.Fatalf("FormatBRL() = %q", got)
	}
}

func TestMoneyInWords(t *testing.T) {
	got := MoneyInWords(NewMoney(25000, "BRL"))
	if got != "duzentos e cinquenta reais" {
		t.Fatalf("MoneyInWords() = %q", got)
	}
}

func TestMoneyInWordsMillions(t *testing.T) {
	cases := map[int64]string{
		100_000_000: "um milhão de reais",
		200_000_000: "dois milhões de reais",
		150_000_000: "um milhão e quinhentos mil reais",
		100_000_050: "um milhão de reais e cinquenta centavos",
		300_005_000: "três milhões e cinquenta reais",
	}
	for cents, want := range cases {
		if got := MoneyInWords(NewMoney(cents, "BRL")); got != want {
			t.Fatalf("MoneyInWords(%d) = %q, want %q", cents, got, want)
		}
	}
}

func TestParseDate(t *testing.T) {
	now := time.Date(2026, 7, 3, 15, 30, 0, 0, time.UTC)
	tests := map[string]string{
		"hoje":       "2026-07-03",
		"ontem":      "2026-07-02",
		"03/07/2026": "2026-07-03",
		"2026-07-03": "2026-07-03",
	}
	for input, want := range tests {
		got, err := ParseDate(input, now)
		if err != nil {
			t.Fatalf("ParseDate(%q) returned error: %v", input, err)
		}
		if got.Format("2006-01-02") != want {
			t.Fatalf("ParseDate(%q) = %s, want %s", input, got.Format("2006-01-02"), want)
		}
	}
}

func TestParseValidityDays(t *testing.T) {
	issued := time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC)
	got, err := ParseValidity("15 dias", issued)
	if err != nil {
		t.Fatalf("ParseValidity() returned error: %v", err)
	}
	if got.Format("2006-01-02") != "2026-07-18" {
		t.Fatalf("ParseValidity(15 dias) = %s, want 2026-07-18", got.Format("2006-01-02"))
	}
}
