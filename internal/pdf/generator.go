package pdf

import (
	"bytes"
	"fmt"
	"strings"

	"emissor/internal/core"
	"emissor/internal/receipt"

	"github.com/go-pdf/fpdf"
)

type Options struct {
	LogoImage      []byte
	SignatureImage []byte
}

func DetectImageType(b []byte) (string, bool) {
	switch {
	case len(b) >= 4 && b[0] == 0x89 && b[1] == 'P' && b[2] == 'N' && b[3] == 'G':
		return "PNG", true
	case len(b) >= 3 && b[0] == 0xFF && b[1] == 0xD8 && b[2] == 0xFF:
		return "JPG", true
	case len(b) >= 4 && string(b[0:4]) == "GIF8":
		return "GIF", true
	default:
		return "", false
	}
}

func drawImageFit(doc *fpdf.Fpdf, name, imgType string, img []byte, boxX, boxY, boxW, boxH float64, align string) float64 {
	info := doc.RegisterImageOptionsReader(name, fpdf.ImageOptions{ImageType: imgType}, bytes.NewReader(img))
	if info == nil {
		return boxY
	}
	iw, ih := info.Width(), info.Height()
	if iw <= 0 || ih <= 0 {
		return boxY
	}
	aspect := iw / ih
	h := boxH
	w := h * aspect
	if w > boxW {
		w = boxW
		h = w / aspect
	}
	x := boxX
	switch align {
	case "C":
		x = boxX + (boxW-w)/2
	case "R":
		x = boxX + (boxW - w)
	}
	doc.ImageOptions(name, x, boxY, w, h, false, fpdf.ImageOptions{ImageType: imgType}, 0, "")
	return boxY + h
}

