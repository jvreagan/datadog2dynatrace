package terraform

import (
	"fmt"
	"strings"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

// GenerateNotifications generates Terraform HCL for Dynatrace notification integrations.
func GenerateNotifications(notifications []dynatrace.NotificationIntegration) string {
	var sb strings.Builder
	sb.WriteString("# Notifications - migrated from DataDog\n\n")

	for i, n := range notifications {
		name := sanitizeTFName(n.Name)

		switch n.Type {
		case "SLACK":
			sb.WriteString(fmt.Sprintf("resource \"dynatrace_slack_notification\" \"%s\" {\n", uniqueName(name, i)))
			sb.WriteString(fmt.Sprintf("  name    = %q\n", n.Name))
			sb.WriteString(fmt.Sprintf("  active  = %t\n", n.Active))
			if url, ok := n.Config["url"]; ok {
				sb.WriteString(fmt.Sprintf("  url     = %q\n", url))
			}
			if ch, ok := n.Config["channel"]; ok {
				sb.WriteString(fmt.Sprintf("  channel = %q\n", ch))
			}

		case "EMAIL":
			sb.WriteString(fmt.Sprintf("resource \"dynatrace_email_notification\" \"%s\" {\n", uniqueName(name, i)))
			sb.WriteString(fmt.Sprintf("  name    = %q\n", n.Name))
			sb.WriteString(fmt.Sprintf("  active  = %t\n", n.Active))
			if emails, ok := n.Config["receivers"]; ok {
				sb.WriteString(fmt.Sprintf("  receivers = %q\n", emails))
			}

		case "WEBHOOK":
			sb.WriteString(fmt.Sprintf("resource \"dynatrace_webhook_notification\" \"%s\" {\n", uniqueName(name, i)))
			sb.WriteString(fmt.Sprintf("  name    = %q\n", n.Name))
			sb.WriteString(fmt.Sprintf("  active  = %t\n", n.Active))
			if url, ok := n.Config["url"]; ok {
				sb.WriteString(fmt.Sprintf("  url     = %q\n", url))
			}

		case "PAGER_DUTY":
			sb.WriteString(fmt.Sprintf("resource \"dynatrace_pagerduty_notification\" \"%s\" {\n", uniqueName(name, i)))
			sb.WriteString(fmt.Sprintf("  name    = %q\n", n.Name))
			sb.WriteString(fmt.Sprintf("  active  = %t\n", n.Active))
			if acct, ok := n.Config["account"]; ok {
				sb.WriteString(fmt.Sprintf("  account = %q\n", acct))
			}

		default:
			sb.WriteString(fmt.Sprintf("# Unsupported notification type: %s\n", n.Type))
			sb.WriteString(fmt.Sprintf("# resource \"dynatrace_notification\" \"%s\" { ... }\n\n", uniqueName(name, i)))
			continue
		}

		sb.WriteString("}\n\n")
	}

	return sb.String()
}

// Helper functions used by all terraform generators

func sanitizeTFName(name string) string {
	// Replace non-alphanumeric characters with underscores
	var sb strings.Builder
	for _, c := range strings.ToLower(name) {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			sb.WriteRune(c)
		} else {
			sb.WriteRune('_')
		}
	}
	result := sb.String()
	// Remove leading underscores and consecutive underscores
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}
	result = strings.Trim(result, "_")
	if result == "" {
		result = "resource"
	}
	// Ensure it starts with a letter
	if result[0] >= '0' && result[0] <= '9' {
		result = "r_" + result
	}
	return result
}

func uniqueName(base string, index int) string {
	if index == 0 {
		return base
	}
	return fmt.Sprintf("%s_%d", base, index)
}
