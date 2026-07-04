package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"emissor/internal/app"
	"emissor/internal/config"
	"emissor/internal/contract"

	"charm.land/huh/v2"
)

// gatherQuoteInput coleta os dados do orçamento. No terminal usa formulários
// huh agrupados (poucos programas = sem travar input); em pipe/script usa o
// modo texto.
func gatherQuoteInput(stdin io.Reader, stdout io.Writer, reader *bufio.Reader, cfg config.Config) (app.CreateQuoteInput, bool, error) {
	if canUseHuh(stdin, stdout) {
		return quoteInputHuh(stdin, stdout, cfg)
	}
	return quoteInputFallback(reader, stdout, cfg)
}

func quoteInputHuh(stdin io.Reader, stdout io.Writer, cfg config.Config) (app.CreateQuoteInput, bool, error) {
	input := app.CreateQuoteInput{}

	// Grupo 1: cliente.
	if err := runHuhFields(stdin, stdout,
		huh.NewInput().Title("Cliente").Value(&input.ClientName).Validate(required("cliente")),
		huh.NewInput().Title("CPF/CNPJ do cliente (opcional)").Value(&input.ClientDocument),
	); err != nil {
		return input, false, quoteCancel(err)
	}

	// Itens: um formulário por item (descrição + quantidade + unidade + valor + "adicionar outro?").
	for {
		desc, qty, unit, price := "", "1", "un", ""
		addMore := false
		if err := runHuhFields(stdin, stdout,
			huh.NewInput().Title(fmt.Sprintf("Descrição do item %d", len(input.Items)+1)).Value(&desc).Validate(required("descrição")),
			huh.NewInput().Title("Quantidade").Value(&qty),
			huh.NewInput().Title("Unidade (un, h, m²...)").Value(&unit),
			huh.NewInput().Title("Valor unitário").Value(&price).Validate(required("valor unitário")),
			huh.NewConfirm().Title("Adicionar outro item?").Value(&addMore),
		); err != nil {
			return input, false, quoteCancel(err)
		}
		input.Items = append(input.Items, app.QuoteItemInput{
			Description: desc, Quantity: qty, Unit: unit, UnitPrice: price,
		})
		if !addMore {
			break
		}
	}

	// Grupo final: totais, validade, condições, observações, Pix e confirmação.
	input.Validity = fmt.Sprintf("%d dias", cfg.Documents.Quote.Defaults.ValidityDays)
	input.PaymentTerms = cfg.Documents.Quote.Defaults.PaymentTerms
	confirmed := true
	fields := []huh.Field{
		huh.NewInput().Title("Desconto (opcional)").Value(&input.Discount),
		huh.NewInput().Title("Validade da proposta").Value(&input.Validity).Validate(required("validade")),
		huh.NewInput().Title("Condições de pagamento").Value(&input.PaymentTerms),
		huh.NewInput().Title("Prazo de execução (opcional)").Value(&input.Deadline),
		huh.NewInput().Title("Observações (opcional)").Value(&input.Notes),
	}
	if pixAvailable(cfg) {
		fields = append(fields, huh.NewConfirm().Title("Gerar QR Code Pix do total?").Value(&input.EnablePix))
	}
	fields = append(fields, huh.NewConfirm().Title("Gerar orçamento em PDF?").Value(&confirmed))
	if err := runHuhFields(stdin, stdout, fields...); err != nil {
		return input, false, quoteCancel(err)
	}
	return input, confirmed, nil
}

func quoteInputFallback(reader *bufio.Reader, stdout io.Writer, cfg config.Config) (app.CreateQuoteInput, bool, error) {
	input := app.CreateQuoteInput{}
	var err error
	if input.ClientName, err = askRequired(reader, stdout, "Cliente"); err != nil {
		return input, false, err
	}
	if input.ClientDocument, err = ask(reader, stdout, "CPF/CNPJ do cliente (opcional)", ""); err != nil {
		return input, false, err
	}
	for {
		item := app.QuoteItemInput{}
		if item.Description, err = askRequired(reader, stdout, fmt.Sprintf("Descrição do item %d", len(input.Items)+1)); err != nil {
			return input, false, err
		}
		if item.Quantity, err = ask(reader, stdout, "Quantidade", "1"); err != nil {
			return input, false, err
		}
		if item.Unit, err = ask(reader, stdout, "Unidade (un, h, m²...)", "un"); err != nil {
			return input, false, err
		}
		if item.UnitPrice, err = askRequired(reader, stdout, "Valor unitário"); err != nil {
			return input, false, err
		}
		input.Items = append(input.Items, item)
		more, err := selectYesNo(reader, stdout, "Adicionar outro item?", false)
		if err != nil {
			return input, false, err
		}
		if !more {
			break
		}
	}
	if input.Discount, err = ask(reader, stdout, "Desconto (opcional)", ""); err != nil {
		return input, false, err
	}
	if input.Validity, err = ask(reader, stdout, "Validade da proposta", fmt.Sprintf("%d dias", cfg.Documents.Quote.Defaults.ValidityDays)); err != nil {
		return input, false, err
	}
	if input.PaymentTerms, err = ask(reader, stdout, "Condições de pagamento", cfg.Documents.Quote.Defaults.PaymentTerms); err != nil {
		return input, false, err
	}
	if input.Deadline, err = ask(reader, stdout, "Prazo de execução (opcional)", ""); err != nil {
		return input, false, err
	}
	if input.Notes, err = ask(reader, stdout, "Observações (opcional)", ""); err != nil {
		return input, false, err
	}
	if pixAvailable(cfg) {
		if input.EnablePix, err = selectYesNo(reader, stdout, "Gerar QR Code Pix do total?", false); err != nil {
			return input, false, err
		}
	}
	confirm, err := selectYesNo(reader, stdout, "Gerar orçamento em PDF?", true)
	if err != nil {
		return input, false, err
	}
	return input, confirm, nil
}

