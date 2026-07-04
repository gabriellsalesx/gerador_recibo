package pdf

import (
	"bytes"
	"strconv"
	"strings"

	"emissor/internal/contract"

	"github.com/go-pdf/fpdf"
)

// GenerateContract gera o PDF do contrato de prestação de serviço.
func GenerateContract(c contract.Contract, opts Options) ([]byte, error) {
	doc := fpdf.New("P", "mm", "A4", "")
	doc.SetMargins(18, 18, 18)
	doc.SetAutoPageBreak(true, 18)
	doc.AddPage()

	tr := doc.UnicodeTranslatorFromDescriptor("cp1252")
	text := func(s string) string { return tr(s) }

	drawDocHeader(doc, text, "CONTRATO DE PRESTAÇÃO DE SERVIÇOS", c.Number)

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

	// Qualificação das partes.
	sectionTitle(doc, text, "Partes")
	doc.SetFont("Helvetica", "B", 10)
	doc.CellFormat(174, 5.5, text("CONTRATANTE"), "", 1, "L", false, 0, "")
	partyBlock(doc, text, c.Contractor)
	doc.Ln(2)
	doc.SetFont("Helvetica", "B", 10)
	doc.CellFormat(174, 5.5, text("CONTRATADO"), "", 1, "L", false, 0, "")
	partyBlock(doc, text, c.Contracted)

	doc.Ln(4)
	sectionTitle(doc, text, "Cláusulas")

	// Distribui as cláusulas entre as páginas de forma equilibrada, reservando
	// espaço para local/data + assinaturas na última página (evita página final
	// quase em branco só com a assinatura).
	_, pageHeight := doc.GetPageSize()
	_, topMargin, _, bottomMargin := doc.GetMargins()
	pageBottom := pageHeight - bottomMargin
	startY := doc.GetY()
	const localDateH = 10.0
	const signatureBlockH = 40.0
	reserveEnd := localDateH + signatureBlockH

	heights := make([]float64, len(c.Clauses))
	for i := range c.Clauses {
		heights[i] = measureClauseHeight(doc, text, i, c.Clauses[i])
	}
	pages := planClausePages(heights, pageBottom-startY, pageBottom-topMargin, reserveEnd)

	idx := 0
	for g, size := range pages {
		if g > 0 {
			doc.AddPage()
		}
		for k := 0; k < size; k++ {
			drawClause(doc, text, idx, c.Clauses[idx])
			idx++
		}
	}

	// Local e data.
	doc.Ln(2)
	doc.SetFont("Helvetica", "", 10)
	local := strings.TrimSpace(c.Place)
	dateStr := c.SignedDate.Format("02/01/2006")
	line := dateStr
	if local != "" {
		line = local + ", " + dateStr
	}
	doc.CellFormat(174, 6, text(line), "", 1, "R", false, 0, "")

	// Assinaturas das duas partes (o espaço já foi reservado acima).
	doc.Ln(12)
	drawSignatureLines(doc, text, c.Contractor.Name, c.Contracted.Name)

	// Testemunhas.
	if len(c.Witnesses) > 0 {
		doc.Ln(8)
		sectionTitle(doc, text, "Testemunhas")
		doc.SetFont("Helvetica", "", 9)
		for _, w := range c.Witnesses {
			line := w.Name
			if strings.TrimSpace(w.Document) != "" {
				line += " — " + w.Document
			}
			doc.CellFormat(174, 6, text(line), "", 1, "L", false, 0, "")
		}
	}

	drawDocFooter(doc, text, c.Contracted, "")

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func clauseHeading(index int, clause contract.Clause) string {
	heading := "CLÁUSULA " + ordinalFeminine(index+1)
	if t := strings.TrimSpace(clause.Title); t != "" {
		heading += " — " + t
	}
	return heading
}

const clauseLineHeight = 5.5

// measureClauseHeight estima a altura de uma cláusula (título + corpo + espaço).
// Usa SplitLines sobre o texto já traduzido (cp1252), adequado à fonte core.
func measureClauseHeight(doc *fpdf.Fpdf, text func(string) string, index int, clause contract.Clause) float64 {
	doc.SetFont("Helvetica", "B", 10)
	lines := len(doc.SplitLines([]byte(text(clauseHeading(index, clause))), 174))
	if lines < 1 {
		lines = 1
	}
	if body := strings.TrimSpace(clause.Body); body != "" {
		doc.SetFont("Helvetica", "", 10)
		bl := len(doc.SplitLines([]byte(text(body)), 174))
		if bl < 1 {
			bl = 1
		}
		lines += bl
	}
	return float64(lines)*clauseLineHeight + 2
}

func drawClause(doc *fpdf.Fpdf, text func(string) string, index int, clause contract.Clause) {
	doc.SetFont("Helvetica", "B", 10)
	doc.MultiCell(174, clauseLineHeight, text(clauseHeading(index, clause)), "", "L", false)
	if strings.TrimSpace(clause.Body) != "" {
		doc.SetFont("Helvetica", "", 10)
		doc.MultiCell(174, clauseLineHeight, text(clause.Body), "", "J", false)
	}
	doc.Ln(2)
}

// splitCounts divide n cláusulas em p grupos contíguos o mais iguais possível.
func splitCounts(n, p int) []int {
	if p < 1 {
		p = 1
	}
	sizes := make([]int, p)
	base := n / p
	rem := n % p
	for i := range sizes {
		sizes[i] = base
		if i < rem {
			sizes[i]++
		}
	}
	return sizes
}

// planClausePages escolhe o menor número de páginas em que a divisão igual por
// contagem cabe, reservando reserveEnd na última página para as assinaturas.
func planClausePages(heights []float64, avail0, availN, reserveEnd float64) []int {
	n := len(heights)
	if n == 0 {
		return nil
	}
	for p := 1; p <= n; p++ {
		sizes := splitCounts(n, p)
		if clausesFit(heights, sizes, avail0, availN, reserveEnd) {
			return sizes
		}
	}
	return splitCounts(n, n)
}

func clausesFit(heights []float64, sizes []int, avail0, availN, reserveEnd float64) bool {
	p := len(sizes)
	idx := 0
	for g, size := range sizes {
		sum := 0.0
		for k := 0; k < size; k++ {
			sum += heights[idx]
			idx++
		}
		capacity := availN
		if g == 0 {
			capacity = avail0
		}
		if g == p-1 {
			capacity -= reserveEnd
		}
		if sum > capacity {
			return false
		}
	}
	return true
}

func drawSignatureLines(doc *fpdf.Fpdf, text func(string) string, left, right string) {
	y := doc.GetY()
	doc.Line(24, y, 96, y)
	doc.Line(114, y, 186, y)
	doc.Ln(2)
	doc.SetFont("Helvetica", "", 9)
	doc.CellFormat(90, 5, text(truncateName(left)), "", 0, "C", false, 0, "")
	doc.CellFormat(6, 5, "", "", 0, "C", false, 0, "")
	doc.CellFormat(90, 5, text(truncateName(right)), "", 1, "C", false, 0, "")
	doc.SetFont("Helvetica", "", 8)
	doc.SetTextColor(107, 114, 128)
	doc.CellFormat(90, 4, text("CONTRATANTE"), "", 0, "C", false, 0, "")
	doc.CellFormat(6, 4, "", "", 0, "C", false, 0, "")
	doc.CellFormat(90, 4, text("CONTRATADO"), "", 1, "C", false, 0, "")
	doc.SetTextColor(17, 24, 39)
}

func truncateName(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 40 {
		return s[:37] + "..."
	}
	return s
}

func ordinalFeminine(n int) string {
	names := []string{
		"PRIMEIRA", "SEGUNDA", "TERCEIRA", "QUARTA", "QUINTA", "SEXTA",
		"SÉTIMA", "OITAVA", "NONA", "DÉCIMA", "DÉCIMA PRIMEIRA", "DÉCIMA SEGUNDA",
		"DÉCIMA TERCEIRA", "DÉCIMA QUARTA", "DÉCIMA QUINTA",
	}
	if n >= 1 && n <= len(names) {
		return names[n-1]
	}
	return strconv.Itoa(n) + "ª"
}
