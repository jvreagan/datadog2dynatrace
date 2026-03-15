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

func TestNewClient(t *testing.T) {
	c := NewClient("https://abc123.dynatrace.com/", "dt0c01.token")
	if c.envURL != "https://abc123.dynatrace.com" {
		t.Errorf("expected trailing slash trimmed, got %q", c.envURL)
	}
	if c.apiToken != "dt0c01.token" {
		t.Errorf("unexpected apiToken: %q", c.apiToken)
	}
	if c.httpClient == nil {
		t.Error("expected httpClient to be set")
	}
	if c.limiter == nil {
		t.Error("expected limiter to be set")
	}
}

func TestNewTestClient(t *testing.T) {
	c := NewTestClient("https://test.dynatrace.com", "tok")
	if c.envURL != "https://test.dynatrace.com" {
		t.Errorf("unexpected envURL: %q", c.envURL)
	}
	if c.limiter == nil {
		t.Error("expected limiter to be set")
	}
}

func TestListSLONames(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"slo":[{"name":"SLO A"},{"name":"SLO B"}]}`))
	})
	names, err := c.ListSLONames()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 2 || names[0] != "SLO A" || names[1] != "SLO B" {
		t.Errorf("unexpected names: %v", names)
	}
}

func TestListSyntheticNames(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"monitors":[{"name":"Mon A"}]}`))
	})
	names, err := c.ListSyntheticNames()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 1 || names[0] != "Mon A" {
		t.Errorf("unexpected names: %v", names)
	}
}

func TestListMaintenanceNames(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"values":[{"name":"MW 1"},{"name":"MW 2"}]}`))
	})
	names, err := c.ListMaintenanceNames()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 2 || names[0] != "MW 1" {
		t.Errorf("unexpected names: %v", names)
	}
}

func TestListNotificationNames(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"values":[{"name":"Slack"},{"name":"Email"}]}`))
	})
	names, err := c.ListNotificationNames()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d", len(names))
	}
}

func TestListMetricEventSummaries(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"values":[{"name":"High CPU"}]}`))
	})
	names, err := c.ListMetricEventSummaries()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 1 || names[0] != "High CPU" {
		t.Errorf("unexpected names: %v", names)
	}
}

func TestListNotebookNames(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"notebooks":[{"name":"NB 1"},{"name":"NB 2"}]}`))
	})
	names, err := c.ListNotebookNames()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 2 || names[0] != "NB 1" {
		t.Errorf("unexpected names: %v", names)
	}
}

