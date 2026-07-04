package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"emissor/internal/app"
	"emissor/internal/config"
	"emissor/internal/contract"
	"emissor/internal/core"
	"emissor/internal/pix"
	"emissor/internal/platform"
	"emissor/internal/storage"

	"github.com/spf13/cobra"
)

type options struct {
	create     app.CreateReceiptInput
	configPath string
	yes        bool
	limit      int
	pixValue   string
	openList   bool
	links      bool
}

func Execute() {
	root := newRootCommand(os.Stdin, os.Stdout, os.Stderr)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCommand(stdin io.Reader, stdout, stderr io.Writer) *cobra.Command {
	opts := &options{}
	root := &cobra.Command{
		Use:           "emissor-cli",
		Short:         "Emissor — suíte de documentos (contrato, orçamento e recibo) em PDF",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if hasCreateInput(opts) {
				return runReceiptNew(stdin, stdout, opts)
			}
			return runHome(stdin, stdout, opts)
		},
	}

	root.AddCommand(newReceiptCommand(stdin, stdout, opts))
	root.AddCommand(newQuoteCommand(stdin, stdout, opts))
	root.AddCommand(newContractCommand(stdin, stdout, opts))
	root.AddCommand(newConfigCommand(stdin, stdout, opts))
	root.AddCommand(newPixCommand(stdin, stdout, opts))
	root.AddCommand(newFolderCommand(stdout, opts))

	root.SetErr(stderr)
	localizeHelp(root)
	return root
}

// ----- Comandos por documento -----

func newReceiptCommand(stdin io.Reader, stdout io.Writer, opts *options) *cobra.Command {
	reciboCmd := &cobra.Command{
		Use:          "recibo",
		Short:        "Recibos: novo, listar e abrir",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReceiptMenu(stdin, stdout, opts)
		},
	}
	novo := &cobra.Command{
		Use:          "novo",
		Short:        "Cria um novo recibo",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReceiptNew(stdin, stdout, opts)
		},
	}
	bindCreateFlags(novo, opts)
	reciboCmd.AddCommand(novo)
	reciboCmd.AddCommand(newListCommand(stdin, stdout, opts, core.DocReceipt))
	reciboCmd.AddCommand(newOpenCommand(stdin, stdout, opts, core.DocReceipt))
	return reciboCmd
}

func newQuoteCommand(stdin io.Reader, stdout io.Writer, opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "orcamento",
		Short:        "Orçamentos: novo, listar e abrir",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQuoteMenu(stdin, stdout, opts)
		},
	}
	novo := &cobra.Command{
		Use:          "novo",
		Short:        "Cria um novo orçamento",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			printBanner(stdout)
			return runQuoteNewFlow(stdin, stdout, opts)
		},
	}
	novo.Flags().StringVar(&opts.configPath, "config", "", "caminho alternativo do config.json")
	cmd.AddCommand(novo)
	cmd.AddCommand(newListCommand(stdin, stdout, opts, core.DocQuote))
	cmd.AddCommand(newOpenCommand(stdin, stdout, opts, core.DocQuote))
	return cmd
}

func newContractCommand(stdin io.Reader, stdout io.Writer, opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "contrato",
		Short:        "Contratos: novo, listar e abrir",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runContractMenu(stdin, stdout, opts)
		},
	}
	novo := &cobra.Command{
		Use:          "novo",
		Short:        "Cria um novo contrato",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			printBanner(stdout)
			return runContractNewFlow(stdin, stdout, opts)
		},
	}
	novo.Flags().StringVar(&opts.configPath, "config", "", "caminho alternativo do config.json")
	cmd.AddCommand(novo)
	cmd.AddCommand(newListCommand(stdin, stdout, opts, core.DocContract))
	cmd.AddCommand(newOpenCommand(stdin, stdout, opts, core.DocContract))
	return cmd
}

func newListCommand(stdin io.Reader, stdout io.Writer, opts *options, docType core.DocType) *cobra.Command {
	singular, _ := docLabels(docType)
	cmd := &cobra.Command{
		Use:          "listar",
		Short:        "Lista " + singular + "s recentes",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDocList(stdin, stdout, opts, docType)
		},
	}
	cmd.Flags().IntVar(&opts.limit, "limite", 10, "quantidade máxima")
	cmd.Flags().StringVar(&opts.configPath, "config", "", "caminho alternativo do config.json")
	cmd.Flags().BoolVar(&opts.openList, "abrir", false, "escolher um item para abrir o PDF")
	cmd.Flags().BoolVar(&opts.links, "links", false, "mostra links clicáveis quando o terminal suportar")
	return cmd
}

func newOpenCommand(stdin io.Reader, stdout io.Writer, opts *options, docType core.DocType) *cobra.Command {
	singular, _ := docLabels(docType)
	cmd := &cobra.Command{
		Use:          "abrir [número]",
		Aliases:      []string{"ver", "visualizar"},
		Short:        "Abre um " + singular + " pelo número ou por seleção",
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			target := ""
			if len(args) > 0 {
				target = args[0]
			}
			return runDocOpen(stdin, stdout, opts, docType, target)
		},
	}
	cmd.Flags().IntVar(&opts.limit, "limite", 10, "quantidade máxima na seleção")
	cmd.Flags().StringVar(&opts.configPath, "config", "", "caminho alternativo do config.json")
	return cmd
}

