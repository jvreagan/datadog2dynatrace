package query

import (
	"fmt"
	"strings"
)

// ParsedQuery represents a parsed DataDog metric query.
type ParsedQuery struct {
	Aggregation string
	Metric      string
	Filters     map[string]string
	GroupBy     []string
	Function    string
	FuncArgs    []string // extra arguments to the wrapping function
	Rollup      *RollupDef
	AsModifier  string // "count", "rate"
	Fill        string // "zero", "null", "last", "linear", or a number
	RawQuery    string
}

// RollupDef represents a .rollup() modifier on a DD query.
type RollupDef struct {
	Method string // "avg", "sum", "min", "max", "count"
	Period int    // seconds
}

// Parse parses a DataDog metric query string.
//
// Supported formats:
//
//	avg:system.cpu.user{host:web01,env:prod} by {host}
//	top(avg:system.cpu.user{*} by {host}, 10, 'mean', 'desc')
//	sum:my.metric{*}.as_count()
//	avg:my.metric{*}.rollup(sum, 60)
//	avg:my.metric{*}.fill(zero)
//	per_second(sum:my.counter{*})
//	system.cpu.user{host:web01}   (no aggregation prefix)
func Parse(query string) (*ParsedQuery, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("empty query")
	}

	pq := &ParsedQuery{
		RawQuery: query,
		Filters:  make(map[string]string),
	}

	// Step 1: Strip outer wrapping function like top(...), per_second(...), abs(...)
	inner, funcName, funcArgs := unwrapFunction(query)
	if funcName != "" {
		pq.Function = funcName
		pq.FuncArgs = funcArgs
		query = inner
	}

	// Step 2: Extract aggregation prefix (e.g., "avg:", "sum:", "p99:")
	// Only valid if the prefix is a known aggregation keyword with no dots/braces.
	if idx := strings.Index(query, ":"); idx > 0 {
		candidate := query[:idx]
		if isAggregation(candidate) {
			pq.Aggregation = candidate
			query = query[idx+1:]
		}
	}

	// Step 3: Split off trailing modifiers (.rollup(), .as_count(), .as_rate(), .fill())
	query = parseTrailingModifiers(query, pq)

	// Step 4: Parse metric name, filters, group by
	if err := parseMetricBody(query, pq); err != nil {
		return nil, err
	}

	return pq, nil
}

// unwrapFunction detects and strips an outer function call.
// Returns the inner expression, function name, and any extra comma-separated args.
//
// Example: top(avg:metric{*} by {host}, 10, 'mean', 'desc')
//
//	→ inner="avg:metric{*} by {host}", name="top", args=["10","mean","desc"]
func unwrapFunction(query string) (inner, name string, args []string) {
	parenIdx := strings.Index(query, "(")
	if parenIdx <= 0 {
		return query, "", nil
	}

	candidate := query[:parenIdx]
	// A wrapping function is a simple identifier: no dots, no braces, no spaces
	if strings.ContainsAny(candidate, ".{} ") {
		return query, "", nil
	}

	// Find the matching closing paren
	depth := 0
	closeIdx := -1
	for i := parenIdx; i < len(query); i++ {
		switch query[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				closeIdx = i
				goto found
			}
		}
	}
found:
	if closeIdx < 0 {
		return query, "", nil
	}

	innerContent := query[parenIdx+1 : closeIdx]

	// Split the inner content on commas, but only at the top level (respect nested braces/parens)
	parts := splitTopLevel(innerContent, ',')

	// First part is always the metric expression; remaining parts are function args
	if len(parts) == 0 {
		return query, "", nil
	}

	var extraArgs []string
	for _, a := range parts[1:] {
		extraArgs = append(extraArgs, strings.Trim(strings.TrimSpace(a), "'\""))
	}

	return strings.TrimSpace(parts[0]), candidate, extraArgs
}