func TestListDashboardNamesError(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	})
	_, err := c.ListDashboardNames()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListDashboardNamesBadJSON(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not json`))
	})
	_, err := c.ListDashboardNames()
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
	if !strings.Contains(err.Error(), "parsing") {
		t.Errorf("expected parsing error, got: %v", err)
	}
}

func TestListSLONamesError(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("forbidden"))
	})
	_, err := c.ListSLONames()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListSLONamesBadJSON(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{broken`))
	})
	_, err := c.ListSLONames()
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestListSyntheticNamesError(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	_, err := c.ListSyntheticNames()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListSyntheticNamesBadJSON(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{`))
	})
	_, err := c.ListSyntheticNames()
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestListMaintenanceNamesError(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	_, err := c.ListMaintenanceNames()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListMaintenanceNamesBadJSON(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not-json`))
	})
	_, err := c.ListMaintenanceNames()
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestListNotificationNamesError(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	_, err := c.ListNotificationNames()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListNotificationNamesBadJSON(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[`))
	})
	_, err := c.ListNotificationNames()
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestListMetricEventSummariesError(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	_, err := c.ListMetricEventSummaries()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListMetricEventSummariesBadJSON(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{bad`))
	})
	_, err := c.ListMetricEventSummaries()
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestListNotebookNamesError(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	_, err := c.ListNotebookNames()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListNotebookNamesBadJSON(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{bad`))
	})
	_, err := c.ListNotebookNames()
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestCreateMetricEventError(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	err := c.CreateMetricEvent(&MetricEvent{Summary: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateSLOError(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	err := c.CreateSLO(&SLO{Name: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateSyntheticMonitorError(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	err := c.CreateSyntheticMonitor(&SyntheticMonitor{Name: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateMaintenanceWindowError(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	err := c.CreateMaintenanceWindow(&MaintenanceWindow{Name: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateNotificationError(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	err := c.CreateNotification(&NotificationIntegration{Name: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateLogProcessingRuleError(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	err := c.CreateLogProcessingRule(&LogProcessingRule{Name: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateMetricDescriptorError(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	err := c.CreateMetricDescriptor(&MetricDescriptor{MetricID: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateNotebookError(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	err := c.CreateNotebook(&DynatraceNotebook{Name: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPushAllWithOptionsSkipAllTypes(t *testing.T) {
	var mu sync.Mutex
	createdPaths := []string{}

	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			switch r.URL.Path {
			case "/api/config/v1/dashboards":
				w.Write([]byte(`{"dashboards":[{"name":"D1"}]}`))
			case "/api/v2/slo":
				w.Write([]byte(`{"slo":[{"name":"S1"}]}`))
			case "/api/v1/synthetic/monitors":
				w.Write([]byte(`{"monitors":[{"name":"Syn1"}]}`))
			case "/api/config/v1/maintenanceWindows":
				w.Write([]byte(`{"values":[{"name":"MW1"}]}`))
			case "/api/config/v1/notifications":
				w.Write([]byte(`{"values":[{"name":"N1"}]}`))
			case "/api/config/v1/anomalyDetection/metricEvents":
				w.Write([]byte(`{"values":[{"name":"ME1"}]}`))
			case "/api/v2/notebooks":
				w.Write([]byte(`{"notebooks":[{"name":"NB1"}]}`))
			default:
				w.Write([]byte(`{}`))
			}
			return
		}
		mu.Lock()
		createdPaths = append(createdPaths, r.URL.Path)
		mu.Unlock()
		w.WriteHeader(http.StatusCreated)
	})

	result := &ConversionResult{
		Dashboards:    []Dashboard{{DashboardMetadata: DashboardMetadata{Name: "D1"}}},
		MetricEvents:  []MetricEvent{{Summary: "ME1"}},
		SLOs:          []SLO{{Name: "S1"}},
		Synthetics:    []SyntheticMonitor{{Name: "Syn1"}},
		Maintenance:   []MaintenanceWindow{{Name: "MW1"}},
		Notifications: []NotificationIntegration{{Name: "N1"}},
		LogRules:      []LogProcessingRule{{Name: "NewRule"}},
		Metrics:       []MetricDescriptor{{MetricID: "custom.new"}},
		Notebooks:     []DynatraceNotebook{{Name: "NB1"}},
	}

	errs := c.PushAllWithOptions(result, PushOptions{SkipExisting: true})
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
	}

	mu.Lock()
	defer mu.Unlock()
	// Only LogRules and Metrics should be created (not in skip lists)
	if len(createdPaths) != 2 {
		t.Errorf("expected 2 creates (log rule + metric), got %d: %v", len(createdPaths), createdPaths)
	}
}

func TestGetErrorPath(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	})
	_, err := c.get("/api/v1/nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("expected 404 in error, got: %v", err)
	}
}

func TestPostErrorPath(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	})
	_, err := c.post("/api/test", map[string]string{"key": "val"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("expected 400 in error, got: %v", err)
	}
}

func TestPutErrorPath(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("conflict"))
	})
	_, err := c.put("/api/test", map[string]string{"key": "val"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "409") {
		t.Errorf("expected 409 in error, got: %v", err)
	}
}

func TestPutSuccess(t *testing.T) {
	var gotMethod string
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	})
	data, err := c.put("/api/test", map[string]string{"key": "val"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "PUT" {
		t.Errorf("expected PUT, got %s", gotMethod)
	}
	if !strings.Contains(string(data), "ok") {
		t.Errorf("unexpected response: %s", data)
	}
}

func TestToNameSet(t *testing.T) {
	s := toNameSet([]string{"a", "b", "c"})
	if !s["a"] || !s["b"] || !s["c"] {
		t.Error("expected all names in set")
	}
	if s["d"] {
		t.Error("unexpected name in set")
	}
}

func TestToNameSetEmpty(t *testing.T) {
	s := toNameSet(nil)
	if len(s) != 0 {
		t.Errorf("expected empty set, got %d", len(s))
	}
}

func TestPushAllWithOptionsSkipNotifications(t *testing.T) {
	var mu sync.Mutex
	createdPaths := []string{}

	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			switch r.URL.Path {
			case "/api/config/v1/notifications":
				w.Write([]byte(`{"values":[{"name":"Slack Alerts"}]}`))
			case "/api/v2/notebooks":
				w.Write([]byte(`{"notebooks":[{"name":"NB Existing"}]}`))
			default:
				w.Write([]byte(`{"dashboards":[],"slo":[],"monitors":[],"values":[]}`))
			}
			return
		}
		mu.Lock()
		createdPaths = append(createdPaths, r.URL.Path)
		mu.Unlock()
		w.WriteHeader(http.StatusCreated)
	})

	result := &ConversionResult{
		Notifications: []NotificationIntegration{
			{Name: "Slack Alerts"},  // should be skipped
			{Name: "New PagerDuty"}, // should be created
		},
		Notebooks: []DynatraceNotebook{
			{Name: "NB Existing"}, // should be skipped
			{Name: "NB New"},      // should be created
		},
		LogRules: []LogProcessingRule{
			{Name: "Log Rule"}, // no skip for log rules
		},
		Metrics: []MetricDescriptor{
			{MetricID: "custom.metric"}, // no skip for metrics
		},
	}

	errs := c.PushAllWithOptions(result, PushOptions{SkipExisting: true})
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
	}

	mu.Lock()
	defer mu.Unlock()
	// Should create: New PagerDuty, NB New, Log Rule, custom.metric = 4
	if len(createdPaths) != 4 {
		t.Errorf("expected 4 creates, got %d: %v", len(createdPaths), createdPaths)
	}
}

func TestPushAllMultipleErrorTypes(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	})

	result := &ConversionResult{
		MetricEvents:  []MetricEvent{{Summary: "ME fail"}},
		SLOs:          []SLO{{Name: "SLO fail"}},
		Synthetics:    []SyntheticMonitor{{Name: "Syn fail"}},
		Maintenance:   []MaintenanceWindow{{Name: "MW fail"}},
		Notifications: []NotificationIntegration{{Name: "N fail"}},
		LogRules:      []LogProcessingRule{{Name: "LR fail"}},
		Metrics:       []MetricDescriptor{{MetricID: "M fail"}},
		Notebooks:     []DynatraceNotebook{{Name: "NB fail"}},
	}

	errs := c.PushAll(result)
	if len(errs) != 8 {
		t.Errorf("expected 8 errors, got %d", len(errs))
	}
}

func TestFetchExistingNamesHandlesErrors(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		// All list endpoints fail
		w.WriteHeader(http.StatusInternalServerError)
	})
	names := c.fetchExistingNames()
	// Should return an empty map (no error propagation, just empty)
	if len(names) != 0 {
		t.Errorf("expected empty map on errors, got %d entries", len(names))
	}
}

func TestDerivePlatformURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://abc12345.live.dynatrace.com", "https://abc12345.apps.dynatrace.com"},
		{"https://umd17470.live.dynatrace.com", "https://umd17470.apps.dynatrace.com"},
		{"https://custom.example.com", "https://custom.example.com"},
	}
	for _, tt := range tests {
		got := derivePlatformURL(tt.input)
		if got != tt.expected {
			t.Errorf("derivePlatformURL(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestNewOAuthClient(t *testing.T) {
	c := NewOAuthClient("https://abc12345.live.dynatrace.com/", "client-id", "client-secret")
	if c.envURL != "https://abc12345.live.dynatrace.com" {
		t.Errorf("expected trailing slash trimmed, got %q", c.envURL)
	}
	if !c.isGen3 {
		t.Error("expected isGen3 to be true")
	}
	if c.platformURL != "https://abc12345.apps.dynatrace.com" {
		t.Errorf("unexpected platformURL: %q", c.platformURL)
	}
	if c.auth.authType() != "oauth" {
		t.Errorf("expected oauth auth type, got %q", c.auth.authType())
	}
}

func TestIsGen3(t *testing.T) {
	classic := NewClient("https://test.dynatrace.com", "tok")
	if classic.IsGen3() {
		t.Error("classic client should not be Gen3")
	}

	oauth := NewOAuthClient("https://test.live.dynatrace.com", "id", "secret")
	if !oauth.IsGen3() {
		t.Error("oauth client should be Gen3")
	}
}

// testGen3Client creates a Gen3 test client with an OAuth mock that always succeeds.
func testGen3Client(t *testing.T, envHandler, platformHandler http.HandlerFunc) (*Client, *httptest.Server, *httptest.Server) {
	t.Helper()

	// Token server
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token":"test-oauth-token","expires_in":300,"token_type":"Bearer"}`))
	}))
	t.Cleanup(tokenSrv.Close)

	// Env API server
	envSrv := httptest.NewServer(envHandler)
	t.Cleanup(envSrv.Close)

	// Platform API server
	platformSrv := httptest.NewServer(platformHandler)
	t.Cleanup(platformSrv.Close)

	c := newOAuthClientWithConfig(envSrv.URL, "test-client-id", "test-client-secret", fastConfig())
	c.platformURL = platformSrv.URL
	c.auth.(*oauthAuth).tokenURL = tokenSrv.URL

	return c, envSrv, platformSrv
}