func newConfigCommand(stdin io.Reader, stdout io.Writer, opts *options) *cobra.Command {
	configCmd := &cobra.Command{
		Use:          "config",
		Short:        "Configura os dados do emitente",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfig(stdin, stdout, opts.configPath)
		},
	}
	configCmd.PersistentFlags().StringVar(&opts.configPath, "config", "", "caminho alternativo do config.json")
	configCmd.AddCommand(&cobra.Command{
		Use:          "ver",
		Short:        "Mostra a configuração atual",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigView(stdout, opts.configPath)
		},
	})
	return configCmd
}

func newPixCommand(stdin io.Reader, stdout io.Writer, opts *options) *cobra.Command {
	pixCmd := &cobra.Command{Use: "pix", Short: "Ferramentas Pix"}
	pixConfigCmd := &cobra.Command{
		Use:          "config",
		Short:        "Configura dados e exibição do Pix",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPixConfig(stdin, stdout, opts.configPath)
		},
	}
	pixConfigCmd.Flags().StringVar(&opts.configPath, "config", "", "caminho alternativo do config.json")
	testCmd := &cobra.Command{
		Use:          "testar",
		Short:        "Gera um Pix Copia e Cola de teste",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPixTest(stdout, opts)
		},
	}
	testCmd.Flags().StringVar(&opts.pixValue, "valor", "10,00", "valor para teste")
	testCmd.Flags().StringVar(&opts.configPath, "config", "", "caminho alternativo do config.json")
	pixCmd.AddCommand(pixConfigCmd)
	pixCmd.AddCommand(testCmd)
	return pixCmd
}

func newFolderCommand(stdout io.Writer, opts *options) *cobra.Command {
	pastaCmd := &cobra.Command{
		Use:          "pasta",
		Short:        "Mostra a pasta de documentos",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFolder(stdout, opts.configPath, false)
		},
	}
	pastaCmd.Flags().StringVar(&opts.configPath, "config", "", "caminho alternativo do config.json")
	return pastaCmd
}

func localizeHelp(root *cobra.Command) {
	root.CompletionOptions.DisableDefaultCmd = true
	root.DisableFlagsInUseLine = true
	root.PersistentFlags().BoolP("help", "h", false, "mostra a ajuda deste comando")

	root.SetUsageTemplate(usageTemplatePTBR)

	helpCmd := &cobra.Command{
		Use:   "ajuda [comando]",
		Short: "Mostra a ajuda de um comando",
		Run: func(c *cobra.Command, args []string) {
			target, _, err := root.Find(args)
			if target == nil || err != nil {
				_ = root.Help()
				return
			}
			_ = target.Help()
		},
	}
	root.SetHelpCommand(helpCmd)
}

const usageTemplatePTBR = `Uso:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [comando]{{end}}{{if gt (len .Aliases) 0}}

Apelidos:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Exemplos:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Comandos disponíveis:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "ajuda"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Opções:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Opções globais:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Tópicos de ajuda adicionais:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [comando] --help" para mais informações sobre um comando.{{end}}
`

func bindCreateFlags(cmd *cobra.Command, opts *options) {
	flags := cmd.Flags()
	flags.StringVar(&opts.configPath, "config", "", "caminho alternativo do config.json")
	flags.StringVar(&opts.create.IssuerName, "emitente", "", "nome do emitente")
	flags.StringVar(&opts.create.IssuerDocument, "emitente-doc", "", "CPF/CNPJ do emitente")
	flags.StringVar(&opts.create.IssuerCity, "emitente-cidade", "", "cidade do emitente")
	flags.StringVar(&opts.create.IssuerState, "emitente-uf", "", "UF do emitente")
	flags.StringVar(&opts.create.PayerName, "pagador", "", "nome do pagador")
	flags.StringVar(&opts.create.PayerDocument, "pagador-doc", "", "CPF/CNPJ do pagador")
	flags.StringVar(&opts.create.Value, "valor", "", "valor recebido")
	flags.StringVar(&opts.create.Description, "referente", "", "descrição do serviço/produto")
	flags.StringVar(&opts.create.PaymentMethod, "pagamento", "", "forma de pagamento")
	flags.StringVar(&opts.create.PaidAt, "data", "", "data do pagamento")
	flags.StringVar(&opts.create.Notes, "observacoes", "", "observações do recibo")
	flags.StringVar(&opts.create.OutputDir, "saida", "", "pasta de saída")
	flags.StringVar(&opts.create.PixKey, "pix-chave", "", "chave Pix")
	flags.StringVar(&opts.create.PixKeyType, "pix-tipo", "", "tipo da chave Pix")
	flags.StringVar(&opts.create.PixReceiverName, "pix-nome", "", "nome do recebedor Pix")
	flags.StringVar(&opts.create.PixReceiverCity, "pix-cidade", "", "cidade do recebedor Pix")
	flags.BoolVar(&opts.create.DisablePix, "sem-pix", false, "não gerar QR Code Pix")
	flags.BoolVar(&opts.yes, "sim", false, "confirma sem perguntar")
}

func hasCreateInput(opts *options) bool {
	c := opts.create
	return c.IssuerName != "" ||
		c.IssuerDocument != "" ||
		c.IssuerCity != "" ||
		c.IssuerState != "" ||
		c.PayerName != "" ||
		c.PayerDocument != "" ||
		c.Value != "" ||
		c.Description != "" ||
		c.PaymentMethod != "" ||
		c.PaidAt != "" ||
		c.Notes != "" ||
		c.OutputDir != "" ||
		c.PixKey != "" ||
		c.PixKeyType != "" ||
		c.PixReceiverName != "" ||
		c.PixReceiverCity != "" ||
		c.DisablePix ||
		opts.yes
}

