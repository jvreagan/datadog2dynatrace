package datadog

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/ratelimit"
)

func fastConfig() ratelimit.Config {
	return ratelimit.Config{
		RequestsPerSecond: 1000,
		MaxRetries:        0,
		InitialBackoff:    1 * time.Millisecond,
		MaxBackoff:        10 * time.Millisecond,
	}
}

func testClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := newClientWithConfig("test-api-key", "test-app-key", srv.URL, fastConfig())
	return c, srv
}

func TestValidateSuccess(t *testing.T) {
	var gotMethod, gotPath string

	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"valid":true}`))
	})

	err := c.Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "GET" {
		t.Errorf("expected GET, got %s", gotMethod)
	}
	if gotPath != "/api/v1/validate" {
		t.Errorf("expected /api/v1/validate, got %s", gotPath)
	}
}

func TestValidateFailure(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"errors":["invalid key"]}`))
	})

	err := c.Validate()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected 401 in error, got %q", err.Error())
	}
}

func TestValidateHeaders(t *testing.T) {
	var gotAPIKey, gotAppKey string

	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAPIKey = r.Header.Get("DD-API-KEY")
		gotAppKey = r.Header.Get("DD-APPLICATION-KEY")
		w.WriteHeader(http.StatusOK)
	})

	c.Validate()
	if gotAPIKey != "test-api-key" {
		t.Errorf("expected DD-API-KEY 'test-api-key', got %q", gotAPIKey)
	}
	if gotAppKey != "test-app-key" {
		t.Errorf("expected DD-APPLICATION-KEY 'test-app-key', got %q", gotAppKey)
	}
}

