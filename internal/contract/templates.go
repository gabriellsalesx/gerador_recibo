package contract

import "strings"

// clauseSpec é a especificação de uma cláusula em um modelo: título + corpo com
// marcadores ({object}, {amount}, {payment_terms}, {term}, {place},
// {contractor}, {contracted}) preenchidos a partir de ClauseData.
type clauseSpec struct {
	Title string
	Body  string
}

// Template é um modelo pronto de contrato (conjunto ordenado de cláusulas).
type Template struct {
	Key         string
	Name        string
	Description string
	specs       []clauseSpec
}

// Build monta as cláusulas do modelo preenchendo os marcadores.
func (t Template) Build(data ClauseData) []Clause {
	out := make([]Clause, 0, len(t.specs))
	for _, s := range t.specs {
		out = append(out, Clause{
			Title: strings.ToUpper(strings.TrimSpace(s.Title)),
			Body:  renderClause(s.Body, data),
		})
	}
	return out
}

// ClauseTitles devolve apenas os títulos das cláusulas do modelo.
func (t Template) ClauseTitles() []string {
	titles := make([]string, 0, len(t.specs))
	for _, s := range t.specs {
		titles = append(titles, strings.ToUpper(strings.TrimSpace(s.Title)))
	}
	return titles
}

func renderClause(body string, data ClauseData) string {
	amountPhrase := "a ser combinado entre as partes"
	if data.Amount.Amount > 0 {
		amountPhrase = "de " + data.Amount.FormatBRL()
	}
	paymentTerms := data.PaymentTerms
	if strings.TrimSpace(paymentTerms) == "" {
		paymentTerms = "conforme acordado entre as partes"
	}
	replacer := strings.NewReplacer(
		"{object}", fallback(data.Object, "(objeto a ser definido)"),
		"{amount}", amountPhrase,
		"{payment_terms}", paymentTerms,
		"{term}", fallback(data.Term, "(prazo a ser definido)"),
		"{place}", fallback(data.Place, "(comarca)"),
		"{contractor}", fallback(data.ContractorName, "CONTRATANTE"),
		"{contracted}", fallback(data.ContractedName, "CONTRATADO"),
	)
	return replacer.Replace(body)
}

// Templates devolve os modelos de contrato disponíveis na suíte.
func Templates() []Template {
	return []Template{templateBalanced(), templateProviderIP(), templateClientIP()}
}

// TemplateByKey busca um modelo pela chave.
func TemplateByKey(key string) (Template, bool) {
	for _, t := range Templates() {
		if t.Key == key {
			return t, true
		}
	}
	return Template{}, false
}

// DefaultTemplateKey é o modelo usado quando nenhum é escolhido.
func DefaultTemplateKey() string { return "equilibrado" }

// --- Modelo 1: equilibrado (protege o prestador sem abusar do contratante) ---

func templateBalanced() Template {
	return Template{
		Key:         "equilibrado",
		Name:        "Prestação de serviço (equilibrado)",
		Description: "Modelo justo para a maioria dos casos; protege o prestador quanto a pagamento, escopo e rescisão.",
		specs: []clauseSpec{
			{"OBJETO", "O presente contrato tem por objeto a prestação, pelo CONTRATADO ao CONTRATANTE, dos seguintes serviços: {object}. Os serviços serão executados com autonomia técnica, sem vínculo empregatício entre as partes."},
			{"PREÇO E FORMA DE PAGAMENTO", "Pela prestação dos serviços, o CONTRATANTE pagará ao CONTRATADO o valor {amount}, {payment_terms}. O atraso no pagamento sujeitará o CONTRATANTE a multa de 2% (dois por cento) sobre o valor em atraso, acrescida de juros de 1% (um por cento) ao mês."},
			{"OBRIGAÇÕES DO CONTRATADO", "O CONTRATADO obriga-se a executar os serviços com zelo, técnica e dentro do prazo ajustado, mantendo o CONTRATANTE informado sobre o andamento e comunicando eventuais impedimentos."},
			{"OBRIGAÇÕES DO CONTRATANTE", "O CONTRATANTE obriga-se a fornecer, em tempo hábil, todas as informações, materiais e acessos necessários à execução dos serviços, bem como a efetuar o pagamento na forma e nos prazos acordados."},
			{"PRAZO E VIGÊNCIA", "O presente contrato vigorará pelo prazo de {term}, contado a partir da assinatura, podendo ser prorrogado mediante acordo escrito entre as partes."},
			{"ALTERAÇÃO DE ESCOPO", "Qualquer serviço não previsto no objeto deste contrato será considerado escopo adicional e somente será executado mediante acordo prévio quanto a prazo e valor, formalizado por escrito, ainda que por meio eletrônico."},
			{"RESCISÃO", "O contrato poderá ser rescindido por qualquer das partes mediante comunicação prévia de 15 (quinze) dias. Em caso de rescisão, o CONTRATANTE pagará ao CONTRATADO os valores correspondentes aos serviços já executados até a data da rescisão."},
			{"CONFIDENCIALIDADE", "As partes obrigam-se a manter sigilo sobre as informações confidenciais a que tiverem acesso em razão deste contrato, mesmo após o seu término."},
			{"FORO", "Fica eleito o foro da comarca de {place} para dirimir quaisquer dúvidas oriundas do presente contrato, com renúncia a qualquer outro, por mais privilegiado que seja."},
		},
	}
}