// ----- Menu principal e submenus -----

func runHome(stdin io.Reader, stdout io.Writer, opts *options) error {
	reader := bufio.NewReader(stdin)
	for {
		clearScreen(stdin, stdout)
		printBanner(stdout)

		action, err := chooseMenu(stdin, stdout, reader, "O que você quer gerar?", mainMenuChoices(), "contract")
		if err != nil {
			if errors.Is(err, errCancelled) {
				fmt.Fprintln(stdout, "Até logo.")
				return nil
			}
			return err
		}

		switch action {
		case "contract":
			if err := runContractMenu(stdin, stdout, opts); err != nil {
				return err
			}
		case "quote":
			if err := runQuoteMenu(stdin, stdout, opts); err != nil {
				return err
			}
		case "receipt":
			if err := runReceiptMenu(stdin, stdout, opts); err != nil {
				return err
			}
		case "config":
			runGlobalAction(stdin, stdout, func() error { return runConfig(stdin, stdout, opts.configPath) })
		case "folder":
			runGlobalAction(stdin, stdout, func() error { return runFolder(stdout, opts.configPath, true) })
		case "exit":
			fmt.Fprintln(stdout, "Até logo.")
			return nil
		}
	}
}

func runGlobalAction(stdin io.Reader, stdout io.Writer, fn func() error) {
	clearScreen(stdin, stdout)
	printBanner(stdout)
	if err := fn(); err != nil && !errors.Is(err, errCancelled) {
		fmt.Fprintln(stdout, "Não consegui concluir essa ação.")
		fmt.Fprintln(stdout, "Detalhe técnico:", err)
	}
	waitEnter(stdin, stdout)
}

func runReceiptMenu(stdin io.Reader, stdout io.Writer, opts *options) error {
	return runDocMenu(stdin, stdout, opts, "Recibo", receiptMenuChoices(), func(action string) error {
		switch action {
		case "new":
			return runReceiptNewFlow(stdin, stdout, opts)
		case "list":
			return runDocList(stdin, stdout, opts, core.DocReceipt)
		case "open":
			return runDocOpen(stdin, stdout, opts, core.DocReceipt, "")
		case "pix":
			return runPixConfig(stdin, stdout, opts.configPath)
		case "pix_test":
			return runPixTest(stdout, opts)
		}
		return nil
	})
}

func runQuoteMenu(stdin io.Reader, stdout io.Writer, opts *options) error {
	return runDocMenu(stdin, stdout, opts, "Orçamento", quoteMenuChoices(), func(action string) error {
		switch action {
		case "new":
			return runQuoteNewFlow(stdin, stdout, opts)
		case "list":
			return runDocList(stdin, stdout, opts, core.DocQuote)
		case "open":
			return runDocOpen(stdin, stdout, opts, core.DocQuote, "")
		}
		return nil
	})
}

func runContractMenu(stdin io.Reader, stdout io.Writer, opts *options) error {
	return runDocMenu(stdin, stdout, opts, "Contrato", contractMenuChoices(), func(action string) error {
		switch action {
		case "new":
			return runContractNewFlow(stdin, stdout, opts)
		case "list":
			return runDocList(stdin, stdout, opts, core.DocContract)
		case "open":
			return runDocOpen(stdin, stdout, opts, core.DocContract, "")
		case "clauses":
			return runClauseModels(stdout, opts.configPath)
		}
		return nil
	})
}

// runDocMenu roda um submenu de documento em loop até "Voltar"/cancelar.
func runDocMenu(stdin io.Reader, stdout io.Writer, opts *options, title string, choices []selectChoice, run func(string) error) error {
	reader := bufio.NewReader(stdin)
	for {
		clearScreen(stdin, stdout)
		printBanner(stdout)
		action, err := chooseMenu(stdin, stdout, reader, title+" — o que você quer fazer?", choices, "new")
		if err != nil {
			if errors.Is(err, errCancelled) {
				return nil
			}
			return err
		}
		if action == "back" {
			return nil
		}
		clearScreen(stdin, stdout)
		printBanner(stdout)
		err = run(action)
		resetCreateInput(opts)
		if err != nil {
			if errors.Is(err, errCancelled) {
				continue
			}
			fmt.Fprintln(stdout, "Não consegui concluir essa ação.")
			fmt.Fprintln(stdout, "Detalhe técnico:", err)
		}
		waitEnter(stdin, stdout)
	}
}

func clearScreen(stdin io.Reader, stdout io.Writer) {
	if canUseHuh(stdin, stdout) {
		fmt.Fprint(stdout, "\033[H\033[2J")
	}
}

func waitEnter(stdin io.Reader, stdout io.Writer) {
	if !canUseHuh(stdin, stdout) {
		return
	}
	fmt.Fprint(stdout, "\nPressione Enter para voltar ao menu...")
	bufio.NewReader(stdin).ReadString('\n')
}

func resetCreateInput(opts *options) {
	opts.create = app.CreateReceiptInput{}
	opts.yes = false
}

