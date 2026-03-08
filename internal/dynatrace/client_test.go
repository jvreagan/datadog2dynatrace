package dynatrace

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/ratelimit"
)

// fastConfig returns a rate limiter config suitable for tests.
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
	c := newClientWithConfig(srv.URL, "test-token-123", fastConfig())
	return c, srv
}

func TestCreateDashboard(t *testing.T) {
	var gotMethod, gotPath, gotAuth string
	var gotBody []byte

	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"dash-1"}`))
	})

	d := &Dashboard{DashboardMetadata: DashboardMetadata{Name: "Test Dashboard"}}
	err := c.CreateDashboard(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "POST" {
		t.Errorf("expected POST, got %s", gotMethod)
	}
	if gotPath != "/api/config/v1/dashboards" {
		t.Errorf("expected /api/config/v1/dashboards, got %s", gotPath)
	}
	if gotAuth != "Api-Token test-token-123" {
		t.Errorf("expected Api-Token auth header, got %s", gotAuth)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(gotBody, &body); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	meta := body["dashboardMetadata"].(map[string]interface{})
	if meta["name"] != "Test Dashboard" {
		t.Errorf("expected dashboard name 'Test Dashboard', got %v", meta["name"])
	}
}

func TestCreateMetricEvent(t *testing.T) {
	var gotPath string

	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusCreated)
	})

	err := c.CreateMetricEvent(&MetricEvent{Summary: "High CPU"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/config/v1/anomalyDetection/metricEvents" {
		t.Errorf("expected /api/config/v1/anomalyDetection/metricEvents, got %s", gotPath)
	}
}

func TestCreateSLO(t *testing.T) {
	var gotPath string

	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusCreated)
	})

	err := c.CreateSLO(&SLO{Name: "Availability"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v2/slo" {
		t.Errorf("expected /api/v2/slo, got %s", gotPath)
	}
}

func TestCreateSyntheticMonitor(t *testing.T) {
	var gotPath string

	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusCreated)
	})

	err := c.CreateSyntheticMonitor(&SyntheticMonitor{Name: "Health Check"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v1/synthetic/monitors" {
		t.Errorf("expected /api/v1/synthetic/monitors, got %s", gotPath)
	}
}

func TestCreateMaintenanceWindow(t *testing.T) {
	var gotPath string

	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusCreated)
	})

	err := c.CreateMaintenanceWindow(&MaintenanceWindow{Name: "Deploy Window"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/config/v1/maintenanceWindows" {
		t.Errorf("expected /api/config/v1/maintenanceWindows, got %s", gotPath)
	}
}

func TestCreateNotification(t *testing.T) {
	var gotPath string

	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusCreated)
	})

	err := c.CreateNotification(&NotificationIntegration{Name: "Slack Alert"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/config/v1/notifications" {
		t.Errorf("expected /api/config/v1/notifications, got %s", gotPath)
	}
}

func TestCreateLogProcessingRule(t *testing.T) {
	var gotPath string

	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusCreated)
	})

	err := c.CreateLogProcessingRule(&LogProcessingRule{Name: "Parse JSON"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v2/logs/processing/rules" {
		t.Errorf("expected /api/v2/logs/processing/rules, got %s", gotPath)
	}
}

func TestCreateMetricDescriptor(t *testing.T) {
	var gotMethod, gotPath string

	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	err := c.CreateMetricDescriptor(&MetricDescriptor{MetricID: "custom.my_metric"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "PUT" {
		t.Errorf("expected PUT, got %s", gotMethod)
	}
	if gotPath != "/api/v2/metrics/custom.my_metric" {
		t.Errorf("expected /api/v2/metrics/custom.my_metric, got %s", gotPath)
	}
}

func TestCreateNotebook(t *testing.T) {
	var gotPath string

	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusCreated)
	})

	err := c.CreateNotebook(&DynatraceNotebook{Name: "Migration Notes"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v2/notebooks" {
		t.Errorf("expected /api/v2/notebooks, got %s", gotPath)
	}
}

func TestCreateDashboardError(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"message":"invalid dashboard"}}`))
	})

	err := c.CreateDashboard(&Dashboard{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("expected error to contain status code 400, got %q", err.Error())
	}
}

func TestPushAllSuccess(t *testing.T) {
	var mu sync.Mutex
	paths := map[string]bool{}

	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths[r.Method+" "+r.URL.Path] = true
		mu.Unlock()
		w.WriteHeader(http.StatusCreated)
	})

	result := &ConversionResult{
		Dashboards:    []Dashboard{{DashboardMetadata: DashboardMetadata{Name: "d1"}}},
		MetricEvents:  []MetricEvent{{Summary: "me1"}},
		SLOs:          []SLO{{Name: "s1"}},
		Synthetics:    []SyntheticMonitor{{Name: "sm1"}},
		Maintenance:   []MaintenanceWindow{{Name: "mw1"}},
		Notifications: []NotificationIntegration{{Name: "n1"}},
		LogRules:      []LogProcessingRule{{Name: "lr1"}},
		Metrics:       []MetricDescriptor{{MetricID: "m1"}},
		Notebooks:     []DynatraceNotebook{{Name: "nb1"}},
	}

	errs := c.PushAll(result)
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
	}

	expected := []string{
		"POST /api/config/v1/dashboards",
		"POST /api/config/v1/anomalyDetection/metricEvents",
		"POST /api/v2/slo",
		"POST /api/v1/synthetic/monitors",
		"POST /api/config/v1/maintenanceWindows",
		"POST /api/config/v1/notifications",
		"POST /api/v2/logs/processing/rules",
		"PUT /api/v2/metrics/m1",
		"POST /api/v2/notebooks",
	}
	for _, ep := range expected {
		if !paths[ep] {
			t.Errorf("expected endpoint %q to be hit", ep)
		}
	}
	if len(paths) != 9 {
		t.Errorf("expected exactly 9 endpoints hit, got %d", len(paths))
	}
}