func Generate(r receipt.Receipt, opts Options) ([]byte, error) {
	doc := fpdf.New("P", "mm", "A4", "")
	doc.SetMargins(18, 18, 18)
	doc.SetAutoPageBreak(true, 18)
	doc.AddPage()

	tr := doc.UnicodeTranslatorFromDescriptor("cp1252")
	text := func(s string) string { return tr(s) }

	doc.SetDrawColor(31, 41, 55)
	doc.SetFillColor(31, 41, 55)
	doc.Rect(0, 0, 210, 18, "F")
	doc.SetTextColor(255, 255, 255)
	doc.SetFont("Helvetica", "B", 18)
	doc.SetXY(18, 6)
	doc.CellFormat(100, 8, text("RECIBO"), "", 0, "L", false, 0, "")
	doc.SetFont("Helvetica", "", 10)
	doc.SetXY(140, 6)
	doc.CellFormat(52, 8, text("Número: "+r.Number), "", 0, "R", false, 0, "")

	doc.SetTextColor(17, 24, 39)
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
	sectionTitle(doc, text, "Emitente")
	partyBlock(doc, text, r.Issuer)

	doc.Ln(4)
	sectionTitle(doc, text, "Pagador")
	partyBlock(doc, text, r.Payer)

	doc.Ln(6)
	doc.SetFillColor(243, 244, 246)
	doc.SetFont("Helvetica", "B", 16)
	doc.CellFormat(174, 14, text("Valor: "+r.Amount.FormatBRL()), "1", 1, "C", true, 0, "")

	doc.Ln(5)
	doc.SetFont("Helvetica", "", 11)
	doc.MultiCell(174, 6, text(receiptText(r)), "", "L", false)

	doc.Ln(4)
	doc.SetFont("Helvetica", "", 10)
	doc.CellFormat(42, 6, text("Forma de pagamento:"), "", 0, "L", false, 0, "")
	doc.SetFont("Helvetica", "B", 10)
	doc.CellFormat(80, 6, text(r.PaymentMethod), "", 1, "L", false, 0, "")
	doc.SetFont("Helvetica", "", 10)
	doc.CellFormat(42, 6, text("Data do pagamento:"), "", 0, "L", false, 0, "")
	doc.SetFont("Helvetica", "B", 10)
	doc.CellFormat(80, 6, text(r.PaidAt.Format("02/01/2006")), "", 1, "L", false, 0, "")

	if r.Notes != "" {
		doc.Ln(4)
		sectionTitle(doc, text, "Observações")
		doc.SetFont("Helvetica", "", 10)
		doc.MultiCell(174, 5.5, text(r.Notes), "", "L", false)
	}

	if r.Pix != nil && r.Pix.Enabled {
		drawPix(doc, text, r)
	}

	if doc.GetY() < 230 {
		doc.SetY(230)
	} else {
		doc.Ln(8)
	}
	doc.SetFont("Helvetica", "", 10)
	local := strings.TrimSpace(strings.Join([]string{r.Issuer.City, r.Issuer.State}, " - "))
	if local == "-" {
		local = ""
	}
	if local != "" {
		local += ", "
	}
	doc.CellFormat(174, 6, text(local+r.IssuedAt.Format("02/01/2006")), "", 1, "R", false, 0, "")

	doc.Ln(12)
	lineY := doc.GetY()
	if len(opts.SignatureImage) > 0 {
		if imgType, ok := DetectImageType(opts.SignatureImage); ok {
			drawImageFit(doc, "signature", imgType, opts.SignatureImage, 60, lineY-11, 90, 10, "C")
		}
	}
	doc.Line(60, lineY, 150, lineY)
	doc.Ln(2)
	doc.CellFormat(174, 6, text(r.Issuer.Name), "", 1, "C", false, 0, "")

	doc.SetY(268)
	doc.SetFont("Helvetica", "", 8)
	doc.SetTextColor(107, 114, 128)
	footer := strings.TrimSpace(r.Issuer.Name)
	if r.Issuer.Document != "" {
		footer += " | " + r.Issuer.Document
	}
	doc.CellFormat(174, 5, text(footer), "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func sectionTitle(doc *fpdf.Fpdf, text func(string) string, title string) {
	doc.SetFont("Helvetica", "B", 11)
	doc.SetTextColor(31, 41, 55)
	doc.CellFormat(174, 7, text(title), "B", 1, "L", false, 0, "")
	doc.Ln(1)
}

func partyBlock(doc *fpdf.Fpdf, text func(string) string, p core.Party) {
	doc.SetFont("Helvetica", "B", 10)
	doc.CellFormat(174, 5.5, text(p.Name), "", 1, "L", false, 0, "")
	doc.SetFont("Helvetica", "", 9)
	lines := []string{}
	if p.TradeName != "" {
		lines = append(lines, "Nome fantasia: "+p.TradeName)
	}
	if p.Document != "" {
		lines = append(lines, "Documento: "+p.Document)
	}
	if p.Address != "" {
		lines = append(lines, "Endereço: "+p.Address)
	}
	location := strings.TrimSpace(strings.Join([]string{p.City, p.State}, " - "))
	if location != "-" && location != "" {
		lines = append(lines, "Cidade/UF: "+location)
	}
	if p.Phone != "" {
		lines = append(lines, "Telefone: "+p.Phone)
	}
	if p.Email != "" {
		lines = append(lines, "E-mail: "+p.Email)
	}
	for _, line := range lines {
		doc.CellFormat(174, 5, text(line), "", 1, "L", false, 0, "")
	}
}

func receiptText(r receipt.Receipt) string {
	value := fmt.Sprintf("%s (%s)", r.Amount.FormatBRL(), core.MoneyInWords(r.Amount))
	if strings.TrimSpace(r.Payer.Document) != "" {
		return fmt.Sprintf("Recebi de %s, inscrito(a) sob o documento %s, a importância de %s, referente a %s.",
			r.Payer.Name, r.Payer.Document, value, r.Description)
	}
	return fmt.Sprintf("Recebi de %s a importância de %s, referente a %s.", r.Payer.Name, value, r.Description)
}

func drawPix(doc *fpdf.Fpdf, text func(string) string, r receipt.Receipt) {
	doc.Ln(7)
	sectionTitle(doc, text, "Pagamento via Pix")
	doc.SetFont("Helvetica", "", 9)
	doc.CellFormat(32, 5, text("Valor:"), "", 0, "L", false, 0, "")
	doc.SetFont("Helvetica", "B", 9)
	doc.CellFormat(80, 5, text(r.Pix.Amount.FormatBRL()), "", 1, "L", false, 0, "")
	if r.Pix.ShowKeyText {
		doc.SetFont("Helvetica", "", 9)
		doc.CellFormat(32, 5, text("Chave Pix:"), "", 0, "L", false, 0, "")
		doc.CellFormat(120, 5, text(r.Pix.Key), "", 1, "L", false, 0, "")
	}

	y := doc.GetY() + 2
	if len(r.Pix.QRCodePNG) > 0 {
		doc.RegisterImageOptionsReader("pix-qrcode", fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(r.Pix.QRCodePNG))
		doc.ImageOptions("pix-qrcode", 18, y, 36, 36, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	}
	if r.Pix.ShowCopyPaste {
		doc.SetXY(60, y+8)
		doc.SetFont("Helvetica", "", 8)
		doc.MultiCell(132, 4, text("Pix Copia e Cola:\n"+r.Pix.CopyPaste), "1", "L", false)
	}
	doc.SetY(y + 40)
}