// gatherContractInput coleta os dados do contrato (huh agrupado ou modo texto).
func gatherContractInput(stdin io.Reader, stdout io.Writer, reader *bufio.Reader, cfg config.Config) (app.CreateContractInput, bool, error) {
	if canUseHuh(stdin, stdout) {
		return contractInputHuh(stdin, stdout, cfg)
	}
	return contractInputFallback(reader, stdout, cfg)
}

func contractInputHuh(stdin io.Reader, stdout io.Writer, cfg config.Config) (app.CreateContractInput, bool, error) {
	input := app.CreateContractInput{}
	input.Place = cfg.Documents.Contract.Defaults.Place
	input.SignedDate = "hoje"

	// Grupo 1: partes e dados principais (um único formulário).
	if err := runHuhFields(stdin, stdout,
		huh.NewInput().Title("Contratante (cliente)").Value(&input.ContractorName).Validate(required("contratante")),
		huh.NewInput().Title("CPF/CNPJ do contratante (opcional)").Value(&input.ContractorDocument),
		huh.NewInput().Title("Endereço do contratante (opcional)").Value(&input.ContractorAddress),
		huh.NewText().Title("Objeto do contrato (escopo do serviço)").Value(&input.Object).Validate(required("objeto")),
		huh.NewInput().Title("Valor do contrato (opcional se houver condições)").Value(&input.Value),
		huh.NewInput().Title("Condições de pagamento").Value(&input.PaymentTerms),
		huh.NewInput().Title("Prazo / vigência (ex.: 30 dias)").Value(&input.Term).Validate(required("prazo")),
		huh.NewInput().Title("Local de assinatura").Value(&input.Place).Validate(required("local")),
		huh.NewInput().Title("Data de assinatura").Value(&input.SignedDate),
	); err != nil {
		return input, false, quoteCancel(err)
	}

	// Grupo 2: modelo + confirmação.
	input.Template = contract.DefaultTemplateKey()
	confirmed := true
	if err := runHuhFields(stdin, stdout,
		huh.NewSelect[string]().Title("Modelo de contrato").
			Options(huhOptions(contractTemplateChoices(), input.Template)...).
			Value(&input.Template),
		huh.NewConfirm().Title("Gerar contrato em PDF?").Value(&confirmed),
	); err != nil {
		return input, false, quoteCancel(err)
	}
	return input, confirmed, nil
}

func contractInputFallback(reader *bufio.Reader, stdout io.Writer, cfg config.Config) (app.CreateContractInput, bool, error) {
	input := app.CreateContractInput{}
	var err error
	if input.ContractorName, err = askRequired(reader, stdout, "Contratante (cliente)"); err != nil {
		return input, false, err
	}
	if input.ContractorDocument, err = ask(reader, stdout, "CPF/CNPJ do contratante (opcional)", ""); err != nil {
		return input, false, err
	}
	if input.ContractorAddress, err = ask(reader, stdout, "Endereço do contratante (opcional)", ""); err != nil {
		return input, false, err
	}
	if input.Object, err = askRequired(reader, stdout, "Objeto do contrato (escopo do serviço)"); err != nil {
		return input, false, err
	}
	if input.Value, err = ask(reader, stdout, "Valor do contrato (opcional se houver condições)", ""); err != nil {
		return input, false, err
	}
	if input.PaymentTerms, err = ask(reader, stdout, "Condições de pagamento", ""); err != nil {
		return input, false, err
	}
	if input.Term, err = askRequired(reader, stdout, "Prazo / vigência (ex.: 30 dias)"); err != nil {
		return input, false, err
	}
	if input.Place, err = ask(reader, stdout, "Local de assinatura", cfg.Documents.Contract.Defaults.Place); err != nil {
		return input, false, err
	}
	if input.SignedDate, err = ask(reader, stdout, "Data de assinatura", "hoje"); err != nil {
		return input, false, err
	}
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Modelos de contrato disponíveis:")
	for _, t := range contract.Templates() {
		fmt.Fprintf(stdout, "  - %s: %s\n", t.Name, t.Description)
	}
	if input.Template, err = selectOption(reader, stdout, "Qual modelo de contrato usar?", contractTemplateChoices(), contract.DefaultTemplateKey()); err != nil {
		return input, false, err
	}
	confirm, err := selectYesNo(reader, stdout, "Gerar contrato em PDF?", true)
	if err != nil {
		return input, false, err
	}
	return input, confirm, nil
}

func pixAvailable(cfg config.Config) bool {
	return cfg.Pix.Enabled && cfg.Issuer.PixKey != ""
}

func quoteCancel(err error) error {
	if errors.Is(err, errCancelled) {
		return errCancelled
	}
	return err
}
