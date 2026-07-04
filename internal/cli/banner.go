package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
)

const bannerTitle = "EMISSOR"

func printBanner(w io.Writer) {
	colored := os.Getenv("NO_COLOR") == ""
	if colored {
		fmt.Fprint(w, "\x1b[38;5;141m")
	}
	for _, line := range renderDotText(bannerTitle) {
		fmt.Fprintln(w, line)
	}
	if colored {
		fmt.Fprint(w, "\x1b[0m")
		fmt.Fprint(w, "\x1b[38;5;245m")
	}
	fmt.Fprint(w, "\nSuíte de documentos: contratos, orçamentos e recibos em PDF")
	if colored {
		fmt.Fprint(w, "\x1b[0m")
		fmt.Fprint(w, "\x1b[38;5;141m")
	}
	fmt.Fprintln(w, "  |  Criado por: Gabriell Sales")
	if colored {
		fmt.Fprint(w, "\x1b[0m")
	}
	fmt.Fprintln(w)
}

func renderDotText(text string) []string {
	const height = 7
	lines := make([]string, height)
	for _, char := range strings.ToUpper(text) {
		glyph, ok := dotFont[char]
		if !ok {
			glyph = dotFont[' ']
		}
		for i := 0; i < height; i++ {
			if lines[i] != "" {
				lines[i] += " "
			}
			lines[i] += strings.ReplaceAll(glyph[i], "#", "o")
		}
	}
	return lines
}

var dotFont = map[rune][]string{
	' ': {
		"   ",
		"   ",
		"   ",
		"   ",
		"   ",
		"   ",
		"   ",
	},
	'A': {
		" ### ",
		"#   #",
		"#   #",
		"#####",
		"#   #",
		"#   #",
		"#   #",
	},
	'B': {
		"#### ",
		"#   #",
		"#   #",
		"#### ",
		"#   #",
		"#   #",
		"#### ",
	},
	'C': {
		" ####",
		"#    ",
		"#    ",
		"#    ",
		"#    ",
		"#    ",
		" ####",
	},
	'D': {
		"#### ",
		"#   #",
		"#   #",
		"#   #",
		"#   #",
		"#   #",
		"#### ",
	},
	'E': {
		"#####",
		"#    ",
		"#    ",
		"#### ",
		"#    ",
		"#    ",
		"#####",
	},
	'G': {
		" ####",
		"#    ",
		"#    ",
		"#  ##",
		"#   #",
		"#   #",
		" ####",
	},
	'I': {
		"#####",
		"  #  ",
		"  #  ",
		"  #  ",
		"  #  ",
		"  #  ",
		"#####",
	},
	'M': {
		"#   #",
		"## ##",
		"# # #",
		"# # #",
		"#   #",
		"#   #",
		"#   #",
	},
	'S': {
		" ####",
		"#    ",
		"#    ",
		" ### ",
		"    #",
		"    #",
		"#### ",
	},
	'O': {
		" ### ",
		"#   #",
		"#   #",
		"#   #",
		"#   #",
		"#   #",
		" ### ",
	},
	'R': {
		"#### ",
		"#   #",
		"#   #",
		"#### ",
		"#  # ",
		"#   #",
		"#   #",
	},
}