func chooseMenu(stdin io.Reader, stdout io.Writer, reader *bufio.Reader, title string, choices []selectChoice, def string) (string, error) {
	if canUseHuh(stdin, stdout) {
		return menuHuh(stdin, stdout, title, choices, def)
	}
	return selectOption(reader, stdout, title, choices, def)
}

func mainMenuChoices() []selectChoice {
	return []selectChoice{
		{Label: "Gerar Contrato", Value: "contract"},
		{Label: "Gerar Orçamento", Value: "quote"},
		{Label: "Gerar Recibo", Value: "receipt"},
		{Label: "Configurações", Value: "config"},
		{Label: "Abrir pasta de documentos", Value: "folder"},
		{Label: "Sair", Value: "exit"},
	}
}

func receiptMenuChoices() []selectChoice {
	return []selectChoice{
		{Label: "Novo recibo", Value: "new"},
		{Label: "Listar recibos", Value: "list"},
		{Label: "Abrir / visualizar recibo", Value: "open"},
		{Label: "Configurar Pix", Value: "pix"},
		{Label: "Testar Pix", Value: "pix_test"},
		{Label: "Voltar", Value: "back"},
	}
}

func quoteMenuChoices() []selectChoice {
	return []selectChoice{
		{Label: "Novo orçamento", Value: "new"},
		{Label: "Listar orçamentos", Value: "list"},
		{Label: "Abrir / visualizar orçamento", Value: "open"},
		{Label: "Voltar", Value: "back"},
	}
}

func contractMenuChoices() []selectChoice {
	return []selectChoice{
		{Label: "Novo contrato", Value: "new"},
		{Label: "Listar contratos", Value: "list"},
		{Label: "Abrir / visualizar contrato", Value: "open"},
		{Label: "Modelos de cláusulas", Value: "clauses"},
		{Label: "Voltar", Value: "back"},
	}
}

// ----- Recibo -----

func runReceiptNew(stdin io.Reader, stdout io.Writer, opts *options) error {
	printBanner(stdout)
	return runReceiptNewFlow(stdin, stdout, opts)
}

func runReceiptNewFlow(stdin io.Reader, stdout io.Writer, opts *options) error {
	reader := bufio.NewReader(stdin)
	cfg, cfgPath, exists, err := config.LoadOrDefault(opts.configPath)
	if err != nil {
		return err
	}
	if !exists && opts.create.IssuerName == "" {
		fmt.Fprintln(stdout, "Não encontrei sua configuração inicial. Vamos criar agora.")
		if canUseHuh(stdin, stdout) {
			if err := editAllConfigHuh(stdin, stdout, &cfg); err != nil {
				if errors.Is(err, errCancelled) {
					fmt.Fprintln(stdout, "Configuração cancelada. Nenhum recibo foi gerado.")
					return nil
				}
				return err
			}
		} else {
			if err := promptConfig(reader, stdout, &cfg); err != nil {
				return err
			}
		}
		if err := config.Save(cfgPath, cfg); err != nil {
			return err
		}
	}

	if canUseHuh(stdin, stdout) {
		confirmed, err := promptReceiptHuh(stdin, stdout, opts, cfg)
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Fprintln(stdout, "Operação cancelada. Nenhum recibo foi gerado.")
			return nil
		}
	} else {
		confirmed, err := promptReceiptFallback(reader, stdout, opts, cfg)
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Fprintln(stdout, "Operação cancelada. Nenhum recibo foi gerado.")
			return nil
		}
	}

	result, err := app.Service{ConfigPath: opts.configPath}.CreateReceipt(opts.create)
	if err != nil {
		return err
	}
	printCreated(stdout, "Recibo", result.PDFPath, result.Warnings)
	if result.Receipt.Pix != nil {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "QR Code Pix gerado offline. A confirmação do pagamento deve ser feita no app do banco.")
	}
	return nil
}

// ----- Orçamento -----

func runQuoteNewFlow(stdin io.Reader, stdout io.Writer, opts *options) error {
	reader := bufio.NewReader(stdin)
	cfg, cfgPath, exists, err := config.LoadOrDefault(opts.configPath)
	if err != nil {
		return err
	}
	if !exists {
		fmt.Fprintln(stdout, "Não encontrei sua configuração inicial. Vamos criar agora.")
		if err := ensureConfig(stdin, stdout, reader, &cfg, cfgPath); err != nil {
			return err
		}
	}

	input, confirmed, err := gatherQuoteInput(stdin, stdout, reader, cfg)
	if err != nil {
		return cancelOr(err)
	}
	if !confirmed {
		fmt.Fprintln(stdout, "Operação cancelada. Nenhum orçamento foi gerado.")
		return nil
	}

	result, err := app.Service{ConfigPath: opts.configPath}.CreateQuote(input)
	if err != nil {
		return err
	}
	printCreated(stdout, "Orçamento", result.PDFPath, result.Warnings)
	fmt.Fprintln(stdout, "Total:", result.Quote.Total.FormatBRL())
	return nil
}

// ----- Contrato -----

