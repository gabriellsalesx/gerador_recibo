package pix

import (
	"errors"

	qrcode "github.com/skip2/go-qrcode"
)

func QRCodePNG(payload string, size int) ([]byte, error) {
	if payload == "" {
		return nil, errors.New("payload Pix vazio")
	}
	if size <= 0 {
		size = 256
	}
	return qrcode.Encode(payload, qrcode.Medium, size)
}

func QRCodeTerminal(payload string) (string, error) {
	if payload == "" {
		return "", errors.New("payload Pix vazio")
	}
	qr, err := qrcode.New(payload, qrcode.Medium)
	if err != nil {
		return "", err
	}
	return qr.ToSmallString(false), nil
}
