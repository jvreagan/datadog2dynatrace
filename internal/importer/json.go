package importer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
)

// ImportFromDirectory imports DataDog resources from exported JSON files in a directory.
func ImportFromDirectory(dir string) (*datadog.ExtractionResult, error) {
	result := &datadog.ExtractionResult{}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		name := strings.ToLower(entry.Name())

		switch {
		case strings.HasSuffix(name, ".tf.json") || strings.HasSuffix(name, ".tf"):
			if err := importTerraformFile(path, result); err != nil {
				return nil, fmt.Errorf("importing %s: %w", entry.Name(), err)
			}
		case strings.HasSuffix(name, ".json"):
			if err := importJSONFile(path, name, result); err != nil {
				return nil, fmt.Errorf("importing %s: %w", entry.Name(), err)
			}
		}
	}

	return result, nil
}

func importJSONFile(path, name string, result *datadog.ExtractionResult) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	switch {
	case strings.Contains(name, "dashboard"):
		return importDashboards(data, result)
	case strings.Contains(name, "monitor"):
		return importMonitors(data, result)
	case strings.Contains(name, "slo"):
		return importSLOs(data, result)
	case strings.Contains(name, "synthetic"):
		return importSynthetics(data, result)
	case strings.Contains(name, "log") && strings.Contains(name, "pipeline"):
		return importLogPipelines(data, result)
	case strings.Contains(name, "downtime"):
		return importDowntimes(data, result)
	case strings.Contains(name, "notebook"):
		return importNotebooks(data, result)
	case strings.Contains(name, "notification"):
		return importNotifications(data, result)
	default:
		// Try auto-detection based on content
		return autoImport(data, result)
	}
}

func importDashboards(data []byte, result *datadog.ExtractionResult) error {
	// Try array first
	var dashboards []datadog.Dashboard
	if err := json.Unmarshal(data, &dashboards); err == nil {
		result.Dashboards = append(result.Dashboards, dashboards...)
		return nil
	}

	// Try single object
	var dashboard datadog.Dashboard
	if err := json.Unmarshal(data, &dashboard); err == nil {
		result.Dashboards = append(result.Dashboards, dashboard)
		return nil
	}

	// Try wrapper format
	var wrapper struct {
		Dashboards []datadog.Dashboard `json:"dashboards"`
	}
	if err := json.Unmarshal(data, &wrapper); err == nil && len(wrapper.Dashboards) > 0 {
		result.Dashboards = append(result.Dashboards, wrapper.Dashboards...)
		return nil
	}

	return fmt.Errorf("could not parse dashboard JSON")
}

func importMonitors(data []byte, result *datadog.ExtractionResult) error {
	var monitors []datadog.Monitor
	if err := json.Unmarshal(data, &monitors); err == nil {
		result.Monitors = append(result.Monitors, monitors...)
		return nil
	}

	var monitor datadog.Monitor
	if err := json.Unmarshal(data, &monitor); err != nil {
		return fmt.Errorf("parsing monitors: %w", err)
	}
	result.Monitors = append(result.Monitors, monitor)
	return nil
}

