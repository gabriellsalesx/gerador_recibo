package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"emissor/internal/config"
	"emissor/internal/storage"

	"charm.land/bubbles/v2/key"
	"charm.land/huh/v2"
)

var errCancelled = errors.New("cancelado")

func canUseHuh(stdin io.Reader, stdout io.Writer) bool {
	in, ok := stdin.(*os.File)
	if !ok || !isTerminalFile(in) {
		return false
	}
	out, ok := stdout.(*os.File)
	return ok && isTerminalFile(out)
}

func isTerminalFile(file *os.File) bool {
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func runHuhFields(stdin io.Reader, stdout io.Writer, fields ...huh.Field) error {
	if len(fields) == 0 {
		return nil
	}
	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(key.WithKeys("ctrl+c", "esc"), key.WithHelp("esc", "cancelar"))
	err := huh.NewForm(
		huh.NewGroup(fields...),
	).
		WithKeyMap(km).
		WithInput(stdin).
		WithOutput(stdout).
		WithShowHelp(true).
		Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return errCancelled
	}
	return err
}

func runHuhSection(stdin io.Reader, stdout io.Writer, title, description string, fields ...huh.Field) error {
	printSectionTitle(stdout, title, description)
	return runHuhFields(stdin, stdout, fields...)
}

func printSectionTitle(stdout io.Writer, title, description string) {
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, title)
	if description != "" {
		fmt.Fprintln(stdout, description)
	}
	fmt.Fprintln(stdout)
}

func promptReceiptHuh(stdin io.Reader, stdout io.Writer, opts *options, cfg config.Config) (bool, error) {
	var fields []huh.Field

	if opts.create.PayerName == "" {
		fields = append(fields, huh.NewInput().
			Title("Pagador").
			Value(&opts.create.PayerName).
			Validate(required("pagador")))
	}
	if opts.create.Value == "" {
		fields = append(fields, huh.NewInput().
			Title("Valor recebido").
			Value(&opts.create.Value).
			Validate(required("valor")))
	}
	if opts.create.Description == "" {
		fields = append(fields, huh.NewInput().
			Title("Referente a").
			Value(&opts.create.Description).
			Validate(required("referente")))
	}
	if opts.create.PaymentMethod == "" {
		opts.create.PaymentMethod = defaultString(cfg.Defaults.PaymentMethod, "Pix")
		fields = append(fields, huh.NewSelect[string]().
			Title("Forma de pagamento").
			Options(huhOptions(paymentMethodChoices(), opts.create.PaymentMethod)...).
			Value(&opts.create.PaymentMethod))
	}
	if opts.create.PaidAt == "" {
		opts.create.PaidAt = "hoje"
		fields = append(fields, huh.NewInput().
			Title("Data do pagamento").
			Value(&opts.create.PaidAt).
			Validate(required("data do pagamento")))
	}

	confirmed := true
	if !opts.yes {
		fields = append(fields, huh.NewConfirm().
			Title("Gerar recibo em PDF?").
			Value(&confirmed))
	}

	if err := runHuhSection(stdin, stdout, "Novo recibo", "Preencha os dados do recebimento.", fields...); err != nil {
		if errors.Is(err, errCancelled) {
			return false, nil
		}
		return false, err
	}
	if opts.create.PaymentMethod == "Outro" {
		customPayment := ""
		if err := runHuhSection(stdin, stdout, "Forma de pagamento", "Informe a forma de pagamento personalizada.", huh.NewInput().
			Title("Informe a forma de pagamento").
			Value(&customPayment).
			Validate(required("forma de pagamento"))); err != nil {
			if errors.Is(err, errCancelled) {
				return false, nil
			}
			return false, err
		}
		opts.create.PaymentMethod = customPayment
	}
	return confirmed, nil
}

