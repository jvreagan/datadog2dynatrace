package terraform

import (
	"fmt"
	"strings"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

// GenerateMaintenance generates Terraform HCL for Dynatrace maintenance windows.
func GenerateMaintenance(windows []dynatrace.MaintenanceWindow) string {
	var sb strings.Builder
	sb.WriteString("# Maintenance Windows - migrated from DataDog Downtimes\n\n")

	for i, mw := range windows {
		name := sanitizeTFName(mw.Name)
		sb.WriteString(fmt.Sprintf("resource \"dynatrace_maintenance\" \"%s\" {\n", uniqueName(name, i)))
		sb.WriteString(fmt.Sprintf("  name        = %q\n", mw.Name))
		sb.WriteString(fmt.Sprintf("  type        = %q\n", mw.Type))
		sb.WriteString(fmt.Sprintf("  suppression = %q\n", mw.Suppression))

		if mw.Description != "" {
			sb.WriteString(fmt.Sprintf("  description = %q\n", mw.Description))
		}

		sb.WriteString("\n  schedule {\n")
		sb.WriteString(fmt.Sprintf("    type    = %q\n", mw.Schedule.RecurrenceType))
		sb.WriteString(fmt.Sprintf("    start   = %q\n", mw.Schedule.Start))
		sb.WriteString(fmt.Sprintf("    end     = %q\n", mw.Schedule.End))
		sb.WriteString(fmt.Sprintf("    zone_id = %q\n", mw.Schedule.ZoneID))
		sb.WriteString("  }\n")

		if mw.Scope != nil && len(mw.Scope.Matches) > 0 {
			sb.WriteString("\n  filter {\n")
			for _, m := range mw.Scope.Matches {
				for _, t := range m.Tags {
					if t.Value != "" {
						sb.WriteString(fmt.Sprintf("    tag = \"%s:%s\"\n", t.Key, t.Value))
					} else {
						sb.WriteString(fmt.Sprintf("    tag = %q\n", t.Key))
					}
				}
			}
			sb.WriteString("  }\n")
		}

		sb.WriteString("}\n\n")
	}

	return sb.String()
}
