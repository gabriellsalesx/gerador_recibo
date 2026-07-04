package pdf

import (
	"bytes"
	"strings"

	"emissor/internal/core"
	"emissor/internal/quote"

	"github.com/go-pdf/fpdf"
)

// drawDocHeader desenha o cabeçalho padrão da suíte (barra escura com título e
// número). Reutilizado por orçamento e contrato.
func drawDocHeader(doc *fpdf.Fpdf, text func(string) string, title, number string) {
	doc.SetDrawColor(31, 41, 55)
	doc.SetFillColor(31, 41, 55)
	doc.Rect(0, 0, 210, 18, "F")
	doc.SetTextColor(255, 255, 255)
	doc.SetFont("Helvetica", "B", 16)
	doc.SetXY(18, 6)
	doc.CellFormat(120, 8, text(title), "", 0, "L", false, 0, "")
	doc.SetFont("Helvetica", "", 10)
	doc.SetXY(140, 6)
	doc.CellFormat(52, 8, text("Número: "+number), "", 0, "R", false, 0, "")
	doc.SetTextColor(17, 24, 39)
}

func drawDocFooter(doc *fpdf.Fpdf, text func(string) string, issuer core.Party, note string) {
	doc.SetY(272)
	doc.SetFont("Helvetica", "", 8)
	doc.SetTextColor(107, 114, 128)
	footer := strings.TrimSpace(issuer.Name)
	if issuer.Document != "" {
		footer += " | " + issuer.Document
	}
	if note != "" {
		footer += " | " + note
	}
	doc.CellFormat(174, 5, text(footer), "", 1, "C", false, 0, "")
}