func menuHuh(stdin io.Reader, stdout io.Writer, title string, choices []selectChoice, def string) (string, error) {
	value := def
	if value == "" && len(choices) > 0 {
		value = choices[0].Value
	}
	err := runHuhFields(stdin, stdout,
		huh.NewSelect[string]().
			Title(title).
			Options(huhOptions(choices, value)...).
			Value(&value),
	)
	return value, err
}

func runConfigMenuHuh(stdin io.Reader, stdout io.Writer, cfg *config.Config, path string, exists bool) error {
	if !exists || strings.TrimSpace(cfg.Issuer.Name) == "" {
		fmt.Fprintln(stdout, "Vamos configurar seus dados iniciais.")
		if err := editAllConfigHuh(stdin, stdout, cfg); err != nil {
			if errors.Is(err, errCancelled) {
				fmt.Fprintln(stdout, "Configuração cancelada.")
				return errCancelled
			}
			return err
		}
		if err := config.Save(path, *cfg); err != nil {
			return err
		}
		fmt.Fprintln(stdout, "Configuração salva em:")
		fmt.Fprintln(stdout, path)
		return nil
	}

	for {
		action, err := configActionHuh(stdin, stdout)
		if err != nil {
			if errors.Is(err, errCancelled) {
				return nil
			}
			return err
		}

		var editErr error
		switch action {
		case "issuer":
			editErr = editIssuerHuh(stdin, stdout, cfg)
		case "pix":
			editErr = editPixHuh(stdin, stdout, cfg)
		case "output":
			editErr = editOutputDefaultsHuh(stdin, stdout, cfg)
		case "pdf":
			editErr = editPDFHuh(stdin, stdout, cfg)
		case "all":
			editErr = editAllConfigHuh(stdin, stdout, cfg)
		case "view":
			printConfigSummary(stdout, *cfg)
			continue
		case "back":
			return nil
		default:
			continue
		}

		if editErr != nil {
			if errors.Is(editErr, errCancelled) {
				if fresh, _, _, e := config.LoadOrDefault(path); e == nil {
					*cfg = fresh
				}
				fmt.Fprintln(stdout, "Edição cancelada. Nada foi alterado.")
				continue
			}
			return editErr
		}
		if err := saveConfigMenu(stdout, path, *cfg); err != nil {
			return err
		}
	}
}

func configActionHuh(stdin io.Reader, stdout io.Writer) (string, error) {
	action := "issuer"
	err := runHuhSection(stdin, stdout, "Configurações", "Escolha uma área para editar ou volte ao menu principal.",
		huh.NewSelect[string]().
			Title("O que você quer configurar?").
			Options(
				huh.NewOption("Editar dados do emitente", "issuer").Selected(true),
				huh.NewOption("Editar Pix", "pix"),
				huh.NewOption("Editar pasta, padrões e observações", "output"),
				huh.NewOption("Editar aparência do PDF", "pdf"),
				huh.NewOption("Editar tudo", "all"),
				huh.NewOption("Ver resumo atual", "view"),
				huh.NewOption("Voltar", "back"),
			).
			Value(&action),
	)
	return action, err
}

func editAllConfigHuh(stdin io.Reader, stdout io.Writer, cfg *config.Config) error {
	if err := editIssuerHuh(stdin, stdout, cfg); err != nil {
		return err
	}
	if err := editPixHuh(stdin, stdout, cfg); err != nil {
		return err
	}
	if err := editOutputDefaultsHuh(stdin, stdout, cfg); err != nil {
		return err
	}
	return editPDFHuh(stdin, stdout, cfg)
}

