package query

import (
	"fmt"
	"strings"
	"unicode"
)

// EvaluateFormula substitutes named query variables in a formula expression
// with their corresponding metric selectors or DQL expressions.
//
// Supports both single-letter variables (a, b, c) and multi-character
// identifiers (query0, query1, cpu_total).
//
// Example: EvaluateFormula("a / b * 100", map[string]string{"a": "sel1", "b": "sel2"})
//
//	→ "(sel1) / (sel2) * 100"
//
// Example: EvaluateFormula("query0 + query1", map[string]string{"query0": "sel0", "query1": "sel1"})
//
//	→ "(sel0) + (sel1)"
func EvaluateFormula(expr string, vars map[string]string) (string, error) {
	if strings.TrimSpace(expr) == "" {
		return "", fmt.Errorf("empty formula expression")
	}

	var result strings.Builder
	i := 0
	for i < len(expr) {
		c := rune(expr[i])

		// Check if this starts an identifier [a-zA-Z_]
		if unicode.IsLetter(c) || c == '_' {
			// Greedily consume the full identifier [a-zA-Z0-9_]*
			start := i
			i++
			for i < len(expr) {
				nc := rune(expr[i])
				if unicode.IsLetter(nc) || unicode.IsDigit(nc) || nc == '_' {
					i++
				} else {
					break
				}
			}
			name := expr[start:i]
			val, ok := vars[name]
			if !ok {
				return "", fmt.Errorf("unknown variable %q in formula %q", name, expr)
			}
			result.WriteString("(" + val + ")")
			continue
		}

		result.WriteByte(expr[i])
		i++
	}

	return result.String(), nil
}