// --- Modelo 2: propriedade intelectual do prestador ---

func templateProviderIP() Template {
	return Template{
		Key:         "propriedade_prestador",
		Name:        "Propriedade intelectual do prestador",
		Description: "O prestador mantém a titularidade do projeto/obra; o cliente recebe apenas licença de uso do resultado entregue.",
		specs: []clauseSpec{
			{"OBJETO", "O presente contrato tem por objeto a prestação, pelo CONTRATADO ao CONTRATANTE, dos seguintes serviços: {object}."},
			{"PREÇO E FORMA DE PAGAMENTO", "Pela prestação dos serviços, o CONTRATANTE pagará ao CONTRATADO o valor {amount}, {payment_terms}. O atraso sujeitará o CONTRATANTE a multa de 2% e juros de 1% ao mês sobre o valor em atraso."},
			{"PROPRIEDADE INTELECTUAL", "Todos os direitos de propriedade intelectual, autoria e titularidade sobre os projetos, criações, códigos, artes, metodologias e materiais desenvolvidos permanecem integralmente com o CONTRATADO. O pagamento remunera a prestação do serviço e concede ao CONTRATANTE uma licença de uso não exclusiva e limitada à finalidade contratada, não implicando cessão ou transferência de titularidade."},
			{"ARQUIVOS-FONTE E REUTILIZAÇÃO", "A entrega de arquivos-fonte, códigos editáveis ou materiais de produção não está incluída, salvo acordo específico e remuneração à parte. O CONTRATADO poderá reutilizar técnicas, componentes e conhecimentos genéricos em outros trabalhos."},
			{"OBRIGAÇÕES DAS PARTES", "O CONTRATADO executará os serviços com zelo e no prazo ajustado; o CONTRATANTE fornecerá as informações necessárias e efetuará o pagamento na forma combinada."},
			{"PRAZO E VIGÊNCIA", "O presente contrato vigorará pelo prazo de {term}, contado a partir da assinatura, podendo ser prorrogado mediante acordo escrito."},
			{"RESCISÃO", "O contrato poderá ser rescindido mediante comunicação prévia de 15 (quinze) dias, sendo devidos ao CONTRATADO os valores dos serviços já executados. A licença de uso somente é concedida após a quitação integral."},
			{"FORO", "Fica eleito o foro da comarca de {place} para dirimir quaisquer dúvidas oriundas do presente contrato."},
		},
	}
}

// --- Modelo 3: cessão de direitos ao cliente (work for hire) ---

func templateClientIP() Template {
	return Template{
		Key:         "propriedade_cliente",
		Name:        "Cessão de direitos ao cliente (work for hire)",
		Description: "Após o pagamento integral, os direitos patrimoniais sobre o resultado são cedidos ao cliente.",
		specs: []clauseSpec{
			{"OBJETO", "O presente contrato tem por objeto a prestação, pelo CONTRATADO ao CONTRATANTE, dos seguintes serviços: {object}, incluindo a entrega do resultado final ao CONTRATANTE."},
			{"PREÇO E FORMA DE PAGAMENTO", "Pela prestação dos serviços e pela cessão de direitos, o CONTRATANTE pagará ao CONTRATADO o valor {amount}, {payment_terms}. O atraso sujeitará o CONTRATANTE a multa de 2% e juros de 1% ao mês sobre o valor em atraso."},
			{"CESSÃO DE DIREITOS PATRIMONIAIS", "Mediante o pagamento integral do valor ajustado, o CONTRATADO cede e transfere ao CONTRATANTE, em caráter definitivo e exclusivo, os direitos patrimoniais sobre o resultado dos serviços, para uso, reprodução, modificação e distribuição, na forma da Lei nº 9.610/98. Os direitos morais de autoria permanecem com o autor, nos termos da lei."},
			{"ENTREGA DE ARQUIVOS", "Após a quitação integral, o CONTRATADO entregará ao CONTRATANTE os arquivos-fonte e materiais necessários ao pleno aproveitamento do resultado contratado."},
			{"OBRIGAÇÕES DAS PARTES", "O CONTRATADO executará os serviços com zelo e no prazo ajustado; o CONTRATANTE fornecerá as informações necessárias e efetuará o pagamento na forma combinada."},
			{"PRAZO E VIGÊNCIA", "O presente contrato vigorará pelo prazo de {term}, contado a partir da assinatura, podendo ser prorrogado mediante acordo escrito."},
			{"RESCISÃO", "Em caso de rescisão antes da conclusão, serão devidos ao CONTRATADO os valores dos serviços já executados, e a cessão de direitos limitar-se-á às entregas efetivamente pagas."},
			{"FORO", "Fica eleito o foro da comarca de {place} para dirimir quaisquer dúvidas oriundas do presente contrato."},
		},
	}
}
