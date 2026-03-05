package converter

import (
	"fmt"
	"strings"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

// ConvertSynthetic converts a DataDog synthetic test to a Dynatrace synthetic monitor.
func ConvertSynthetic(dd *datadog.SyntheticTest) (*dynatrace.SyntheticMonitor, error) {
	sm := &dynatrace.SyntheticMonitor{
		Name:         dd.Name,
		Enabled:      dd.Status == "live",
		FrequencyMin: mapFrequency(dd.Options.TickEvery),
		Locations:    mapLocations(dd.Locations),
	}

	// Map tags
	for _, tag := range dd.Tags {
		parts := strings.SplitN(tag, ":", 2)
		t := dynatrace.SyntheticTag{Key: parts[0]}
		if len(parts) == 2 {
			t.Value = parts[1]
		}
		sm.Tags = append(sm.Tags, t)
	}

	switch dd.Type {
	case "api":
		return convertAPISynthetic(dd, sm)
	case "browser":
		return convertBrowserSynthetic(dd, sm)
	default:
		return nil, fmt.Errorf("unsupported synthetic test type: %s", dd.Type)
	}
}

func convertAPISynthetic(dd *datadog.SyntheticTest, sm *dynatrace.SyntheticMonitor) (*dynatrace.SyntheticMonitor, error) {
	sm.Type = "HTTP"

	if dd.Config.Request == nil {
		return nil, fmt.Errorf("API synthetic test has no request configuration")
	}

	req := dd.Config.Request
	scriptReq := dynatrace.ScriptRequest{
		Description: dd.Name,
		URL:         req.URL,
		Method:      req.Method,
		RequestBody: req.Body,
	}

	if len(req.Headers) > 0 {
		scriptReq.Configuration = &dynatrace.RequestConfiguration{
			Headers:         req.Headers,
			FollowRedirects: dd.Options.FollowRedirects,
		}
	}

	// Convert assertions to validation rules
	if len(dd.Config.Assertions) > 0 {
		var rules []dynatrace.ValidationRule
		for _, a := range dd.Config.Assertions {
			rule := convertAssertion(&a)
			if rule != nil {
				rules = append(rules, *rule)
			}
		}
		if len(rules) > 0 {
			scriptReq.Validation = &dynatrace.RequestValidation{Rules: rules}
		}
	}

	sm.Script = &dynatrace.SyntheticScript{
		Version:  "1.0",
		Type:     "availability",
		Requests: []dynatrace.ScriptRequest{scriptReq},
	}

	// Set anomaly detection
	sm.AnomalyDetection = &dynatrace.AnomalyDetection{
		OutageHandling: &dynatrace.OutageHandling{
			GlobalOutage: true,
			LocalOutage:  true,
			RetryOnError: dd.Options.Retry != nil && dd.Options.Retry.Count > 0,
			LocalOutagePolicy: &dynatrace.LocalOutagePolicy{
				AffectedLocations: maxInt(dd.Options.MinLocFailed, 1),
				ConsecutiveRuns:   3,
			},
		},
	}

	return sm, nil
}

func convertBrowserSynthetic(dd *datadog.SyntheticTest, sm *dynatrace.SyntheticMonitor) (*dynatrace.SyntheticMonitor, error) {
	sm.Type = "BROWSER"

	sm.KeyPerformanceMetrics = &dynatrace.KeyPerformanceMetrics{
		LoadActionKPM: "VISUALLY_COMPLETE",
		XHRActionKPM:  "VISUALLY_COMPLETE",
	}

	var events []dynatrace.ScriptEvent

	// Add navigate event for the URL
	if dd.Config.Request != nil {
		events = append(events, dynatrace.ScriptEvent{
			Type:        "navigate",
			Description: fmt.Sprintf("Navigate to %s", dd.Config.Request.URL),
			URL:         dd.Config.Request.URL,
		})
	}

	// Convert browser steps
	for _, step := range dd.Config.Steps {
		event := convertBrowserStep(&step)
		if event != nil {
			events = append(events, *event)
		}
	}

	sm.Script = &dynatrace.SyntheticScript{
		Version: "1.0",
		Type:    "clickpath",
		Events:  events,
	}

	return sm, nil
}

func convertAssertion(a *datadog.SyntheticAssertion) *dynatrace.ValidationRule {
	switch a.Type {
	case "statusCode":
		return &dynatrace.ValidationRule{
			Type:        "httpStatusesList",
			Value:       fmt.Sprintf("%v", a.Target),
			PassIfFound: true,
		}
	case "body":
		return &dynatrace.ValidationRule{
			Type:        "patternConstraint",
			Value:       fmt.Sprintf("%v", a.Target),
			PassIfFound: a.Operator == "contains",
		}
	case "responseTime":
		return &dynatrace.ValidationRule{
			Type:        "patternConstraint",
			Value:       fmt.Sprintf("%v", a.Target),
			PassIfFound: true,
		}
	default:
		return nil
	}
}

func convertBrowserStep(step *datadog.BrowserStep) *dynatrace.ScriptEvent {
	switch step.Type {
	case "click":
		return &dynatrace.ScriptEvent{
			Type:        "click",
			Description: step.Name,
			Target: &dynatrace.EventTarget{
				Locators: []dynatrace.Locator{
					{Type: "css", Value: getStringParam(step.Params, "element")},
				},
			},
		}
	case "typeText":
		return &dynatrace.ScriptEvent{
			Type:        "keystrokes",
			Description: step.Name,
			Target: &dynatrace.EventTarget{
				Locators: []dynatrace.Locator{
					{Type: "css", Value: getStringParam(step.Params, "element")},
				},
			},
		}
	case "wait":
		return &dynatrace.ScriptEvent{
			Type:        "javascript",
			Description: step.Name,
			Wait: &dynatrace.EventWait{
				WaitFor:      "page_complete",
				Milliseconds: 3000,
			},
		}
	case "assertElementPresent":
		return &dynatrace.ScriptEvent{
			Type:        "javascript",
			Description: fmt.Sprintf("Assert: %s", step.Name),
		}
	default:
		return &dynatrace.ScriptEvent{
			Type:        "javascript",
			Description: fmt.Sprintf("[Migrated] %s (type: %s)", step.Name, step.Type),
		}
	}
}

func getStringParam(params map[string]interface{}, key string) string {
	if v, ok := params[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func mapFrequency(tickEvery int) int {
	// DD uses seconds, DT uses minutes
	if tickEvery <= 0 {
		return 15 // default 15 min
	}
	mins := tickEvery / 60
	if mins < 1 {
		return 1
	}
	// DT supports: 1, 2, 5, 10, 15, 30, 60
	validFreqs := []int{1, 2, 5, 10, 15, 30, 60}
	for _, f := range validFreqs {
		if mins <= f {
			return f
		}
	}
	return 60
}

func mapLocations(ddLocations []string) []string {
	// Map DD location identifiers to DT location IDs
	locationMap := map[string]string{
		"aws:us-east-1":      "GEOLOCATION-0A41430434C388A9", // N. Virginia
		"aws:us-west-1":      "GEOLOCATION-95196F3C9A4F4215", // N. California
		"aws:us-west-2":      "GEOLOCATION-D8D1B2B891DE92B8", // Oregon
		"aws:eu-west-1":      "GEOLOCATION-6D3D516E6E02ECFE", // Ireland
		"aws:eu-central-1":   "GEOLOCATION-F262DFF465079C68", // Frankfurt
		"aws:ap-northeast-1": "GEOLOCATION-B0E1B3B5C8F59547", // Tokyo
		"aws:ap-southeast-1": "GEOLOCATION-A8D034E12B17A49C", // Singapore
		"aws:ap-southeast-2": "GEOLOCATION-F2E296F4E5FE94A8", // Sydney
	}

	var dtLocations []string
	for _, loc := range ddLocations {
		if dt, ok := locationMap[loc]; ok {
			dtLocations = append(dtLocations, dt)
		}
	}

	// If no locations could be mapped, use a default
	if len(dtLocations) == 0 {
		dtLocations = []string{"GEOLOCATION-0A41430434C388A9"} // N. Virginia
	}

	return dtLocations
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
