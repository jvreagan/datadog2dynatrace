package converter

import (
	"fmt"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

// ConvertNotification converts a DataDog notification channel to a Dynatrace notification integration.
func ConvertNotification(dd *datadog.NotificationChannel) (*dynatrace.NotificationIntegration, error) {
	ni := &dynatrace.NotificationIntegration{
		Name:   dd.Name,
		Active: true,
		Config: make(map[string]interface{}),
	}

	switch dd.Type {
	case "slack":
		ni.Type = "SLACK"
		if url, ok := dd.Config["url"]; ok {
			ni.Config["url"] = url
		}
		if channel, ok := dd.Config["channel"]; ok {
			ni.Config["channel"] = channel
		}

	case "pagerduty":
		ni.Type = "PAGER_DUTY"
		if name, ok := dd.Config["service_name"]; ok {
			ni.Config["account"] = name
		}
		if key, ok := dd.Config["service_key"]; ok {
			ni.Config["integrationKey"] = key
		}

	case "email":
		ni.Type = "EMAIL"
		if emails, ok := dd.Config["emails"]; ok {
			ni.Config["receivers"] = emails
		}

	case "webhook":
		ni.Type = "WEBHOOK"
		if url, ok := dd.Config["url"]; ok {
			ni.Config["url"] = url
		}
		if payload, ok := dd.Config["payload"]; ok {
			ni.Config["payload"] = payload
		}

	case "opsgenie":
		ni.Type = "OPS_GENIE"
		if key, ok := dd.Config["api_key"]; ok {
			ni.Config["apiKey"] = key
		}

	case "victorops":
		ni.Type = "VICTOR_OPS"
		if key, ok := dd.Config["api_key"]; ok {
			ni.Config["apiKey"] = key
		}

	case "microsoft-teams", "msteams":
		ni.Type = "MICROSOFT_TEAMS"
		if url, ok := dd.Config["url"]; ok {
			ni.Config["url"] = url
		}

	case "xmatters":
		ni.Type = "XMATTERS"
		if url, ok := dd.Config["url"]; ok {
			ni.Config["url"] = url
		}

	case "jira":
		ni.Type = "JIRA"
		if url, ok := dd.Config["url"]; ok {
			ni.Config["url"] = url
		}
		if user, ok := dd.Config["username"]; ok {
			ni.Config["username"] = user
		}
		if pass, ok := dd.Config["password"]; ok {
			ni.Config["password"] = pass
		}
		if proj, ok := dd.Config["project_key"]; ok {
			ni.Config["projectKey"] = proj
		}
		if issue, ok := dd.Config["issue_type"]; ok {
			ni.Config["issueType"] = issue
		}

	case "servicenow":
		ni.Type = "SERVICE_NOW"
		if url, ok := dd.Config["url"]; ok {
			ni.Config["url"] = url
		}
		if user, ok := dd.Config["username"]; ok {
			ni.Config["username"] = user
		}
		if pass, ok := dd.Config["password"]; ok {
			ni.Config["password"] = pass
		}

	default:
		return nil, fmt.Errorf("unsupported notification type: %s", dd.Type)
	}

	return ni, nil
}
