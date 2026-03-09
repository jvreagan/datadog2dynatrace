package query

import (
	"fmt"
	"strings"
	"unicode"
)

// EvaluateFormula substitutes named query variables in a formula expression
// with their corresponding metric selectors or DQL expressions.
//
// Example: EvaluateFormula("a / b * 100", map[string]string{"a": "sel1", "b": "sel2"})
//
//	→ "(sel1) / (sel2) * 100"
func EvaluateFormula(expr string, vars map[string]string) (string, error) {
	if strings.TrimSpace(expr) == "" {
		return "", fmt.Errorf("empty formula expression")
	}

	var result strings.Builder
	i := 0
	for i < len(expr) {
		c := expr[i]

		// Check if this is a single-letter identifier (a-z) that is not part
		// of a longer identifier or number.
		if c >= 'a' && c <= 'z' {
			// Make sure it's not part of a multi-character identifier
			isStandalone := true
			if i > 0 && (unicode.IsLetter(rune(expr[i-1])) || unicode.IsDigit(rune(expr[i-1])) || expr[i-1] == '_') {
				isStandalone = false
			}
			if i+1 < len(expr) && (unicode.IsLetter(rune(expr[i+1])) || unicode.IsDigit(rune(expr[i+1])) || expr[i+1] == '_') {
				isStandalone = false
			}

			if isStandalone {
				name := string(c)
				val, ok := vars[name]
				if !ok {
					return "", fmt.Errorf("unknown variable %q in formula %q", name, expr)
				}
				result.WriteString("(" + val + ")")
				i++
				continue
			}
		}

		result.WriteByte(c)
		i++
	}

	return result.String(), nil
}