func TestGetDashboardList(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/dashboard" {
			w.Write([]byte(`{"dashboards":[{"id":"abc-123","title":"Test Dash"}]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	dashboards, err := c.GetDashboardList()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dashboards) != 1 {
		t.Fatalf("expected 1 dashboard, got %d", len(dashboards))
	}
	if dashboards[0].ID != "abc-123" {
		t.Errorf("expected ID abc-123, got %q", dashboards[0].ID)
	}
}

func TestGetDashboard(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/dashboard/abc-123" {
			w.Write([]byte(`{"id":"abc-123","title":"My Dashboard","widgets":[]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	dash, err := c.GetDashboard("abc-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dash.Title != "My Dashboard" {
		t.Errorf("expected title 'My Dashboard', got %q", dash.Title)
	}
}

func TestGetMonitors(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/monitor" {
			w.Write([]byte(`[{"id":1,"name":"High CPU","type":"metric alert","query":"avg:system.cpu.user{*} > 90"}]`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	monitors, err := c.GetMonitors()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(monitors) != 1 {
		t.Fatalf("expected 1 monitor, got %d", len(monitors))
	}
	if monitors[0].Name != "High CPU" {
		t.Errorf("expected 'High CPU', got %q", monitors[0].Name)
	}
}

func TestGetSLOs(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/slo" {
			w.Write([]byte(`{"data":[{"id":"slo-1","name":"API Availability","type":"metric"}]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	slos, err := c.GetSLOs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(slos) != 1 {
		t.Fatalf("expected 1 SLO, got %d", len(slos))
	}
	if slos[0].Name != "API Availability" {
		t.Errorf("expected 'API Availability', got %q", slos[0].Name)
	}
}

func TestGetSynthetics(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/synthetics/tests" {
			w.Write([]byte(`{"tests":[{"public_id":"syn-1","name":"Health Check","type":"api","status":"live"}]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	synthetics, err := c.GetSynthetics()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(synthetics) != 1 {
		t.Fatalf("expected 1 synthetic, got %d", len(synthetics))
	}
	if synthetics[0].Name != "Health Check" {
		t.Errorf("expected 'Health Check', got %q", synthetics[0].Name)
	}
}

func TestGetLogPipelines(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/logs/config/pipelines" {
			w.Write([]byte(`[{"id":"pipe-1","name":"JSON Parser","is_enabled":true}]`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	pipelines, err := c.GetLogPipelines()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pipelines) != 1 {
		t.Fatalf("expected 1 pipeline, got %d", len(pipelines))
	}
	if pipelines[0].Name != "JSON Parser" {
		t.Errorf("expected 'JSON Parser', got %q", pipelines[0].Name)
	}
}

func TestGetDowntimes(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/downtime" {
			w.Write([]byte(`[{"id":100,"message":"Deploy window","scope":["*"]}]`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	downtimes, err := c.GetDowntimes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(downtimes) != 1 {
		t.Fatalf("expected 1 downtime, got %d", len(downtimes))
	}
	if downtimes[0].Message != "Deploy window" {
		t.Errorf("expected 'Deploy window', got %q", downtimes[0].Message)
	}
}

func TestGetNotebooks(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/notebooks") {
			w.Write([]byte(`{"data":[{"id":1,"attributes":{"name":"My Notebook","cells":[]}}]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	notebooks, err := c.GetNotebooks()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(notebooks) != 1 {
		t.Fatalf("expected 1 notebook, got %d", len(notebooks))
	}
	if notebooks[0].Name != "My Notebook" {
		t.Errorf("expected 'My Notebook', got %q", notebooks[0].Name)
	}
}

func TestGetNotificationChannels(t *testing.T) {
	t.Run("all endpoints", func(t *testing.T) {
		c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.HasPrefix(r.URL.Path, "/api/v1/integration/webhooks"):
				w.Write([]byte(`[{"name":"Alert Hook","url":"https://hooks.example.com/alert","payload":"{\"msg\":\"alert\"}"}]`))
			case r.URL.Path == "/api/v1/integration/slack/configuration/accounts":
				w.Write([]byte(`[{"name":"engineering"}]`))
			case r.URL.Path == "/api/v1/integration/slack/configuration/channels/engineering":
				w.Write([]byte(`[{"channel_name":"#alerts","webhook_url":"https://hooks.slack.com/services/T00/B00/xxx"}]`))
			case strings.HasPrefix(r.URL.Path, "/api/v1/integration/pagerduty"):
				w.Write([]byte(`[{"service_name":"Prod Oncall","service_key":"pd-key-123"}]`))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		})

		channels, err := c.GetNotificationChannels()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(channels) != 3 {
			t.Fatalf("expected 3 channels, got %d", len(channels))
		}

		// Webhook
		if channels[0].Name != "Alert Hook" {
			t.Errorf("webhook name: got %q, want %q", channels[0].Name, "Alert Hook")
		}
		if channels[0].Type != "webhook" {
			t.Errorf("webhook type: got %q, want %q", channels[0].Type, "webhook")
		}

		// Slack (with channel-level data)
		if channels[1].Type != "slack" {
			t.Errorf("slack type: got %q, want %q", channels[1].Type, "slack")
		}
		if channels[1].Config["url"] != "https://hooks.slack.com/services/T00/B00/xxx" {
			t.Errorf("slack url: got %v", channels[1].Config["url"])
		}
		if channels[1].Config["channel"] != "#alerts" {
			t.Errorf("slack channel: got %v", channels[1].Config["channel"])
		}

		// PagerDuty
		if channels[2].Name != "Prod Oncall" {
			t.Errorf("pagerduty name: got %q, want %q", channels[2].Name, "Prod Oncall")
		}
		if channels[2].Config["service_key"] != "pd-key-123" {
			t.Errorf("pagerduty service_key: got %v", channels[2].Config["service_key"])
		}
		if channels[2].Config["service_name"] != "Prod Oncall" {
			t.Errorf("pagerduty service_name: got %v", channels[2].Config["service_name"])
		}
	})

	t.Run("slack channel endpoint fallback", func(t *testing.T) {
		c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == "/api/v1/integration/slack/configuration/accounts":
				w.Write([]byte(`[{"name":"legacy-workspace"}]`))
			case strings.HasPrefix(r.URL.Path, "/api/v1/integration/slack/configuration/channels/"):
				w.WriteHeader(http.StatusForbidden)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		})

		channels, err := c.GetNotificationChannels()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(channels) != 1 {
			t.Fatalf("expected 1 channel, got %d", len(channels))
		}
		if channels[0].Name != "legacy-workspace" {
			t.Errorf("expected fallback name %q, got %q", "legacy-workspace", channels[0].Name)
		}
		if channels[0].Config["account"] != "legacy-workspace" {
			t.Errorf("expected account in config, got %v", channels[0].Config["account"])
		}
		// Should NOT have url/channel since channel endpoint failed
		if _, ok := channels[0].Config["url"]; ok {
			t.Error("expected no url in fallback config")
		}
	})
}

func TestExtractAllSuccess(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/dashboard":
			w.Write([]byte(`{"dashboards":[{"id":"d1","title":"Dash"}]}`))
		case strings.HasPrefix(r.URL.Path, "/api/v1/dashboard/"):
			w.Write([]byte(`{"id":"d1","title":"Dash","widgets":[]}`))
		case r.URL.Path == "/api/v1/monitor":
			w.Write([]byte(`[{"id":1,"name":"Mon","type":"metric alert","query":"avg:system.cpu.user{*} > 90"}]`))
		case r.URL.Path == "/api/v1/slo":
			w.Write([]byte(`{"data":[{"id":"s1","name":"SLO","type":"metric"}]}`))
		case r.URL.Path == "/api/v1/synthetics/tests":
			w.Write([]byte(`{"tests":[{"public_id":"t1","name":"Syn","type":"api","status":"live"}]}`))
		case r.URL.Path == "/api/v1/logs/config/pipelines":
			w.Write([]byte(`[{"id":"p1","name":"Log","is_enabled":true}]`))
		case r.URL.Path == "/api/v1/downtime":
			w.Write([]byte(`[{"id":1,"message":"Down","scope":["*"]}]`))
		case strings.HasPrefix(r.URL.Path, "/api/v1/notebooks"):
			w.Write([]byte(`{"data":[{"id":1,"attributes":{"name":"NB","cells":[]}}]}`))
		case strings.HasPrefix(r.URL.Path, "/api/v1/integration/webhooks"):
			w.Write([]byte(`[{"name":"Hook","url":"https://example.com/hook"}]`))
		case r.URL.Path == "/api/v1/integration/slack/configuration/accounts":
			w.Write([]byte(`[{"name":"team"}]`))
		case r.URL.Path == "/api/v1/integration/slack/configuration/channels/team":
			w.Write([]byte(`[{"channel_name":"#alerts","webhook_url":"https://hooks.slack.com/test"}]`))
		case strings.HasPrefix(r.URL.Path, "/api/v1/integration/pagerduty"):
			w.Write([]byte(`[{"service_name":"PD","service_key":"key-1"}]`))
		default:
			w.Write([]byte(`[]`))
		}
	})

	result, err := c.ExtractAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Dashboards) != 1 {
		t.Errorf("expected 1 dashboard, got %d", len(result.Dashboards))
	}
	if len(result.Monitors) != 1 {
		t.Errorf("expected 1 monitor, got %d", len(result.Monitors))
	}
	if len(result.SLOs) != 1 {
		t.Errorf("expected 1 SLO, got %d", len(result.SLOs))
	}
	if len(result.Notifications) != 3 {
		t.Errorf("expected 3 notifications (webhook+slack+pagerduty), got %d", len(result.Notifications))
	}
}

func TestExtractAllPartialFailure(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/dashboard":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("server error"))
		case r.URL.Path == "/api/v1/monitor":
			w.Write([]byte(`[{"id":1,"name":"Mon","type":"metric alert","query":"avg:system.cpu.user{*} > 90"}]`))
		case r.URL.Path == "/api/v1/slo":
			w.Write([]byte(`{"data":[]}`))
		case r.URL.Path == "/api/v1/synthetics/tests":
			w.Write([]byte(`{"tests":[]}`))
		case r.URL.Path == "/api/v1/logs/config/pipelines":
			w.Write([]byte(`[]`))
		case r.URL.Path == "/api/v1/downtime":
			w.Write([]byte(`[]`))
		case strings.HasPrefix(r.URL.Path, "/api/v1/notebooks"):
			w.Write([]byte(`{"data":[]}`))
		default:
			w.Write([]byte(`[]`))
		}
	})

	result, err := c.ExtractAll()
	// Should have error from dashboards but continue with monitors
	if err == nil {
		t.Fatal("expected error from partial failure")
	}
	if len(result.Monitors) != 1 {
		t.Errorf("expected monitors to be fetched despite dashboard failure, got %d", len(result.Monitors))
	}
}

func TestGetError(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"server error"}`))
	})

	_, err := c.GetMonitors()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected 500 in error, got %q", err.Error())
	}
}
