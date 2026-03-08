package terraform

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

// GenerateNotebooks generates Terraform HCL for Dynatrace notebooks (documents).
func GenerateNotebooks(notebooks []dynatrace.DynatraceNotebook) string {
	var sb strings.Builder
	sb.WriteString("# Notebooks - migrated from DataDog\n\n")

	for i, nb := range notebooks {
		name := sanitizeTFName(nb.Name)
		sb.WriteString(fmt.Sprintf("resource \"dynatrace_document\" \"%s\" {\n", uniqueName(name, i)))
		sb.WriteString(fmt.Sprintf("  name = %q\n", nb.Name))
		sb.WriteString("  type = \"notebook\"\n")

		contentBytes, err := json.Marshal(map[string]interface{}{
			"sections": nb.Sections,
		})
		if err != nil {
			sb.WriteString(fmt.Sprintf("  # Error marshaling notebook content: %s\n", err))
		} else {
			sb.WriteString(fmt.Sprintf("  content = jsonencode(%s)\n", string(contentBytes)))
		}

		sb.WriteString("}\n\n")
	}

	return sb.String()
}
