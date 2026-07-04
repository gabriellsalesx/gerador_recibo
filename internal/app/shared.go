package app

import (
	"strings"

	"emissor/internal/config"
	"emissor/internal/core"
	"emissor/internal/pix"
)

// buildPix monta um Pix estático offline (Copia e Cola + QR Code) para o valor
// e número informados. Compartilhado por recibo e orçamento.
func buildPix(cfg config.Config, amount core.Money, number, fallbackDescription, pixFileName string) (*core.PixPayment, error) {
	receiverName := cfg.Pix.ReceiverName
	if receiverName == "" {
		receiverName = cfg.Issuer.Name
	}
	receiverCity := cfg.Pix.ReceiverCity
	if receiverCity == "" {
		receiverCity = cfg.Issuer.City
	}
	description := cfg.Pix.DefaultDescription
	if description == "" {
		description = fallbackDescription
	}
	txid := cfg.Pix.TxIDPrefix + number
	payment := pix.Payment{
		Key:          cfg.Issuer.PixKey,
		KeyType:      cfg.Issuer.PixKeyType,
		ReceiverName: receiverName,
		ReceiverCity: receiverCity,
		AmountCents:  amount.Amount,
		Description:  description,
		TxID:         txid,
	}
	copyPaste, err := pix.BuildPayload(payment)
	if err != nil {
		return nil, err
	}
	qr, err := pix.QRCodePNG(copyPaste, 256)
	if err != nil {
		return nil, err
	}
	return &core.PixPayment{
		Enabled:       true,
		Key:           payment.Key,
		KeyType:       payment.KeyType,
		ReceiverName:  pix.NormalizeName(receiverName),
		ReceiverCity:  pix.NormalizeCity(receiverCity),
		Amount:        amount,
		Description:   description,
		TxID:          pix.NormalizeTxID(txid),
		CopyPaste:     copyPaste,
		QRCodeFile:    pixFileName,
		ShowCopyPaste: cfg.Pix.ShowCopyPaste,
		ShowKeyText:   cfg.Pix.ShowKeyText,
		QRCodePNG:     qr,
	}, nil
}

// applyIssuerOverrides aplica sobrescritas de emitente vindas de argumentos de
// linha de comando à config carregada, sem persistir.
func applyIssuerOverrides(cfg *config.Config, name, document, city, state string) {
	if strings.TrimSpace(name) != "" {
		cfg.Issuer.Name = strings.TrimSpace(name)
	}
	if strings.TrimSpace(document) != "" {
		cfg.Issuer.Document = strings.TrimSpace(document)
	}
	if strings.TrimSpace(city) != "" {
		cfg.Issuer.City = strings.TrimSpace(city)
	}
	if strings.TrimSpace(state) != "" {
		cfg.Issuer.State = strings.TrimSpace(state)
	}
}