func TestGen3CreateDashboardRoutesToDocumentsAPI(t *testing.T) {
	var gotPlatformPath string
	var gotAuth string

	c, _, _ := testGen3Client(t,
		func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("unexpected request to env server: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		},
		func(w http.ResponseWriter, r *http.Request) {
			gotPlatformPath = r.URL.Path
			gotAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id":"doc-1"}`))
		},
	)

	d := &Dashboard{DashboardMetadata: DashboardMetadata{Name: "Gen3 Dashboard"}}
	err := c.CreateDashboard(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPlatformPath != "/platform/document/v1/documents" {
		t.Errorf("expected /platform/document/v1/documents, got %s", gotPlatformPath)
	}
	if !strings.HasPrefix(gotAuth, "Bearer ") {
		t.Errorf("expected Bearer auth, got %q", gotAuth)
	}
}

func TestGen3CreateMetricEventRoutesToSettings(t *testing.T) {
	var gotPath string

	c, _, _ := testGen3Client(t,
		func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`[{"code":200}]`))
		},
		func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("unexpected request to platform server: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		},
	)

	me := &MetricEvent{
		Summary:        "High CPU",
		MetricSelector: "builtin:host.cpu.usage",
		Enabled:        true,
		EventType:      "CUSTOM_ALERT",
		MonitoringStrategy: MonitoringStrategy{
			Type:      "STATIC_THRESHOLD",
			Threshold: 90,
		},
	}
	err := c.CreateMetricEvent(me)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v2/settings/objects" {
		t.Errorf("expected /api/v2/settings/objects, got %s", gotPath)
	}
}

