package importer

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
)

// importTerraformFile imports DataDog resources from a Terraform file.
// Supports .tf.json format (JSON-based Terraform configs).
func importTerraformFile(path string, result *datadog.ExtractionResult) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	// For .tf.json files, parse the JSON Terraform format
	if strings.HasSuffix(path, ".tf.json") {
		return importTFJSON(data, result)
	}

	// For .tf files (HCL), do a best-effort extraction of resource blocks
	return importTFHCL(string(data), result)
}

func importTFJSON(data []byte, result *datadog.ExtractionResult) error {
	var tfConfig struct {
		Resource map[string]map[string]json.RawMessage `json:"resource"`
	}
	if err := json.Unmarshal(data, &tfConfig); err != nil {
		return fmt.Errorf("parsing Terraform JSON: %w", err)
	}

	for resType, resources := range tfConfig.Resource {
		for _, rawRes := range resources {
			switch {
			case strings.Contains(resType, "datadog_dashboard"):
				var d datadog.Dashboard
				if err := json.Unmarshal(rawRes, &d); err == nil {
					result.Dashboards = append(result.Dashboards, d)
				}
			case strings.Contains(resType, "datadog_monitor"):
				var m datadog.Monitor
				if err := json.Unmarshal(rawRes, &m); err == nil {
					result.Monitors = append(result.Monitors, m)
				}
			case strings.Contains(resType, "datadog_service_level_objective"):
				var s datadog.SLO
				if err := json.Unmarshal(rawRes, &s); err == nil {
					result.SLOs = append(result.SLOs, s)
				}
			case strings.Contains(resType, "datadog_synthetics_test"):
				var s datadog.SyntheticTest
				if err := json.Unmarshal(rawRes, &s); err == nil {
					result.Synthetics = append(result.Synthetics, s)
				}
			case strings.Contains(resType, "datadog_logs_custom_pipeline"):
				var l datadog.LogPipeline
				if err := json.Unmarshal(rawRes, &l); err == nil {
					result.LogPipelines = append(result.LogPipelines, l)
				}
			case strings.Contains(resType, "datadog_downtime"):
				var d datadog.Downtime
				if err := json.Unmarshal(rawRes, &d); err == nil {
					result.Downtimes = append(result.Downtimes, d)
				}
			case strings.Contains(resType, "datadog_notebook"):
				var nb datadog.Notebook
				if err := json.Unmarshal(rawRes, &nb); err == nil {
					result.Notebooks = append(result.Notebooks, nb)
				}
			}
		}
	}

	return nil
}

// importTFHCL does a basic extraction from HCL .tf files.
// This is a simplified parser that looks for resource blocks and extracts key fields.
// For full HCL parsing, users should export to .tf.json format first.
func importTFHCL(content string, result *datadog.ExtractionResult) error {
	// Look for datadog resource blocks and extract names
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "resource ") {
			continue
		}

		// Extract resource type
		parts := strings.Fields(trimmed)
		if len(parts) < 3 {
			continue
		}

		resType := strings.Trim(parts[1], "\"")
		_ = i // line number for future error reporting

		switch {
		case strings.Contains(resType, "datadog_dashboard"):
			name := extractHCLStringField(lines[i:], "title")
			if name != "" {
				result.Dashboards = append(result.Dashboards, datadog.Dashboard{Title: name})
			}
		case strings.Contains(resType, "datadog_monitor"):
			name := extractHCLStringField(lines[i:], "name")
			query := extractHCLStringField(lines[i:], "query")
			monType := extractHCLStringField(lines[i:], "type")
			result.Monitors = append(result.Monitors, datadog.Monitor{
				Name:  name,
				Query: query,
				Type:  monType,
			})
		case strings.Contains(resType, "datadog_service_level_objective"):
			name := extractHCLStringField(lines[i:], "name")
			if name != "" {
				result.SLOs = append(result.SLOs, datadog.SLO{Name: name})
			}
		case strings.Contains(resType, "datadog_synthetics_test"):
			name := extractHCLStringField(lines[i:], "name")
			sType := extractHCLStringField(lines[i:], "type")
			if name != "" {
				result.Synthetics = append(result.Synthetics, datadog.SyntheticTest{Name: name, Type: sType})
			}
		case strings.Contains(resType, "datadog_logs_custom_pipeline"):
			name := extractHCLStringField(lines[i:], "name")
			if name != "" {
				result.LogPipelines = append(result.LogPipelines, datadog.LogPipeline{Name: name})
			}
		case strings.Contains(resType, "datadog_downtime"):
			message := extractHCLStringField(lines[i:], "message")
			result.Downtimes = append(result.Downtimes, datadog.Downtime{Message: message})
		}
	}

	return nil
}

func extractHCLStringField(lines []string, field string) string {
	depth := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "{") {
			depth++
		}
		if strings.Contains(trimmed, "}") {
			depth--
			if depth <= 0 {
				break
			}
		}
		if depth == 1 && strings.HasPrefix(trimmed, field) {
			// Extract value after = sign
			if idx := strings.Index(trimmed, "="); idx >= 0 {
				val := strings.TrimSpace(trimmed[idx+1:])
				val = strings.Trim(val, "\"")
				return val
			}
		}
	}
	return ""
}