func runContractNewFlow(stdin io.Reader, stdout io.Writer, opts *options) error {
	reader := bufio.NewReader(stdin)
	cfg, cfgPath, exists, err := config.LoadOrDefault(opts.configPath)
	if err != nil {
		return err
	}
	if !exists {
		fmt.Fprintln(stdout, "Não encontrei sua configuração inicial. Vamos criar agora.")
		if err := ensureConfig(stdin, stdout, reader, &cfg, cfgPath); err != nil {
			return err
		}
	}

	input, confirmed, err := gatherContractInput(stdin, stdout, reader, cfg)
	if err != nil {
		return cancelOr(err)
	}
	if !confirmed {
		fmt.Fprintln(stdout, "Operação cancelada. Nenhum contrato foi gerado.")
		return nil
	}

	result, err := app.Service{ConfigPath: opts.configPath}.CreateContract(input)
	if err != nil {
		return err
	}
	printCreated(stdout, "Contrato", result.PDFPath, result.Warnings)
	return nil
}

func contractTemplateChoices() []selectChoice {
	var choices []selectChoice
	for _, t := range contract.Templates() {
		choices = append(choices, selectChoice{Label: t.Name, Value: t.Key})
	}
	return choices
}

func runClauseModels(stdout io.Writer, configPath string) error {
	cfg, _, _, err := config.LoadOrDefault(configPath)
	if err != nil {
		return err
	}
	data := contract.ClauseData{
		Object: "(objeto do serviço)",
		Term:   "(prazo)",
		Place:  cfg.Documents.Contract.Defaults.Place,
	}
	for _, tmpl := range contract.Templates() {
		fmt.Fprintln(stdout, "=== "+tmpl.Name+" ===")
		fmt.Fprintln(stdout, tmpl.Description)
		fmt.Fprintln(stdout)
		for i, clause := range tmpl.Build(data) {
			fmt.Fprintf(stdout, "CLÁUSULA %d — %s\n", i+1, clause.Title)
			if clause.Body != "" {
				fmt.Fprintln(stdout, "   "+clause.Body)
			}
		}
		fmt.Fprintln(stdout)
	}
	return nil
}

// ----- Config compartilhada -----

func ensureConfig(stdin io.Reader, stdout io.Writer, reader *bufio.Reader, cfg *config.Config, cfgPath string) error {
	if canUseHuh(stdin, stdout) {
		if err := editAllConfigHuh(stdin, stdout, cfg); err != nil {
			return err
		}
	} else {
		if err := promptConfig(reader, stdout, cfg); err != nil {
			return err
		}
	}
	return config.Save(cfgPath, *cfg)
}

func cancelOr(err error) error {
	if errors.Is(err, errCancelled) {
		return errCancelled
	}
	return err
}

func printCreated(stdout io.Writer, label, pdfPath string, warnings []string) {
	fmt.Fprintln(stdout, label+" gerado com sucesso.")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Arquivo:")
	fmt.Fprintln(stdout, pdfPath)
	for _, warning := range warnings {
		fmt.Fprintln(stdout, "Aviso:", warning)
	}
}

func runConfig(stdin io.Reader, stdout io.Writer, configPath string) error {
	reader := bufio.NewReader(stdin)
	cfg, path, exists, err := config.LoadOrDefault(configPath)
	if err != nil {
		return err
	}
	if canUseHuh(stdin, stdout) {
		if err := runConfigMenuHuh(stdin, stdout, &cfg, path, exists); err != nil && !errors.Is(err, errCancelled) {
			return err
		}
		return nil
	}
	if err := promptConfig(reader, stdout, &cfg); err != nil {
		return err
	}
	if err := config.Save(path, cfg); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "Configuração salva em:")
	fmt.Fprintln(stdout, path)
	return nil
}

func promptReceiptFallback(reader *bufio.Reader, stdout io.Writer, opts *options, cfg config.Config) (bool, error) {
	var err error
	if opts.create.PayerName == "" {
		opts.create.PayerName, err = askRequired(reader, stdout, "Pagador")
		if err != nil {
			return false, err
		}
	}
	if opts.create.Value == "" {
		opts.create.Value, err = askRequired(reader, stdout, "Valor recebido")
		if err != nil {
			return false, err
		}
	}
	if opts.create.Description == "" {
		opts.create.Description, err = askRequired(reader, stdout, "Referente a")
		if err != nil {
			return false, err
		}
	}
	if opts.create.PaymentMethod == "" && !opts.yes {
		opts.create.PaymentMethod, err = selectOption(reader, stdout, "Forma de pagamento", paymentMethodChoices(), defaultString(cfg.Defaults.PaymentMethod, "Pix"))
		if err != nil {
			return false, err
		}
		if opts.create.PaymentMethod == "Outro" {
			opts.create.PaymentMethod, err = askRequired(reader, stdout, "Informe a forma de pagamento")
			if err != nil {
				return false, err
			}
		}
	}
	if opts.create.PaidAt == "" && !opts.yes {
		opts.create.PaidAt, err = ask(reader, stdout, "Data do pagamento", "hoje")
		if err != nil {
			return false, err
		}
	}
	if !opts.yes {
		confirm, err := selectYesNo(reader, stdout, "Gerar recibo em PDF?", true)
		if err != nil {
			return false, err
		}
		if !confirm {
			return false, nil
		}
	}
	return true, nil
}

