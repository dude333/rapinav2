package infra

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func removeAccent(str string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, str)
	return result
}

func normalizeString(s string) string {
	// Converte para minúsculas e remove espaços em branco
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "")

	// Remove caracteres especiais
	var normalized []rune
	for _, char := range s {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			normalized = append(normalized, char)
		}
	}

	return removeAccent(string(normalized))
}

func AreSimilar(s1, s2 string) bool {
	normalizedS1 := normalizeString(s1)
	normalizedS2 := normalizeString(s2)

	return normalizedS1 == normalizedS2
}
