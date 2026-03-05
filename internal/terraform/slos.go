package terraform

import (
	"fmt"
	"strings"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

// GenerateSLOs generates Terraform HCL for Dynatrace SLOs.
func GenerateSLOs(slos []dynatrace.SLO) string {
	var sb strings.Builder
	sb.WriteString("# SLOs - migrated from DataDog\n\n")

	for i, s := range slos {
		name := sanitizeTFName(s.Name)
		sb.WriteString(fmt.Sprintf("resource \"dynatrace_slo_v2\" \"%s\" {\n", uniqueName(name, i)))
		sb.WriteString(fmt.Sprintf("  name              = %q\n", s.Name))
		sb.WriteString(fmt.Sprintf("  enabled           = %t\n", s.Enabled))
		sb.WriteString(fmt.Sprintf("  metric_expression = %q\n", s.MetricExpression))
		sb.WriteString(fmt.Sprintf("  evaluation_type   = %q\n", s.EvaluationType))
		sb.WriteString(fmt.Sprintf("  target            = %g\n", s.Target))
		sb.WriteString(fmt.Sprintf("  warning           = %g\n", s.Warning))
		sb.WriteString(fmt.Sprintf("  timeframe         = %q\n", s.Timeframe))

		if s.Description != "" {
			sb.WriteString(fmt.Sprintf("  description       = %q\n", s.Description))
		}

		sb.WriteString("}\n\n")
	}

	return sb.String()
}