func editIssuerHuh(stdin io.Reader, stdout io.Writer, cfg *config.Config) error {
	cfg.Issuer.Type = defaultString(cfg.Issuer.Type, "person")
	err := runHuhSection(stdin, stdout, "Dados do emitente", "Edite as informações que aparecem no recibo.",
		huh.NewSelect[string]().
			Title("Você emite recibos como").
			Options(huhOptions(issuerTypeChoices(), cfg.Issuer.Type)...).
			Value(&cfg.Issuer.Type),
		huh.NewInput().
			Title("Nome do emitente").
			Value(&cfg.Issuer.Name).
			Validate(required("nome do emitente")),
		huh.NewInput().
			Title("Nome fantasia").
			Value(&cfg.Issuer.TradeName),
		huh.NewInput().
			Title("CPF/CNPJ do emitente").
			Value(&cfg.Issuer.Document),
		huh.NewInput().
			Title("Telefone").
			Value(&cfg.Issuer.Phone),
		huh.NewInput().
			Title("E-mail").
			Value(&cfg.Issuer.Email),
		huh.NewInput().
			Title("Endereço").
			Value(&cfg.Issuer.Address),
		huh.NewInput().
			Title("Cidade").
			Value(&cfg.Issuer.City),
		huh.NewInput().
			Title("UF").
			Value(&cfg.Issuer.State),
	)
	normalizeIssuerConfig(cfg)
	return err
}

func editPixHuh(stdin io.Reader, stdout io.Writer, cfg *config.Config) error {
	if cfg.Pix.ReceiverName == "" {
		cfg.Pix.ReceiverName = cfg.Issuer.Name
	}
	if cfg.Pix.ReceiverCity == "" {
		cfg.Pix.ReceiverCity = cfg.Issuer.City
	}
	cfg.Issuer.PixKeyType = defaultString(cfg.Issuer.PixKeyType, "email")
	return runHuhSection(stdin, stdout, "Configuração Pix", "Defina a chave Pix e como ela aparece no PDF.",
		huh.NewConfirm().
			Title("Ativar Pix nos recibos?").
			Value(&cfg.Pix.Enabled),
		huh.NewInput().
			Title("Chave Pix padrão").
			Value(&cfg.Issuer.PixKey),
		huh.NewSelect[string]().
			Title("Tipo da chave Pix").
			Options(huhOptions(pixKeyTypeChoices(), cfg.Issuer.PixKeyType)...).
			Value(&cfg.Issuer.PixKeyType),
		huh.NewInput().
			Title("Nome Pix exibido no QR Code").
			Value(&cfg.Pix.ReceiverName),
		huh.NewInput().
			Title("Cidade Pix exibida no QR Code").
			Value(&cfg.Pix.ReceiverCity),
		huh.NewInput().
			Title("Descricao padrão do Pix").
			Value(&cfg.Pix.DefaultDescription),
		huh.NewConfirm().
			Title("Gerar QR Code Pix automáticamente quando o pagamento for Pix?").
			Value(&cfg.Pix.GenerateQRCodeByDefault),
		huh.NewConfirm().
			Title("Mostrar chave Pix escrita no recibo?").
			Value(&cfg.Pix.ShowKeyText),
		huh.NewConfirm().
			Title("Mostrar Pix Copia e Cola no recibo?").
			Value(&cfg.Pix.ShowCopyPaste),
	)
}

func editOutputDefaultsHuh(stdin io.Reader, stdout io.Writer, cfg *config.Config) error {
	cfg.Defaults.PaymentMethod = defaultString(cfg.Defaults.PaymentMethod, "Pix")
	return runHuhSection(stdin, stdout, "Saida e padrões", "Edite pasta, pagamento padrão e observacoes.",
		huh.NewInput().
			Title("Pasta de saida").
			Value(&cfg.Output.Directory).
			Validate(required("pasta de saida")),
		huh.NewSelect[string]().
			Title("Forma de pagamento padrão").
			Options(huhOptions(paymentMethodChoices(), cfg.Defaults.PaymentMethod)...).
			Value(&cfg.Defaults.PaymentMethod),
		huh.NewInput().
			Title("Observação padrão").
			Value(&cfg.Defaults.Notes),
	)
}