func promptConfig(reader *bufio.Reader, stdout io.Writer, cfg *config.Config) error {
	var err error
	cfg.Issuer.Type, err = selectOption(reader, stdout, "Você emite documentos como", issuerTypeChoices(), defaultString(cfg.Issuer.Type, "person"))
	if err != nil {
		return err
	}
	cfg.Issuer.Name, err = askRequiredWithDefault(reader, stdout, "Nome do emitente", cfg.Issuer.Name)
	if err != nil {
		return err
	}
	cfg.Issuer.Document, err = ask(reader, stdout, "CPF/CNPJ do emitente", cfg.Issuer.Document)
	if err != nil {
		return err
	}
	cfg.Issuer.City, err = ask(reader, stdout, "Cidade do emitente", cfg.Issuer.City)
	if err != nil {
		return err
	}
	cfg.Issuer.State, err = ask(reader, stdout, "UF do emitente", cfg.Issuer.State)
	if err != nil {
		return err
	}
	cfg.Issuer.PixKey, err = ask(reader, stdout, "Chave Pix padrão", cfg.Issuer.PixKey)
	if err != nil {
		return err
	}
	if cfg.Issuer.PixKey != "" {
		cfg.Issuer.PixKeyType, err = selectOption(reader, stdout, "Tipo da chave Pix", pixKeyTypeChoices(), defaultString(cfg.Issuer.PixKeyType, "email"))
		if err != nil {
			return err
		}
		cfg.Pix.ReceiverName, err = ask(reader, stdout, "Nome Pix exibido", defaultString(cfg.Pix.ReceiverName, cfg.Issuer.Name))
		if err != nil {
			return err
		}
		cfg.Pix.ReceiverCity, err = ask(reader, stdout, "Cidade Pix exibida", defaultString(cfg.Pix.ReceiverCity, cfg.Issuer.City))
		if err != nil {
			return err
		}
	}
	cfg.Output.Directory, err = ask(reader, stdout, "Pasta de saída", cfg.Output.Directory)
	return err
}

func runConfigView(stdout io.Writer, configPath string) error {
	cfg, path, exists, err := config.LoadOrDefault(configPath)
	if err != nil {
		return err
	}
	if !exists {
		fmt.Fprintln(stdout, "Não encontrei configuração inicial.")
		fmt.Fprintln(stdout, "Caminho esperado:", path)
		return nil
	}
	printConfigSummary(stdout, cfg)
	return nil
}

func runPixConfig(stdin io.Reader, stdout io.Writer, configPath string) error {
	reader := bufio.NewReader(stdin)
	cfg, path, _, err := config.LoadOrDefault(configPath)
	if err != nil {
		return err
	}
	if canUseHuh(stdin, stdout) {
		if err := editPixHuh(stdin, stdout, &cfg); err != nil {
			if errors.Is(err, errCancelled) {
				fmt.Fprintln(stdout, "Edição cancelada. Nada foi alterado.")
				return nil
			}
			return err
		}
	} else {
		if err := promptPixConfig(reader, stdout, &cfg); err != nil {
			return err
		}
	}
	if err := config.Save(path, cfg); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "Configuração Pix salva em:")
	fmt.Fprintln(stdout, path)
	return nil
}

func runFolder(stdout io.Writer, configPath string, open bool) error {
	cfg, _, _, err := config.LoadOrDefault(configPath)
	if err != nil {
		return err
	}
	dir, err := storage.ExpandPath(cfg.Output.Directory)
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, "Pasta de documentos:")
	fmt.Fprintln(stdout, dir)
	if open {
		if err := platform.OpenFile(dir); err != nil {
			fmt.Fprintln(stdout, "Não consegui abrir automaticamente. Use o caminho acima.")
			return nil
		}
	}
	return nil
}

func promptPixConfig(reader *bufio.Reader, stdout io.Writer, cfg *config.Config) error {
	var err error
	cfg.Pix.Enabled, err = selectYesNo(reader, stdout, "Ativar Pix nos documentos?", cfg.Pix.Enabled)
	if err != nil {
		return err
	}
	cfg.Issuer.PixKey, err = ask(reader, stdout, "Chave Pix padrão", cfg.Issuer.PixKey)
	if err != nil {
		return err
	}
	cfg.Issuer.PixKeyType, err = selectOption(reader, stdout, "Tipo da chave Pix", pixKeyTypeChoices(), defaultString(cfg.Issuer.PixKeyType, "email"))
	if err != nil {
		return err
	}
	cfg.Pix.ReceiverName, err = ask(reader, stdout, "Nome Pix exibido", defaultString(cfg.Pix.ReceiverName, cfg.Issuer.Name))
	if err != nil {
		return err
	}
	cfg.Pix.ReceiverCity, err = ask(reader, stdout, "Cidade Pix exibida", defaultString(cfg.Pix.ReceiverCity, cfg.Issuer.City))
	if err != nil {
		return err
	}
	cfg.Pix.GenerateQRCodeByDefault, err = selectYesNo(reader, stdout, "Gerar QR Code Pix automaticamente quando o pagamento for Pix?", cfg.Pix.GenerateQRCodeByDefault)
	if err != nil {
		return err
	}
	cfg.Pix.ShowKeyText, err = selectYesNo(reader, stdout, "Mostrar chave Pix escrita no documento?", cfg.Pix.ShowKeyText)
	if err != nil {
		return err
	}
	cfg.Pix.ShowCopyPaste, err = selectYesNo(reader, stdout, "Mostrar Pix Copia e Cola no documento?", cfg.Pix.ShowCopyPaste)
	return err
}

// ----- Listagem e abertura genéricas -----

func docLabels(docType core.DocType) (singular, header string) {
	switch docType {
	case core.DocQuote:
		return "orçamento", "Orçamentos recentes:"
	case core.DocContract:
		return "contrato", "Contratos recentes:"
	default:
		return "recibo", "Recibos recentes:"
	}
}

