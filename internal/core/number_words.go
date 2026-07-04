package core

import "strings"

// MoneyInWords devolve o valor por extenso em português (reais e centavos).
func MoneyInWords(m Money) string {
	amount := m.Amount
	if amount < 0 {
		amount = -amount
	}
	reais := amount / 100
	centavos := amount % 100

	realWord := "reais"
	if reais == 1 {
		realWord = "real"
	}
	reaisWords := integerInWords(reais)
	reaisPart := reaisWords + realConnector(reais, reaisWords) + realWord

	if centavos == 0 {
		return reaisPart
	}

	centWord := "centavos"
	if centavos == 1 {
		centWord = "centavo"
	}
	return reaisPart + " e " + integerInWords(centavos) + " " + centWord
}

func realConnector(reais int64, words string) string {
	if reais > 0 && endsWithScaleWord(words) {
		return " de "
	}
	return " "
}

func endsWithScaleWord(s string) bool {
	for _, suffix := range []string{"milhão", "milhões", "bilhão", "bilhões"} {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}

func integerInWords(n int64) string {
	if n == 0 {
		return "zero"
	}
	if n < 0 {
		n = -n
	}
	if n >= 1_000_000 {
		millions := n / 1_000_000
		rest := n % 1_000_000
		word := integerInWords(millions) + " milhões"
		if millions == 1 {
			word = "um milhão"
		}
		return appendRest(word, rest)
	}
	if n >= 1_000 {
		thousands := n / 1_000
		rest := n % 1_000
		word := "mil"
		if thousands > 1 {
			word = underThousand(int(thousands)) + " mil"
		}
		return appendRest(word, rest)
	}
	return underThousand(int(n))
}

func appendRest(prefix string, rest int64) string {
	if rest == 0 {
		return prefix
	}
	sep := " "
	if rest < 100 || rest%100 == 0 {
		sep = " e "
	}
	return prefix + sep + integerInWords(rest)
}

func underThousand(n int) string {
	if n < 20 {
		return unitWord(n)
	}
	if n < 100 {
		tens := (n / 10) * 10
		rest := n % 10
		if rest == 0 {
			return tensWord(tens)
		}
		return tensWord(tens) + " e " + unitWord(rest)
	}
	if n == 100 {
		return "cem"
	}
	hundreds := (n / 100) * 100
	rest := n % 100
	if rest == 0 {
		return hundredsWord(hundreds)
	}
	return hundredsWord(hundreds) + " e " + underThousand(rest)
}

func unitWord(n int) string {
	switch n {
	case 0:
		return "zero"
	case 1:
		return "um"
	case 2:
		return "dois"
	case 3:
		return "três"
	case 4:
		return "quatro"
	case 5:
		return "cinco"
	case 6:
		return "seis"
	case 7:
		return "sete"
	case 8:
		return "oito"
	case 9:
		return "nove"
	case 10:
		return "dez"
	case 11:
		return "onze"
	case 12:
		return "doze"
	case 13:
		return "treze"
	case 14:
		return "quatorze"
	case 15:
		return "quinze"
	case 16:
		return "dezesseis"
	case 17:
		return "dezessete"
	case 18:
		return "dezoito"
	case 19:
		return "dezenove"
	default:
		return ""
	}
}

func tensWord(n int) string {
	switch n {
	case 20:
		return "vinte"
	case 30:
		return "trinta"
	case 40:
		return "quarenta"
	case 50:
		return "cinquenta"
	case 60:
		return "sessenta"
	case 70:
		return "setenta"
	case 80:
		return "oitenta"
	case 90:
		return "noventa"
	default:
		return ""
	}
}

func hundredsWord(n int) string {
	switch n {
	case 100:
		return "cem"
	case 200:
		return "duzentos"
	case 300:
		return "trezentos"
	case 400:
		return "quatrocentos"
	case 500:
		return "quinhentos"
	case 600:
		return "seiscentos"
	case 700:
		return "setecentos"
	case 800:
		return "oitocentos"
	case 900:
		return "novecentos"
	default:
		return "cento"
	}
}