func TestGen3CreateMaintenanceRoutesToSettings(t *testing.T) {
	var gotPath string

	c, _, _ := testGen3Client(t,
		func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`[{"code":200}]`))
		},
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	)

	mw := &MaintenanceWindow{Name: "Deploy Window", Type: "PLANNED", Suppression: "DETECT_PROBLEMS_AND_ALERT"}
	err := c.CreateMaintenanceWindow(mw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v2/settings/objects" {
		t.Errorf("expected /api/v2/settings/objects, got %s", gotPath)
	}
}

func TestGen3CreateNotificationRoutesToSettings(t *testing.T) {
	var gotPath string

	c, _, _ := testGen3Client(t,
		func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`[{"code":200}]`))
		},
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	)

	n := &NotificationIntegration{Name: "Slack", Type: "SLACK", Active: true, Config: map[string]interface{}{"url": "https://hooks.slack.com/test"}}
	err := c.CreateNotification(n)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v2/settings/objects" {
		t.Errorf("expected /api/v2/settings/objects, got %s", gotPath)
	}
}

func TestGen3ValidateUsesV2Time(t *testing.T) {
	var gotPath string

	c, _, _ := testGen3Client(t,
		func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"time":1234567890}`))
		},
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	)

	err := c.Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v2/time" {
		t.Errorf("expected /api/v2/time for Gen3, got %s", gotPath)
	}
}

func TestGen3ListDashboardNames(t *testing.T) {
	c, _, _ := testGen3Client(t,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"documents":[{"name":"Doc A"},{"name":"Doc B"}]}`))
		},
	)

	names, err := c.ListDashboardNames()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 2 || names[0] != "Doc A" || names[1] != "Doc B" {
		t.Errorf("unexpected names: %v", names)
	}
}

func TestGen3ListMetricEventSummaries(t *testing.T) {
	c, _, _ := testGen3Client(t,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"items":[{"value":{"title":"Alert A"}},{"value":{"title":"Alert B"}}]}`))
		},
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	)

	names, err := c.ListMetricEventSummaries()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 2 || names[0] != "Alert A" {
		t.Errorf("unexpected names: %v", names)
	}
}

func TestMapEventType(t *testing.T) {
	tests := []struct{ in, out string }{
		{"CUSTOM_ALERT", "CUSTOM_ALERT"},
		{"ERROR", "ERROR_EVENT"},
		{"INFO", "CUSTOM_INFO"},
		{"UNKNOWN", "CUSTOM_ALERT"},
	}
	for _, tt := range tests {
		got := mapEventType(tt.in)
		if got != tt.out {
			t.Errorf("mapEventType(%q) = %q, want %q", tt.in, got, tt.out)
		}
	}
}

func TestMapModelType(t *testing.T) {
	tests := []struct{ in, out string }{
		{"STATIC_THRESHOLD", "STATIC"},
		{"AUTO_ADAPTIVE", "AUTO_ADAPTIVE"},
		{"UNKNOWN", "STATIC"},
	}
	for _, tt := range tests {
		got := mapModelType(tt.in)
		if got != tt.out {
			t.Errorf("mapModelType(%q) = %q, want %q", tt.in, got, tt.out)
		}
	}
}
