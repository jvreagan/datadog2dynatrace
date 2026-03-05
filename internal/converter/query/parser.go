package query

import (
	"fmt"
	"strings"
)

// ParsedQuery represents a parsed DataDog metric query.
type ParsedQuery struct {
	Function    string
	Metric      string
	Filters     map[string]string
	GroupBy     []string
	Aggregation string
	RawQuery    string
}

// Parse parses a DataDog metric query string.
// DD query format: function(metric{filters}by{groupby})
// Example: avg:system.cpu.user{host:web01,env:prod}by{host}
func Parse(query string) (*ParsedQuery, error) {
	if query == "" {
		return nil, fmt.Errorf("empty query")
	}

	pq := &ParsedQuery{
		RawQuery: query,
		Filters:  make(map[string]string),
	}

	// Split on first colon to get aggregation prefix (e.g., "avg:", "sum:")
	// Only treat it as an aggregation if the part before the colon looks like
	// a simple keyword (no dots, braces, or spaces — which would indicate it's
	// part of a metric name or filter, e.g., "system.cpu.user{host:web01}")
	if idx := strings.Index(query, ":"); idx > 0 {
		candidate := query[:idx]
		if !strings.ContainsAny(candidate, ".{} ") {
			pq.Aggregation = candidate
			query = query[idx+1:]
		}
	}

	// Check for wrapping function like top(), rollup(), etc.
	if idx := strings.Index(query, "("); idx > 0 {
		possibleFunc := query[:idx]
		// If this looks like a function (not a metric name with dots)
		if !strings.Contains(possibleFunc, ".") && !strings.Contains(possibleFunc, "{") {
			pq.Function = possibleFunc
			// Find the matching closing paren
			depth := 0
			end := -1
			for i := idx; i < len(query); i++ {
				if query[i] == '(' {
					depth++
				} else if query[i] == ')' {
					depth--
					if depth == 0 {
						end = i
						break
					}
				}
			}
			if end > 0 {
				query = query[idx+1 : end]
			}
		}
	}

	// Parse metric name and filters
	if braceStart := strings.Index(query, "{"); braceStart >= 0 {
		pq.Metric = query[:braceStart]

		braceEnd := strings.Index(query, "}")
		if braceEnd < 0 {
			return nil, fmt.Errorf("unclosed brace in query: %s", pq.RawQuery)
		}

		filterStr := query[braceStart+1 : braceEnd]
		if filterStr != "" && filterStr != "*" {
			for _, f := range strings.Split(filterStr, ",") {
				f = strings.TrimSpace(f)
				if kv := strings.SplitN(f, ":", 2); len(kv) == 2 {
					pq.Filters[kv[0]] = kv[1]
				}
			}
		}

		// Check for "by{...}" group by clause
		remaining := query[braceEnd+1:]
		if strings.HasPrefix(remaining, "by{") || strings.HasPrefix(remaining, " by {") || strings.HasPrefix(remaining, ".by{") {
			byStart := strings.Index(remaining, "{")
			byEnd := strings.Index(remaining, "}")
			if byStart >= 0 && byEnd > byStart {
				groupByStr := remaining[byStart+1 : byEnd]
				for _, g := range strings.Split(groupByStr, ",") {
					g = strings.TrimSpace(g)
					if g != "" {
						pq.GroupBy = append(pq.GroupBy, g)
					}
				}
			}
		}
	} else {
		// No filters, just metric name (possibly with .as(), .rollup(), etc.)
		if dotIdx := strings.Index(query, ".as("); dotIdx > 0 {
			pq.Metric = query[:dotIdx]
		} else if dotIdx := strings.Index(query, ".rollup("); dotIdx > 0 {
			pq.Metric = query[:dotIdx]
		} else {
			pq.Metric = strings.TrimSpace(query)
		}
	}

	// Clean up metric name
	pq.Metric = strings.TrimSpace(pq.Metric)

	return pq, nil
}

// ParseLogQuery parses a DataDog log search query.
func ParseLogQuery(query string) string {
	// DD log queries use Lucene-like syntax
	// DT uses DQL - we return a best-effort translation
	return query
}