// splitTopLevel splits s on the given separator, but only at depth 0
// (not inside parentheses or braces).
func splitTopLevel(s string, sep byte) []string {
	var parts []string
	depth := 0
	start := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(', '{':
			depth++
		case ')', '}':
			depth--
		case sep:
			if depth == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, s[start:])
	return parts
}

// isAggregation returns true if the candidate looks like a DD aggregation keyword.
func isAggregation(s string) bool {
	if strings.ContainsAny(s, ".{} /+*-") {
		return false
	}
	known := map[string]bool{
		"avg": true, "sum": true, "min": true, "max": true,
		"count": true, "last": true,
		"p50": true, "p75": true, "p90": true, "p95": true, "p99": true,
	}
	return known[s]
}

// parseTrailingModifiers strips .rollup(), .as_count(), .as_rate(), .fill() from the query tail.
func parseTrailingModifiers(query string, pq *ParsedQuery) string {
	for {
		changed := false

		// .as_count()
		if idx := strings.Index(query, ".as_count()"); idx >= 0 {
			pq.AsModifier = "count"
			query = query[:idx] + query[idx+len(".as_count()"):]
			changed = true
		}

		// .as_rate()
		if idx := strings.Index(query, ".as_rate()"); idx >= 0 {
			pq.AsModifier = "rate"
			query = query[:idx] + query[idx+len(".as_rate()"):]
			changed = true
		}

		// .fill(...)
		if idx := strings.Index(query, ".fill("); idx >= 0 {
			end := strings.Index(query[idx:], ")")
			if end >= 0 {
				fillArg := query[idx+len(".fill(") : idx+end]
				pq.Fill = strings.TrimSpace(fillArg)
				query = query[:idx] + query[idx+end+1:]
				changed = true
			}
		}

		// .rollup(method) or .rollup(method, period)
		if idx := strings.Index(query, ".rollup("); idx >= 0 {
			end := strings.Index(query[idx:], ")")
			if end >= 0 {
				rollupContent := query[idx+len(".rollup(") : idx+end]
				pq.Rollup = parseRollupArgs(rollupContent)
				query = query[:idx] + query[idx+end+1:]
				changed = true
			}
		}

		if !changed {
			break
		}
	}
	return query
}

// parseRollupArgs parses "sum" or "sum, 60" into a RollupDef.
func parseRollupArgs(content string) *RollupDef {
	parts := strings.SplitN(content, ",", 2)
	rd := &RollupDef{
		Method: strings.TrimSpace(parts[0]),
	}
	if len(parts) == 2 {
		fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &rd.Period)
	}
	return rd
}

// parseMetricBody parses the metric name, {filters}, and by{groupby} from the query body.
func parseMetricBody(query string, pq *ParsedQuery) error {
	query = strings.TrimSpace(query)

	braceStart := strings.Index(query, "{")
	if braceStart < 0 {
		// No braces — just a bare metric name
		pq.Metric = strings.TrimSpace(query)
		return nil
	}

	pq.Metric = strings.TrimSpace(query[:braceStart])

	// Find matching close brace (handle nested braces just in case)
	braceEnd := findMatchingBrace(query, braceStart)
	if braceEnd < 0 {
		return fmt.Errorf("unclosed brace in query: %s", pq.RawQuery)
	}

	// Parse filters
	filterStr := query[braceStart+1 : braceEnd]
	if filterStr != "" && filterStr != "*" {
		parseFilters(filterStr, pq)
	}

	// Parse group by — look for " by {" or "by{" or " by{" after the filter brace
	remaining := query[braceEnd+1:]
	remaining = strings.TrimSpace(remaining)

	// Normalize: strip leading dot if present (some DD queries use ".by{")
	remaining = strings.TrimPrefix(remaining, ".")

	if strings.HasPrefix(remaining, "by{") || strings.HasPrefix(remaining, "by {") {
		byBraceStart := strings.Index(remaining, "{")
		byBraceEnd := findMatchingBrace(remaining, byBraceStart)
		if byBraceStart >= 0 && byBraceEnd > byBraceStart {
			groupByStr := remaining[byBraceStart+1 : byBraceEnd]
			for _, g := range strings.Split(groupByStr, ",") {
				g = strings.TrimSpace(g)
				if g != "" {
					pq.GroupBy = append(pq.GroupBy, g)
				}
			}
		}
	}

	return nil
}

// parseFilters parses "host:web01,env:prod" or "host:web01 AND env:prod" or negation "!env:staging".
func parseFilters(filterStr string, pq *ParsedQuery) {
	// DD supports comma-separated and AND/OR in filters
	// Normalize: replace " AND " with ","
	filterStr = strings.ReplaceAll(filterStr, " AND ", ",")
	filterStr = strings.ReplaceAll(filterStr, " and ", ",")

	for _, f := range strings.Split(filterStr, ",") {
		f = strings.TrimSpace(f)
		if f == "" || f == "*" {
			continue
		}

		// Handle negation (!key:value) — store with "!" prefix on key for downstream handling
		negated := false
		if strings.HasPrefix(f, "!") {
			negated = true
			f = f[1:]
		}

		if kv := strings.SplitN(f, ":", 2); len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			val := strings.TrimSpace(kv[1])
			if negated {
				key = "!" + key
			}
			pq.Filters[key] = val
		}
	}
}

// findMatchingBrace finds the index of the closing brace matching the opening brace at openIdx.
func findMatchingBrace(s string, openIdx int) int {
	depth := 0
	for i := openIdx; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// ParseLogQuery parses a DataDog log search query into a translated form.
func ParseLogQuery(query string) string {
	return query
}