func listDocItems(docType core.DocType, outputDir, configPath string, limit int) ([]storage.DocItem, error) {
	return storage.ListDocItems(docType, []string{
		config.MetadataDirFor(configPath, docType),
		outputDir,
	}, limit)
}

func runDocList(stdin io.Reader, stdout io.Writer, opts *options, docType core.DocType) error {
	cfg, _, _, err := config.LoadOrDefault(opts.configPath)
	if err != nil {
		return err
	}
	limit := opts.limit
	if limit == 0 {
		limit = 10
	}
	items, err := listDocItems(docType, cfg.Output.Directory, opts.configPath, limit)
	if err != nil {
		return err
	}
	singular, _ := docLabels(docType)
	if len(items) == 0 {
		fmt.Fprintf(stdout, "Nenhum %s encontrado.\n", singular)
		return nil
	}
	printDocItems(stdout, docType, items, opts.links)
	if opts.openList {
		item, ok, err := chooseDocItem(stdin, stdout, items, singular)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintf(stdout, "Nenhum %s foi aberto.\n", singular)
			return nil
		}
		return openDocPDF(stdout, item)
	}
	return nil
}

func runDocOpen(stdin io.Reader, stdout io.Writer, opts *options, docType core.DocType, target string) error {
	cfg, _, _, err := config.LoadOrDefault(opts.configPath)
	if err != nil {
		return err
	}
	limit := opts.limit
	if limit == 0 {
		limit = 10
	}
	if target != "" {
		limit = 0
	}
	items, err := listDocItems(docType, cfg.Output.Directory, opts.configPath, limit)
	if err != nil {
		return err
	}
	singular, _ := docLabels(docType)
	if len(items) == 0 {
		fmt.Fprintf(stdout, "Nenhum %s encontrado.\n", singular)
		return nil
	}
	if target != "" {
		item, ok := findDocItem(items, target)
		if !ok {
			return fmt.Errorf("não encontrei %s com número %q", singular, target)
		}
		return openDocPDF(stdout, item)
	}
	printDocItems(stdout, docType, items, false)
	item, ok, err := chooseDocItem(stdin, stdout, items, singular)
	if err != nil {
		return err
	}
	if !ok {
		fmt.Fprintf(stdout, "Nenhum %s foi aberto.\n", singular)
		return nil
	}
	return openDocPDF(stdout, item)
}

func printDocItems(stdout io.Writer, docType core.DocType, items []storage.DocItem, links bool) {
	_, header := docLabels(docType)
	fmt.Fprintln(stdout, header)
	fmt.Fprintln(stdout)
	for i, item := range items {
		label := fmt.Sprintf("%02d. %s | %s | %-24s | %s",
			i+1, item.Number, item.CreatedAt.Format("02/01/2006"), item.Counterparty, item.Amount.FormatBRL())
		if links {
			label = terminalFileLink(label, item.PDFPath)
		}
		fmt.Fprintln(stdout, label)
		fmt.Fprintln(stdout, "    PDF:", item.PDFPath)
	}
}

func chooseDocItem(stdin io.Reader, stdout io.Writer, items []storage.DocItem, docLabel string) (storage.DocItem, bool, error) {
	if canUseHuh(stdin, stdout) {
		return chooseDocItemHuh(stdin, stdout, items, docLabel)
	}
	reader := bufio.NewReader(stdin)
	for {
		answer, err := ask(reader, stdout, "Digite o número da lista para abrir o PDF ou Enter para cancelar", "")
		if err != nil {
			return storage.DocItem{}, false, err
		}
		if strings.TrimSpace(answer) == "" {
			return storage.DocItem{}, false, nil
		}
		index, err := strconv.Atoi(answer)
		if err != nil || index < 1 || index > len(items) {
			fmt.Fprintln(stdout, "Escolha um número válido da lista.")
			continue
		}
		return items[index-1], true, nil
	}
}

func findDocItem(items []storage.DocItem, target string) (storage.DocItem, bool) {
	target = strings.TrimSpace(target)
	for _, item := range items {
		if item.Number == target || item.ID == target {
			return item, true
		}
	}
	return storage.DocItem{}, false
}

func openDocPDF(stdout io.Writer, item storage.DocItem) error {
	if err := platform.OpenFile(item.PDFPath); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "Abrindo PDF:")
	fmt.Fprintln(stdout, item.PDFPath)
	return nil
}

func terminalFileLink(label, path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	u := url.URL{Scheme: "file", Path: abs}
	return "\x1b]8;;" + u.String() + "\x1b\\" + label + "\x1b]8;;\x1b\\"
}

