package terraform

import (
	"fmt"
	"strings"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

// GenerateLogProcessing generates Terraform HCL for Dynatrace log processing rules.
func GenerateLogProcessing(rules []dynatrace.LogProcessingRule) string {
	var sb strings.Builder
	sb.WriteString("# Log Processing Rules - migrated from DataDog Log Pipelines\n\n")

	for i, r := range rules {
		name := sanitizeTFName(r.Name)
		sb.WriteString(fmt.Sprintf("resource \"dynatrace_log_processing\" \"%s\" {\n", uniqueName(name, i)))
		sb.WriteString(fmt.Sprintf("  name      = %q\n", r.Name))
		sb.WriteString(fmt.Sprintf("  enabled   = %t\n", r.Enabled))
		sb.WriteString(fmt.Sprintf("  query     = %q\n", r.Query))

		if r.Processor != "" {
			sb.WriteString(fmt.Sprintf("  processor = %q\n", r.Processor))
		}

		sb.WriteString("}\n\n")
	}

	return sb.String()
}
