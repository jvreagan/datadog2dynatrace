package terraform

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

// GenerateDashboards generates Terraform HCL for Dynatrace dashboards.
func GenerateDashboards(dashboards []dynatrace.Dashboard) string {
	var sb strings.Builder
	sb.WriteString("# Dashboards - migrated from DataDog\n\n")

	for i, d := range dashboards {
		name := sanitizeTFName(d.DashboardMetadata.Name)
		sb.WriteString(fmt.Sprintf("resource \"dynatrace_json_dashboard\" \"%s\" {\n", uniqueName(name, i)))

		// Marshal the dashboard to JSON for the json_dashboard resource
		jsonBytes, err := json.MarshalIndent(d, "", "  ")
		if err != nil {
			sb.WriteString(fmt.Sprintf("  # Error marshaling dashboard: %s\n", err))
			sb.WriteString("}\n\n")
			continue
		}

		sb.WriteString(fmt.Sprintf("  contents = jsonencode(%s)\n", string(jsonBytes)))
		sb.WriteString("}\n\n")
	}

	return sb.String()
}