func runPixTest(stdout io.Writer, opts *options) error {
	cfg, _, exists, err := config.LoadOrDefault(opts.configPath)
	if err != nil {
		return err
	}
	if !exists || cfg.Issuer.PixKey == "" {
		return fmt.Errorf("chave Pix não configurada. Execute: emissor-cli pix config")
	}
	amount, err := core.ParseMoney(opts.pixValue)
	if err != nil {
		return err
	}
	payload, err := pix.BuildPayload(pix.Payment{
		Key:          cfg.Issuer.PixKey,
		KeyType:      cfg.Issuer.PixKeyType,
		ReceiverName: defaultString(cfg.Pix.ReceiverName, cfg.Issuer.Name),
		ReceiverCity: defaultString(cfg.Pix.ReceiverCity, cfg.Issuer.City),
		AmountCents:  amount.Amount,
		Description:  cfg.Pix.DefaultDescription,
		TxID:         cfg.Pix.TxIDPrefix + "TESTE",
	})
	if err != nil {
		return err
	}
	qr, err := pix.QRCodePNG(payload, 256)
	if err != nil {
		return err
	}
	terminalQR, err := pix.QRCodeTerminal(payload)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp("", "emissor-pix-*.png")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(qr); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "Chave Pix:", cfg.Issuer.PixKey)
	fmt.Fprintln(stdout, "Valor:", amount.FormatBRL())
	fmt.Fprintln(stdout, "Pix Copia e Cola:")
	fmt.Fprintln(stdout, payload)
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "QR Code Pix para teste:")
	fmt.Fprintln(stdout, terminalQR)
	fmt.Fprintln(stdout, "QR Code PNG:", tmp.Name())
	fmt.Fprintln(stdout, "A confirmação do pagamento depende do app bancário.")
	return nil
}

// ----- Prompts de baixo nível (modo texto) -----

func askRequired(reader *bufio.Reader, stdout io.Writer, label string) (string, error) {
	return askRequiredWithDefault(reader, stdout, label, "")
}

type selectChoice struct {
	Label string
	Value string
}

func issuerTypeChoices() []selectChoice {
	return []selectChoice{
		{Label: "Pessoa física", Value: "person"},
		{Label: "Empresa / MEI", Value: "company"},
	}
}

func paymentMethodChoices() []selectChoice {
	return []selectChoice{
		{Label: "Pix", Value: "Pix"},
		{Label: "Dinheiro", Value: "Dinheiro"},
		{Label: "Cartão", Value: "Cartão"},
		{Label: "Transferência", Value: "Transferência"},
		{Label: "Boleto", Value: "Boleto"},
		{Label: "Outro", Value: "Outro"},
	}
}

func pixKeyTypeChoices() []selectChoice {
	return []selectChoice{
		{Label: "E-mail", Value: "email"},
		{Label: "CPF", Value: "cpf"},
		{Label: "CNPJ", Value: "cnpj"},
		{Label: "Telefone", Value: "phone"},
		{Label: "Chave aleatória", Value: "random"},
	}
}

func issuerTypeLabel(value string) string {
	switch value {
	case "company":
		return "Empresa / MEI"
	case "person":
		return "Pessoa física"
	default:
		return maskIfEmpty(value)
	}
}

func selectYesNo(reader *bufio.Reader, stdout io.Writer, label string, def bool) (bool, error) {
	defValue := "não"
	if def {
		defValue = "sim"
	}
	value, err := selectOption(reader, stdout, label, []selectChoice{
		{Label: "Sim", Value: "sim"},
		{Label: "Não", Value: "não"},
	}, defValue)
	if err != nil {
		return false, err
	}
	return value == "sim", nil
}

func selectOption(reader *bufio.Reader, stdout io.Writer, label string, choices []selectChoice, def string) (string, error) {
	if len(choices) == 0 {
		return "", fmt.Errorf("nenhuma opção disponível para %s", label)
	}
	defaultIndex := 0
	for i, choice := range choices {
		if strings.EqualFold(choice.Value, def) {
			defaultIndex = i
			break
		}
	}

	for {
		fmt.Fprintf(stdout, "? %s:\n", label)
		for i, choice := range choices {
			fmt.Fprintf(stdout, "  %d) %s%s\n", i+1, choice.Label, defaultSuffix(i == defaultIndex))
		}
		fmt.Fprintf(stdout, "Escolha [%d]: ", defaultIndex+1)

		answer, err := reader.ReadString('\n')
		if err != nil && len(answer) == 0 {
			return "", err
		}
		answer = strings.TrimSpace(answer)
		if answer == "" {
			return choices[defaultIndex].Value, nil
		}
		index, err := strconv.Atoi(answer)
		if err == nil && index >= 1 && index <= len(choices) {
			return choices[index-1].Value, nil
		}
		for _, choice := range choices {
			if strings.EqualFold(answer, choice.Value) || strings.EqualFold(answer, choice.Label) {
				return choice.Value, nil
			}
		}
		fmt.Fprintln(stdout, "Escolha uma opção válida.")
	}
}

func defaultSuffix(ok bool) string {
	if ok {
		return " (padrão)"
	}
	return ""
}

func askRequiredWithDefault(reader *bufio.Reader, stdout io.Writer, label, def string) (string, error) {
	for {
		value, err := ask(reader, stdout, label, def)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(value) != "" {
			return value, nil
		}
		fmt.Fprintln(stdout, "Este campo é obrigatório.")
	}
}

func ask(reader *bufio.Reader, stdout io.Writer, label, def string) (string, error) {
	if def != "" {
		fmt.Fprintf(stdout, "? %s [%s]: ", label, def)
	} else {
		fmt.Fprintf(stdout, "? %s: ", label)
	}
	value, err := reader.ReadString('\n')
	if err != nil && len(value) == 0 {
		return "", err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return def, nil
	}
	return value, nil
}

func defaultString(value, def string) string {
	if strings.TrimSpace(value) == "" {
		return def
	}
	return value
}

func maskIfEmpty(value string) string {
	if value == "" {
		return "(não informado)"
	}
	return value
}

func yesNo(ok bool) string {
	if ok {
		return "sim"
	}
	return "não"
}
