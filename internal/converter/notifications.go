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

	default:
		return nil, fmt.Errorf("unsupported notification type: %s", dd.Type)
	}

	return ni, nil
}