func TestPushAllPartialFailure(t *testing.T) {
	callCount := 0

	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.URL.Path == "/api/config/v1/dashboards" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("server error"))
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	result := &ConversionResult{
		Dashboards: []Dashboard{{DashboardMetadata: DashboardMetadata{Name: "fail-dash"}}},
		SLOs:       []SLO{{Name: "ok-slo"}},
	}

	errs := c.PushAll(result)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0].Error(), "fail-dash") {
		t.Errorf("expected error to mention 'fail-dash', got %q", errs[0].Error())
	}
	// Verify it continued past the failure to hit the SLO endpoint
	if callCount < 2 {
		t.Errorf("expected at least 2 HTTP calls (continues on error), got %d", callCount)
	}
}

func TestPushAllEmpty(t *testing.T) {
	callCount := 0

	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})

	errs := c.PushAll(&ConversionResult{})
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got %d", len(errs))
	}
	if callCount != 0 {
		t.Errorf("expected 0 HTTP requests for empty result, got %d", callCount)
	}
}

func TestValidateSuccess(t *testing.T) {
	var gotMethod, gotPath string

	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"time":1234567890}`))
	})

	err := c.Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "GET" {
		t.Errorf("expected GET, got %s", gotMethod)
	}
	if gotPath != "/api/v1/time" {
		t.Errorf("expected /api/v1/time, got %s", gotPath)
	}
}

func TestListDashboardNames(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/config/v1/dashboards" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"dashboards":[{"id":"d1","name":"Dashboard A"},{"id":"d2","name":"Dashboard B"}]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	names, err := c.ListDashboardNames()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}
	if names[0] != "Dashboard A" || names[1] != "Dashboard B" {
		t.Errorf("unexpected names: %v", names)
	}
}

func TestPushAllSkipExisting(t *testing.T) {
	var mu sync.Mutex
	createdPaths := []string{}

	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		// List endpoints return existing resources
		if r.Method == "GET" {
			switch r.URL.Path {
			case "/api/config/v1/dashboards":
				w.Write([]byte(`{"dashboards":[{"name":"Existing Dashboard"}]}`))
			case "/api/v2/slo":
				w.Write([]byte(`{"slo":[{"name":"Existing SLO"}]}`))
			default:
				w.Write([]byte(`{"values":[],"monitors":[],"notebooks":[]}`))
			}
			return
		}
		// Track POST/PUT calls
		mu.Lock()
		createdPaths = append(createdPaths, r.Method+" "+r.URL.Path)
		mu.Unlock()
		w.WriteHeader(http.StatusCreated)
	})

	result := &ConversionResult{
		Dashboards: []Dashboard{
			{DashboardMetadata: DashboardMetadata{Name: "Existing Dashboard"}},
			{DashboardMetadata: DashboardMetadata{Name: "New Dashboard"}},
		},
		SLOs: []SLO{
			{Name: "Existing SLO"},
			{Name: "New SLO"},
		},
	}

	errs := c.PushAllWithOptions(result, PushOptions{SkipExisting: true})
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
	}

	// Only "New Dashboard" and "New SLO" should be created
	mu.Lock()
	defer mu.Unlock()
	if len(createdPaths) != 2 {
		t.Fatalf("expected 2 creates (skipping existing), got %d: %v", len(createdPaths), createdPaths)
	}
}

func TestPushAllNoSkipExisting(t *testing.T) {
	var mu sync.Mutex
	createCount := 0

	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" || r.Method == "PUT" {
			mu.Lock()
			createCount++
			mu.Unlock()
		}
		w.WriteHeader(http.StatusCreated)
	})

	result := &ConversionResult{
		Dashboards: []Dashboard{
			{DashboardMetadata: DashboardMetadata{Name: "Dashboard"}},
		},
	}

	// Without SkipExisting, no GET calls should be made
	errs := c.PushAllWithOptions(result, PushOptions{SkipExisting: false})
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got %d", len(errs))
	}
	mu.Lock()
	defer mu.Unlock()
	if createCount != 1 {
		t.Errorf("expected 1 create, got %d", createCount)
	}
}

func TestValidateFailure(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"message":"invalid token"}}`))
	})

	err := c.Validate()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected error to contain 401, got %q", err.Error())
	}
}
