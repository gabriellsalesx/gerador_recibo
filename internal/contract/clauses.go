package contract

import (
	"strings"

	"emissor/internal/core"
)

// ClauseData reúne os dados usados para preencher os modelos de cláusula.
type ClauseData struct {
	Object         string
	Amount         core.Money
	PaymentTerms   string
	Term           string
	Place          string
	ContractorName string
	ContractedName string
}

// BuildClauses monta as cláusulas a partir dos títulos escolhidos, preenchendo
// o corpo padrão de cada título conhecido. Títulos desconhecidos recebem um
// corpo em branco para o usuário editar depois.
func BuildClauses(titles []string, data ClauseData) []Clause {
	clauses := make([]Clause, 0, len(titles))
	for _, title := range titles {
		clauses = append(clauses, Clause{
			Title: strings.ToUpper(strings.TrimSpace(title)),
			Body:  DefaultClauseBody(title, data),
		})
	}
	return clauses
}

// DefaultClauseBody devolve o texto-base de uma cláusula conhecida.
func DefaultClauseBody(title string, data ClauseData) string {
	amount := ""
	if data.Amount.Amount > 0 {
		amount = data.Amount.FormatBRL()
	}
	switch strings.ToUpper(strings.TrimSpace(title)) {
	case "OBJETO":
		return "O presente contrato tem por objeto a prestação, pelo CONTRATADO ao CONTRATANTE, dos seguintes serviços: " + fallback(data.Object, "(descrever o objeto)") + "."
	case "PAGAMENTO":
		body := "Pelos serviços prestados, o CONTRATANTE pagará ao CONTRATADO"
		if amount != "" {
			body += " o valor de " + amount
		} else {
			body += " o valor acordado entre as partes"
		}
		if strings.TrimSpace(data.PaymentTerms) != "" {
			body += ", nas seguintes condições: " + data.PaymentTerms
		}
		return body + "."
	case "PRAZO":
		return "O presente contrato vigorará pelo prazo de " + fallback(data.Term, "(definir prazo/vigência)") + ", podendo ser prorrogado mediante acordo entre as partes."
	case "OBRIGAÇÕES":
		return "O CONTRATADO obriga-se a executar os serviços com zelo e dentro do prazo ajustado. O CONTRATANTE obriga-se a fornecer as informações necessárias e a efetuar o pagamento na forma combinada."
	case "RESCISÃO":
		return "O presente contrato poderá ser rescindido por qualquer das partes, mediante comunicação prévia, respondendo a parte que der causa por eventuais perdas e danos."
	case "FORO":
		return "Fica eleito o foro da comarca de " + fallback(data.Place, "(cidade/UF)") + " para dirimir quaisquer dúvidas oriundas do presente contrato."
	default:
		return ""
	}
}

func fallback(value, def string) string {
	if strings.TrimSpace(value) == "" {
		return def
	}
	return value
}