// GenerateQuote gera o PDF do orçamento.
func GenerateQuote(q quote.Quote, opts Options) ([]byte, error) {
	doc := fpdf.New("P", "mm", "A4", "")
	doc.SetMargins(18, 18, 18)
	doc.SetAutoPageBreak(true, 18)
	doc.AddPage()

	tr := doc.UnicodeTranslatorFromDescriptor("cp1252")
	text := func(s string) string { return tr(s) }

	drawDocHeader(doc, text, "ORÇAMENTO", q.Number)

	contentTop := 28.0
	if len(opts.LogoImage) > 0 {
		if imgType, ok := DetectImageType(opts.LogoImage); ok {
			bottom := drawImageFit(doc, "logo", imgType, opts.LogoImage, 150, 22, 42, 20, "R")
			if bottom+3 > contentTop {
				contentTop = bottom + 3
			}
		}
	}
	doc.SetY(contentTop)

	// Datas em destaque.
	doc.SetFont("Helvetica", "", 10)
	doc.CellFormat(87, 6, text("Emissão: "+q.IssuedAt.Format("02/01/2006")), "", 0, "L", false, 0, "")
	doc.SetFont("Helvetica", "B", 10)
	doc.CellFormat(87, 6, text("Válido até: "+q.ValidUntil.Format("02/01/2006")), "", 1, "R", false, 0, "")
	doc.Ln(3)

	sectionTitle(doc, text, "Prestador")
	partyBlock(doc, text, q.Issuer)
	doc.Ln(3)
	sectionTitle(doc, text, "Cliente")
	partyBlock(doc, text, q.Client)

	doc.Ln(4)
	sectionTitle(doc, text, "Itens")
	drawQuoteItems(doc, text, q)

	doc.Ln(2)
	drawQuoteTotals(doc, text, q)

	if strings.TrimSpace(q.PaymentTerms) != "" {
		doc.Ln(4)
		drawLabeledField(doc, text, "Condições de pagamento", q.PaymentTerms)
	}
	if strings.TrimSpace(q.Deadline) != "" {
		drawLabeledField(doc, text, "Prazo de execução", q.Deadline)
	}
	if strings.TrimSpace(q.Notes) != "" {
		doc.Ln(2)
		sectionTitle(doc, text, "Observações")
		doc.SetFont("Helvetica", "", 10)
		doc.MultiCell(174, 5.5, text(q.Notes), "", "L", false)
	}

	if q.Pix != nil && q.Pix.Enabled {
		drawQuotePix(doc, text, q)
	}

	doc.Ln(6)
	doc.SetFont("Helvetica", "I", 9)
	doc.SetTextColor(107, 114, 128)
	doc.MultiCell(174, 5, text("Este documento é uma proposta comercial e não constitui comprovante de pagamento."), "", "L", false)
	doc.SetTextColor(17, 24, 39)

	drawDocFooter(doc, text, q.Issuer, "Documento gerado por Emissor")

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drawQuoteItems(doc *fpdf.Fpdf, text func(string) string, q quote.Quote) {
	// Larguras das colunas (soma = 174).
	const (
		wDesc  = 82.0
		wQty   = 16.0
		wUnit  = 14.0
		wPrice = 31.0
		wTotal = 31.0
	)
	doc.SetFont("Helvetica", "B", 9)
	doc.SetFillColor(243, 244, 246)
	doc.CellFormat(wDesc, 7, text("Descrição"), "1", 0, "L", true, 0, "")
	doc.CellFormat(wQty, 7, text("Qtd"), "1", 0, "C", true, 0, "")
	doc.CellFormat(wUnit, 7, text("Un"), "1", 0, "C", true, 0, "")
	doc.CellFormat(wPrice, 7, text("V. Unit."), "1", 0, "R", true, 0, "")
	doc.CellFormat(wTotal, 7, text("Total"), "1", 1, "R", true, 0, "")

	doc.SetFont("Helvetica", "", 9)
	for _, item := range q.Items {
		desc := item.Description
		if len(desc) > 52 {
			desc = desc[:49] + "..."
		}
		doc.CellFormat(wDesc, 6.5, text(desc), "1", 0, "L", false, 0, "")
		doc.CellFormat(wQty, 6.5, text(quote.FormatQuantity(item.QuantityMilli)), "1", 0, "C", false, 0, "")
		doc.CellFormat(wUnit, 6.5, text(defaultUnit(item.Unit)), "1", 0, "C", false, 0, "")
		doc.CellFormat(wPrice, 6.5, text(item.UnitPrice.FormatBRL()), "1", 0, "R", false, 0, "")
		doc.CellFormat(wTotal, 6.5, text(item.Total.FormatBRL()), "1", 1, "R", false, 0, "")
	}
}

func defaultUnit(u string) string {
	if strings.TrimSpace(u) == "" {
		return "un"
	}
	return u
}

// drawLabeledField imprime "Rótulo: " (largura ajustada ao texto) seguido do
// valor em negrito, garantindo espaço entre o rótulo e o conteúdo.
func drawLabeledField(doc *fpdf.Fpdf, text func(string) string, label, value string) {
	doc.SetFont("Helvetica", "", 10)
	labelStr := text(label + ": ")
	w := doc.GetStringWidth(labelStr) + 1
	doc.CellFormat(w, 6, labelStr, "", 0, "L", false, 0, "")
	doc.SetFont("Helvetica", "B", 10)
	doc.MultiCell(174-w, 6, text(value), "", "L", false)
}

func drawQuoteTotals(doc *fpdf.Fpdf, text func(string) string, q quote.Quote) {
	row := func(label, value string, bold bool) {
		doc.CellFormat(112, 6, "", "", 0, "L", false, 0, "")
		doc.SetFont("Helvetica", "", 10)
		doc.CellFormat(31, 6, text(label), "", 0, "R", false, 0, "")
		if bold {
			doc.SetFont("Helvetica", "B", 10)
		}
		doc.CellFormat(31, 6, text(value), "", 1, "R", false, 0, "")
	}
	row("Subtotal:", q.Subtotal.FormatBRL(), false)
	if q.Discount.Amount > 0 {
		row("Desconto:", "-"+q.Discount.FormatBRL(), false)
	}
	if q.Surcharge.Amount > 0 {
		row("Acréscimo:", q.Surcharge.FormatBRL(), false)
	}
	row("Total:", q.Total.FormatBRL(), true)
	doc.SetFont("Helvetica", "I", 9)
	doc.SetTextColor(107, 114, 128)
	doc.CellFormat(174, 5, text("("+core.MoneyInWords(q.Total)+")"), "", 1, "R", false, 0, "")
	doc.SetTextColor(17, 24, 39)
}

func drawQuotePix(doc *fpdf.Fpdf, text func(string) string, q quote.Quote) {
	doc.Ln(5)
	sectionTitle(doc, text, "Pagamento via Pix")
	doc.SetFont("Helvetica", "", 9)
	doc.CellFormat(32, 5, text("Valor:"), "", 0, "L", false, 0, "")
	doc.SetFont("Helvetica", "B", 9)
	doc.CellFormat(80, 5, text(q.Pix.Amount.FormatBRL()), "", 1, "L", false, 0, "")
	if q.Pix.ShowKeyText {
		doc.SetFont("Helvetica", "", 9)
		doc.CellFormat(32, 5, text("Chave Pix:"), "", 0, "L", false, 0, "")
		doc.CellFormat(120, 5, text(q.Pix.Key), "", 1, "L", false, 0, "")
	}
	y := doc.GetY() + 2
	if len(q.Pix.QRCodePNG) > 0 {
		doc.RegisterImageOptionsReader("quote-pix-qrcode", fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(q.Pix.QRCodePNG))
		doc.ImageOptions("quote-pix-qrcode", 18, y, 34, 34, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	}
	if q.Pix.ShowCopyPaste {
		doc.SetXY(56, y+6)
		doc.SetFont("Helvetica", "", 8)
		doc.MultiCell(136, 4, text("Pix Copia e Cola:\n"+q.Pix.CopyPaste), "1", "L", false)
	}
	doc.SetY(y + 38)
}
