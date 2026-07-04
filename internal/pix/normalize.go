package pix

import (
	"strings"
	"unicode"
)

func NormalizeName(s string) string {
	return normalizeText(s, 25)
}

func NormalizeCity(s string) string {
	return normalizeText(s, 15)
}

func NormalizeTxID(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "***"
	}
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToUpper(r))
		}
	}
	out := b.String()
	if out == "" {
		return "***"
	}
	if len(out) > 25 {
		return out[:25]
	}
	return out
}

func normalizeText(s string, max int) string {
	s = strings.ToUpper(removeAccents(strings.TrimSpace(s)))
	var b strings.Builder
	lastSpace := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastSpace = false
			continue
		}
		if unicode.IsSpace(r) && !lastSpace {
			b.WriteByte(' ')
			lastSpace = true
		}
	}
	out := strings.TrimSpace(b.String())
	if len(out) > max {
		return out[:max]
	}
	return out
}

func removeAccents(s string) string {
	replacer := strings.NewReplacer(
		"Á", "A", "À", "A", "Â", "A", "Ã", "A", "Ä", "A",
		"É", "E", "È", "E", "Ê", "E", "Ë", "E",
		"Í", "I", "Ì", "I", "Î", "I", "Ï", "I",
		"Ó", "O", "Ò", "O", "Ô", "O", "Õ", "O", "Ö", "O",
		"Ú", "U", "Ù", "U", "Û", "U", "Ü", "U",
		"Ç", "C", "Ñ", "N",
		"á", "A", "à", "A", "â", "A", "ã", "A", "ä", "A",
		"é", "E", "è", "E", "ê", "E", "ë", "E",
		"í", "I", "ì", "I", "î", "I", "ï", "I",
		"ó", "O", "ò", "O", "ô", "O", "õ", "O", "ö", "O",
		"ú", "U", "ù", "U", "û", "U", "ü", "U",
		"ç", "C", "ñ", "N",
	)
	return replacer.Replace(s)
}
