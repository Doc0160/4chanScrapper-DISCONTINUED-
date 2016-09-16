package main
import (
	"regexp"
	"strings"
	"bytes"
)
var baseNameSeparators = regexp.MustCompile(`[./]`)
var illegalName = regexp.MustCompile(`[^[:alnum:]-.]`)
var transliterations = map[rune]string{
	'À': "A",
	'Á': "A",
	'Â': "A",
	'Ã': "A",
	'Ä': "A",
	'Å': "AA",
	'Æ': "AE",
	'Ç': "C",
	'È': "E",
	'É': "E",
	'Ê': "E",
	'Ë': "E",
	'Ì': "I",
	'Í': "I",
	'Î': "I",
	'Ï': "I",
	'Ð': "D",
	'L': "L",
	'Ñ': "N",
	'Ò': "O",
	'Ó': "O",
	'Ô': "O",
	'Õ': "O",
	'Ö': "O",
	'Ø': "OE",
	'Ù': "U",
	'Ú': "U",
	'Ü': "U",
	'Û': "U",
	'Ý': "Y",
	'Þ': "Th",
	'ß': "ss",
	'à': "a",
	'á': "a",
	'â': "a",
	'ã': "a",
	'ä': "a",
	'å': "aa",
	'æ': "ae",
	'ç': "c",
	'è': "e",
	'é': "e",
	'ê': "e",
	'ë': "e",
	'ì': "i",
	'í': "i",
	'î': "i",
	'ï': "i",
	'ð': "d",
	'l': "l",
	'ñ': "n",
	'n': "n",
	'ò': "o",
	'ó': "o",
	'ô': "o",
	'õ': "o",
	'o': "o",
	'ö': "o",
	'ø': "oe",
	's': "s",
	'ù': "u",
	'ú': "u",
	'û': "u",
	'u': "u",
	'ü': "u",
	'ý': "y",
	'þ': "th",
	'ÿ': "y",
	'z': "z",
	'Œ': "OE",
	'œ': "oe",
}
var separators = regexp.MustCompile(`[ &_=+:]`)
var dashes = regexp.MustCompile(`[\-]+`)
func CleanName(s string) string {
	if len(s) > 200 {
		s = s[:200]
	}
	baseName := baseNameSeparators.ReplaceAllString(s, "-")
	baseName = cleanString(baseName, illegalName)
	return baseName
}
func cleanString(s string, r *regexp.Regexp) string {
	s = strings.Trim(s, " ")
	s = Accents(s)
	s = separators.ReplaceAllString(s, "-")
	s = r.ReplaceAllString(s, "")
	s = dashes.ReplaceAllString(s, "-")
	return s
}
func Accents(s string) string {
	b := bytes.NewBufferString("")
	for _, c := range s {
		if val, ok := transliterations[c]; ok {
			b.WriteString(val)
		} else {
			b.WriteRune(c)
		}
	}
	return b.String()
}