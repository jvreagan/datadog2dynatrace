package terraform

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

// GenerateSynthetics generates Terraform HCL for Dynatrace synthetic monitors.
func GenerateSynthetics(monitors []dynatrace.SyntheticMonitor) string {
	var sb strings.Builder
	sb.WriteString("# Synthetic Monitors - migrated from DataDog\n\n")

	for i, sm := range monitors {
		name := sanitizeTFName(sm.Name)
		resType := "dynatrace_http_monitor"
		if sm.Type == "BROWSER" {
			resType = "dynatrace_browser_monitor"
		}

		sb.WriteString(fmt.Sprintf("resource %q \"%s\" {\n", resType, uniqueName(name, i)))
		sb.WriteString(fmt.Sprintf("  name      = %q\n", sm.Name))
		sb.WriteString(fmt.Sprintf("  enabled   = %t\n", sm.Enabled))
		sb.WriteString(fmt.Sprintf("  frequency = %d\n", sm.FrequencyMin))

		if len(sm.Locations) > 0 {
			locJSON, _ := json.Marshal(sm.Locations)
			sb.WriteString(fmt.Sprintf("  locations = %s\n", string(locJSON)))
		}

		if sm.Script != nil {
			scriptJSON, err := json.MarshalIndent(sm.Script, "    ", "  ")
			if err == nil {
				sb.WriteString(fmt.Sprintf("  script = jsonencode(%s)\n", string(scriptJSON)))
			}
		}

		if sm.AnomalyDetection != nil {
			sb.WriteString("\n  anomaly_detection {\n")
			if sm.AnomalyDetection.OutageHandling != nil {
				oh := sm.AnomalyDetection.OutageHandling
				sb.WriteString("    outage_handling {\n")
				sb.WriteString(fmt.Sprintf("      global_outage  = %t\n", oh.GlobalOutage))
				sb.WriteString(fmt.Sprintf("      local_outage   = %t\n", oh.LocalOutage))
				sb.WriteString(fmt.Sprintf("      retry_on_error = %t\n", oh.RetryOnError))
				sb.WriteString("    }\n")
			}
			sb.WriteString("  }\n")
		}

		sb.WriteString("}\n\n")
	}

	return sb.String()
}
