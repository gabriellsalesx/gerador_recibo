package pix

import (
	"bytes"
	"testing"
)

func TestCRC16KnownVector(t *testing.T) {
	if got := CRC16("123456789"); got != "29B1" {
		t.Fatalf("CRC16() = %s, want 29B1", got)
	}
}

func TestNormalizeName(t *testing.T) {
	got := NormalizeName("Joao da Costa")
	if got != "JOAO DA COSTA" {
		t.Fatalf("NormalizeName() = %q", got)
	}
}

func TestBuildPayloadIncludesFinalCRC(t *testing.T) {
	payload, err := BuildPayload(Payment{
		Key:          "contato@example.com",
		KeyType:      "email",
		ReceiverName: "Minha Empresa",
		ReceiverCity: "Fortaleza",
		AmountCents:  1000,
		Description:  "Teste",
		TxID:         "REC0001",
	})
	if err != nil {
		t.Fatalf("BuildPayload() returned error: %v", err)
	}
	if len(payload) < 8 || payload[len(payload)-8:len(payload)-4] != "6304" {
		t.Fatalf("payload does not include CRC field marker: %s", payload)
	}
	gotCRC := payload[len(payload)-4:]
	wantCRC := CRC16(payload[:len(payload)-4])
	if gotCRC != wantCRC {
		t.Fatalf("payload CRC = %s, want %s", gotCRC, wantCRC)
	}
}

func TestQRCodePNG(t *testing.T) {
	png, err := QRCodePNG("000201010212", 128)
	if err != nil {
		t.Fatalf("QRCodePNG() returned error: %v", err)
	}
	if !bytes.HasPrefix(png, []byte{0x89, 'P', 'N', 'G'}) {
		t.Fatal("QRCodePNG() did not return a PNG")
	}
}

func TestQRCodeTerminal(t *testing.T) {
	output, err := QRCodeTerminal("000201010212")
	if err != nil {
		t.Fatalf("QRCodeTerminal() returned error: %v", err)
	}
	if output == "" || !bytes.Contains([]byte(output), []byte("█")) {
		t.Fatal("QRCodeTerminal() did not render a terminal QR code")
	}
}
