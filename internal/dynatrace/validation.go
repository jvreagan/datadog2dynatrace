package dynatrace

import (
	"net/url"
	"strings"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/logging"
)

// ValidationResult holds the results of validating metric selectors against
// the Dynatrace metrics API.
type ValidationResult struct {
	Selectors []SelectorValidation
	Summary   ValidationSummary
}

// SelectorValidation captures the validation outcome for a single metric selector.
type SelectorValidation struct {
	Selector string
	Sources  []string // e.g. "monitor: High CPU", "dashboard tile: CPU Usage"
	Valid    bool
	Error    string
	Skipped  bool // true for placeholder selectors
}

// ValidationSummary provides aggregate counts.
type ValidationSummary struct {
	Total   int
	Valid   int
	Invalid int
	Skipped int
}

// ValidateMetricSelector queries the Dynatrace metrics API to check whether
// the given selector is syntactically and semantically valid.
func (c *Client) ValidateMetricSelector(selector string) error {
	path := "/api/v2/metrics/query?metricSelector=" + url.QueryEscape(selector) + "&from=now-5m&to=now"
	_, err := c.get(path)
	return err
}

// ValidateAll validates all unique metric selectors found in the conversion
// result. Selectors that appear in log alert or composite monitor placeholders
// (identified by a "Migration Note" in the description) are skipped.
func (c *Client) ValidateAll(result *ConversionResult) *ValidationResult {
	// Collect unique selectors → list of source labels.
	selectorSources := map[string][]string{}

	// Track which selectors come from placeholder monitors.
	placeholderSelectors := map[string]bool{}

	for _, me := range result.MetricEvents {
		if me.MetricSelector == "" {
			continue
		}
		source := "monitor: " + me.Summary
		selectorSources[me.MetricSelector] = append(selectorSources[me.MetricSelector], source)
		if strings.Contains(me.Description, "Migration Note") {
			placeholderSelectors[me.MetricSelector] = true
		}
	}

	for _, d := range result.Dashboards {
		for _, tile := range d.Tiles {
			for _, q := range tile.Queries {
				if q.MetricSelector == "" {
					continue
				}
				source := "dashboard tile: " + tile.Name
				selectorSources[q.MetricSelector] = append(selectorSources[q.MetricSelector], source)
			}
		}
	}

	vr := &ValidationResult{}

	for selector, sources := range selectorSources {
		sv := SelectorValidation{
			Selector: selector,
			Sources:  sources,
		}

		if placeholderSelectors[selector] {
			sv.Skipped = true
			vr.Summary.Skipped++
			logging.Info("skipping placeholder selector: %s", selector)
		} else {
			err := c.ValidateMetricSelector(selector)
			if err != nil {
				sv.Error = err.Error()
				logging.Info("invalid selector: %s → %s", selector, err)
			} else {
				sv.Valid = true
			}
		}

		vr.Selectors = append(vr.Selectors, sv)
	}

	vr.Summary.Total = len(vr.Selectors)
	for _, sv := range vr.Selectors {
		if sv.Valid {
			vr.Summary.Valid++
		} else if !sv.Skipped {
			vr.Summary.Invalid++
		}
	}

	return vr
}
