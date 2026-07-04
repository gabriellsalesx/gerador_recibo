package pix

import (
	"fmt"
	"strings"
)

type Payment struct {
	Key          string
	KeyType      string
	ReceiverName string
	ReceiverCity string
	AmountCents  int64
	Description  string
	TxID         string
}

func BuildPayload(p Payment) (string, error) {
	if err := Validate(p); err != nil {
		return "", err
	}

	merchantInfo := tlv("00", "br.gov.bcb.pix") + tlv("01", strings.TrimSpace(p.Key))
	if desc := normalizeText(p.Description, 72); desc != "" {
		merchantInfo += tlv("02", desc)
	}

	payload := ""
	payload += tlv("00", "01")
	payload += tlv("26", merchantInfo)
	payload += tlv("52", "0000")
	payload += tlv("53", "986")
	payload += tlv("54", amountString(p.AmountCents))
	payload += tlv("58", "BR")
	payload += tlv("59", NormalizeName(p.ReceiverName))
	payload += tlv("60", NormalizeCity(p.ReceiverCity))
	payload += tlv("62", tlv("05", NormalizeTxID(p.TxID)))
	payload += "6304"
	return payload + CRC16(payload), nil
}

func tlv(id, value string) string {
	return fmt.Sprintf("%s%02d%s", id, len(value), value)
}

func amountString(cents int64) string {
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}