func editPDFHuh(stdin io.Reader, stdout io.Writer, cfg *config.Config) error {
	cfg.PDF.Template = defaultString(cfg.PDF.Template, "professional")
	return runHuhSection(stdin, stdout, "Aparencia do PDF", "Ajuste logo e assinatura.",
		huh.NewConfirm().
			Title("Mostrar logo no recibo?").
			Value(&cfg.PDF.ShowLogo),
		huh.NewInput().
			Title("Caminho do logo").
			Value(&cfg.PDF.LogoPath),
		huh.NewConfirm().
			Title("Mostrar assinatura no recibo?").
			Value(&cfg.PDF.ShowSignature),
		huh.NewInput().
			Title("Caminho da assinatura").
			Value(&cfg.PDF.SignaturePath),
	)
}

func chooseDocItemHuh(stdin io.Reader, stdout io.Writer, items []storage.DocItem, docLabel string) (storage.DocItem, bool, error) {
	if len(items) == 0 {
		return storage.DocItem{}, false, nil
	}
	choices := make([]huh.Option[int], 0, len(items)+1)
	for i, item := range items {
		label := fmt.Sprintf("%s | %s | %s | %s", item.Number, item.CreatedAt.Format("02/01/2006"), item.Counterparty, item.Amount.FormatBRL())
		choices = append(choices, huh.NewOption(label, i))
	}
	cancelIndex := len(items)
	choices = append(choices, huh.NewOption("Voltar", cancelIndex))
	selected := 0
	if err := runHuhFields(stdin, stdout,
		huh.NewSelect[int]().
			Title("Escolha um "+docLabel+" para abrir").
			Options(choices...).
			Value(&selected),
	); err != nil {
		if errors.Is(err, errCancelled) {
			return storage.DocItem{}, false, nil
		}
		return storage.DocItem{}, false, err
	}
	if selected == cancelIndex {
		return storage.DocItem{}, false, nil
	}
	return items[selected], true, nil
}

func saveConfigMenu(stdout io.Writer, path string, cfg config.Config) error {
	if err := config.Save(path, cfg); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "Configuração atualizada.")
	return nil
}

func printConfigSummary(stdout io.Writer, cfg config.Config) {
	fmt.Fprintln(stdout, "Configuração atual:")
	fmt.Fprintln(stdout, "Emitente:", cfg.Issuer.Name)
	fmt.Fprintln(stdout, "Tipo:", issuerTypeLabel(cfg.Issuer.Type))
	fmt.Fprintln(stdout, "Documento:", maskIfEmpty(cfg.Issuer.Document))
	fmt.Fprintln(stdout, "Telefone:", maskIfEmpty(cfg.Issuer.Phone))
	fmt.Fprintln(stdout, "E-mail:", maskIfEmpty(cfg.Issuer.Email))
	fmt.Fprintln(stdout, "Cidade/UF:", strings.TrimSpace(cfg.Issuer.City+" "+cfg.Issuer.State))
	fmt.Fprintln(stdout, "Pix configurado:", yesNo(cfg.Issuer.PixKey != ""))
	fmt.Fprintln(stdout, "Pasta de saida:", cfg.Output.Directory)
}

func normalizeIssuerConfig(cfg *config.Config) {
	if cfg.Issuer.Type == "company" {
		cfg.Issuer.DocumentType = "cnpj"
	} else {
		cfg.Issuer.DocumentType = "cpf"
	}
	cfg.Issuer.State = strings.ToUpper(strings.TrimSpace(cfg.Issuer.State))
	if cfg.Defaults.City == "" {
		cfg.Defaults.City = cfg.Issuer.City
	}
	if cfg.Defaults.State == "" {
		cfg.Defaults.State = cfg.Issuer.State
	}
}

func huhOptions(choices []selectChoice, def string) []huh.Option[string] {
	options := make([]huh.Option[string], 0, len(choices))
	for _, choice := range choices {
		options = append(options, huh.NewOption(choice.Label, choice.Value).Selected(strings.EqualFold(choice.Value, def)))
	}
	return options
}

func required(label string) func(string) error {
	return func(value string) error {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("o campo %q é obrigatório", label)
		}
		return nil
	}
}
