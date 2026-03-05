package converter

import (
	"fmt"
	"strings"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

// ConvertLogPipeline converts a DataDog log pipeline to a Dynatrace log processing rule.
func ConvertLogPipeline(dd *datadog.LogPipeline) (*dynatrace.LogProcessingRule, error) {
	rule := &dynatrace.LogProcessingRule{
		Name:    dd.Name,
		Enabled: dd.IsEnabled,
	}

	// Convert the pipeline filter to a DT query
	if dd.Filter != nil && dd.Filter.Query != "" {
		rule.Query = translateLogFilter(dd.Filter.Query)
	} else {
		rule.Query = "*"
	}

	// Convert processors to a DQL processing definition
	var processorDefs []string
	for _, p := range dd.Processors {
		def := convertLogProcessor(&p)
		if def != "" {
			processorDefs = append(processorDefs, def)
		}
	}

	if len(processorDefs) > 0 {
		rule.Processor = strings.Join(processorDefs, "\n| ")
	}

	return rule, nil
}

func translateLogFilter(ddFilter string) string {
	// Basic translation of DD log filter to DT log query
	// DD uses Lucene-like syntax, DT uses DQL
	filter := ddFilter

	// Replace common DD filter patterns
	filter = strings.ReplaceAll(filter, "source:", "dt.source.name == ")
	filter = strings.ReplaceAll(filter, "service:", "dt.entity.service == ")
	filter = strings.ReplaceAll(filter, "status:", "loglevel == ")

	return filter
}

func convertLogProcessor(p *datadog.LogProcessor) string {
	if !p.IsEnabled {
		return ""
	}

	switch p.Type {
	case "grok-parser":
		if p.Grok != nil {
			return fmt.Sprintf("PARSE(content, \"%s\")", escapeGrokPattern(p.Grok.MatchRules))
		}
	case "attribute-remapper":
		if len(p.Sources) > 0 && p.Target != "" {
			return fmt.Sprintf("FIELDS_RENAME(%s, %s)", p.Sources[0], p.Target)
		}
	case "date-remapper":
		if len(p.Sources) > 0 {
			return fmt.Sprintf("FIELDS_RENAME(%s, timestamp)", p.Sources[0])
		}
	case "status-remapper":
		if len(p.Sources) > 0 {
			return fmt.Sprintf("FIELDS_RENAME(%s, loglevel)", p.Sources[0])
		}
	case "message-remapper":
		if len(p.Sources) > 0 {
			return fmt.Sprintf("FIELDS_RENAME(%s, content)", p.Sources[0])
		}
	case "category-processor":
		return fmt.Sprintf("// Category processor '%s' - requires manual DQL configuration", p.Name)
	case "arithmetic-processor":
		return fmt.Sprintf("// Arithmetic processor '%s' - requires manual DQL configuration", p.Name)
	case "string-builder-processor":
		return fmt.Sprintf("// String builder '%s' - requires manual DQL configuration", p.Name)
	case "pipeline":
		return fmt.Sprintf("// Nested pipeline '%s' - requires manual DQL configuration", p.Name)
	}

	return fmt.Sprintf("// Unsupported processor type: %s (%s)", p.Type, p.Name)
}

func escapeGrokPattern(pattern string) string {
	// Basic escaping for DQL PARSE patterns
	pattern = strings.ReplaceAll(pattern, "\"", "\\\"")
	return pattern
}