func importSLOs(data []byte, result *datadog.ExtractionResult) error {
	// Try the API response format
	var resp struct {
		Data []datadog.SLO `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err == nil && len(resp.Data) > 0 {
		result.SLOs = append(result.SLOs, resp.Data...)
		return nil
	}

	var slos []datadog.SLO
	if err := json.Unmarshal(data, &slos); err == nil {
		result.SLOs = append(result.SLOs, slos...)
		return nil
	}

	var slo datadog.SLO
	if err := json.Unmarshal(data, &slo); err != nil {
		return fmt.Errorf("parsing SLOs: %w", err)
	}
	result.SLOs = append(result.SLOs, slo)
	return nil
}

func importSynthetics(data []byte, result *datadog.ExtractionResult) error {
	var resp struct {
		Tests []datadog.SyntheticTest `json:"tests"`
	}
	if err := json.Unmarshal(data, &resp); err == nil && len(resp.Tests) > 0 {
		result.Synthetics = append(result.Synthetics, resp.Tests...)
		return nil
	}

	var tests []datadog.SyntheticTest
	if err := json.Unmarshal(data, &tests); err == nil {
		result.Synthetics = append(result.Synthetics, tests...)
		return nil
	}

	var test datadog.SyntheticTest
	if err := json.Unmarshal(data, &test); err != nil {
		return fmt.Errorf("parsing synthetics: %w", err)
	}
	result.Synthetics = append(result.Synthetics, test)
	return nil
}

func importLogPipelines(data []byte, result *datadog.ExtractionResult) error {
	var pipelines []datadog.LogPipeline
	if err := json.Unmarshal(data, &pipelines); err == nil {
		result.LogPipelines = append(result.LogPipelines, pipelines...)
		return nil
	}

	var pipeline datadog.LogPipeline
	if err := json.Unmarshal(data, &pipeline); err != nil {
		return fmt.Errorf("parsing log pipelines: %w", err)
	}
	result.LogPipelines = append(result.LogPipelines, pipeline)
	return nil
}

func importDowntimes(data []byte, result *datadog.ExtractionResult) error {
	var downtimes []datadog.Downtime
	if err := json.Unmarshal(data, &downtimes); err == nil {
		result.Downtimes = append(result.Downtimes, downtimes...)
		return nil
	}

	var downtime datadog.Downtime
	if err := json.Unmarshal(data, &downtime); err != nil {
		return fmt.Errorf("parsing downtimes: %w", err)
	}
	result.Downtimes = append(result.Downtimes, downtime)
	return nil
}

func importNotebooks(data []byte, result *datadog.ExtractionResult) error {
	var notebooks []datadog.Notebook
	if err := json.Unmarshal(data, &notebooks); err == nil {
		result.Notebooks = append(result.Notebooks, notebooks...)
		return nil
	}

	var notebook datadog.Notebook
	if err := json.Unmarshal(data, &notebook); err != nil {
		return fmt.Errorf("parsing notebooks: %w", err)
	}
	result.Notebooks = append(result.Notebooks, notebook)
	return nil
}

func importNotifications(data []byte, result *datadog.ExtractionResult) error {
	var notifications []datadog.NotificationChannel
	if err := json.Unmarshal(data, &notifications); err == nil {
		result.Notifications = append(result.Notifications, notifications...)
		return nil
	}

	var notification datadog.NotificationChannel
	if err := json.Unmarshal(data, &notification); err != nil {
		return fmt.Errorf("parsing notifications: %w", err)
	}
	result.Notifications = append(result.Notifications, notification)
	return nil
}

// autoImport tries to detect the type of JSON data and import it.
func autoImport(data []byte, result *datadog.ExtractionResult) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil // Skip unrecognizable files
	}

	// Detect by key presence
	if _, ok := raw["widgets"]; ok {
		return importDashboards(data, result)
	}
	if _, ok := raw["query"]; ok {
		if _, ok := raw["type"]; ok {
			return importMonitors(data, result)
		}
	}
	if _, ok := raw["thresholds"]; ok {
		return importSLOs(data, result)
	}
	if _, ok := raw["public_id"]; ok {
		if _, ok2 := raw["type"]; ok2 {
			return importSynthetics(data, result)
		}
	}
	if _, ok := raw["processors"]; ok {
		if _, ok2 := raw["is_enabled"]; ok2 {
			return importLogPipelines(data, result)
		}
	}
	if _, ok := raw["scope"]; ok {
		if _, ok2 := raw["monitor_id"]; ok2 {
			return importDowntimes(data, result)
		}
		if _, ok2 := raw["monitor_tags"]; ok2 {
			return importDowntimes(data, result)
		}
	}
	if _, ok := raw["cells"]; ok {
		if _, ok2 := raw["author"]; ok2 {
			return importNotebooks(data, result)
		}
	}

	return nil // Unknown format, skip silently
}
