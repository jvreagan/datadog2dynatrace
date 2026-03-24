package dynatrace

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// noopHandler is a convenience handler that returns 200 with an empty body.
func noopHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// errAuth is a mock authProvider that always returns an error from setAuth.
type errAuth struct{}

func (a *errAuth) setAuth(req *http.Request) error { return fmt.Errorf("mock auth error") }
func (a *errAuth) authType() string                { return "error" }

// ---------------------------------------------------------------------------
// delete / DeleteDashboard / DeleteMetricEvent
// ---------------------------------------------------------------------------

func TestDeleteDashboard(t *testing.T) {
	var gotMethod, gotPath string
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.DeleteDashboard("dash-abc"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "DELETE" {
		t.Errorf("expected DELETE, got %s", gotMethod)
	}
	if gotPath != "/api/config/v1/dashboards/dash-abc" {
		t.Errorf("unexpected path: %s", gotPath)
	}
}

func TestDeleteDashboardNotFound(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	if err := c.DeleteDashboard("missing"); err != nil {
		t.Fatalf("expected nil for 404, got: %v", err)
	}
}

func TestDeleteDashboardError(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	})
	if err := c.DeleteDashboard("dash-err"); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeleteDashboardGen3(t *testing.T) {
	var gotMethod, gotPath string
	c, _, _ := testGen3Client(t,
		noopHandler,
		func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusNoContent)
		},
	)
	if err := c.DeleteDashboard("doc-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "DELETE" {
		t.Errorf("expected DELETE, got %s", gotMethod)
	}
	if !strings.Contains(gotPath, "doc-123") {
		t.Errorf("expected path to contain doc-123, got %s", gotPath)
	}
}

func TestDeleteDashboardGen3Error(t *testing.T) {
	c, _, _ := testGen3Client(t,
		noopHandler,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("forbidden"))
		},
	)
	err := c.DeleteDashboard("doc-forbidden")
	if err == nil {
		t.Fatal("expected error for 403, got nil")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("expected 403 in error, got %q", err.Error())
	}
}

func TestDeleteDashboardGen3NotFound(t *testing.T) {
	c, _, _ := testGen3Client(t,
		noopHandler,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	)
	if err := c.DeleteDashboard("doc-gone"); err != nil {
		t.Fatalf("expected nil for 404, got: %v", err)
	}
}

func TestDeleteMetricEvent(t *testing.T) {
	var gotMethod, gotPath string
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.DeleteMetricEvent("me-001"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "DELETE" {
		t.Errorf("expected DELETE, got %s", gotMethod)
	}
	if gotPath != "/api/config/v1/anomalyDetection/metricEvents/me-001" {
		t.Errorf("unexpected path: %s", gotPath)
	}
}

func TestDeleteMetricEventGen3(t *testing.T) {
	var gotPath string
	c, envSrv, _ := testGen3Client(t,
		func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusNoContent)
		},
		noopHandler,
	)
	_ = envSrv
	if err := c.DeleteMetricEvent("obj-456"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotPath, "obj-456") {
		t.Errorf("expected path to contain obj-456, got %s", gotPath)
	}
}

// ---------------------------------------------------------------------------
// deleteExisting
// ---------------------------------------------------------------------------

func TestPushAllConflictResolverReplace(t *testing.T) {
	var deletedPath string
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			switch r.URL.Path {
			case "/api/config/v1/dashboards":
				w.Write([]byte(`{"dashboards":[{"id":"old-id","name":"Replace Dashboard"}]}`))
			default:
				w.Write([]byte(`{"values":[],"slo":[],"monitors":[],"notebooks":[]}`))
			}
			return
		}
		if r.Method == "DELETE" {
			deletedPath = r.URL.Path
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	result := &ConversionResult{
		Dashboards: []Dashboard{
			{DashboardMetadata: DashboardMetadata{Name: "Replace Dashboard"}},
		},
	}

	errs := c.PushAllWithOptions(result, PushOptions{
		ConflictResolver: func(resourceType, name string) ConflictAction {
			return ConflictReplace
		},
	})
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
	}
	if deletedPath != "/api/config/v1/dashboards/old-id" {
		t.Errorf("expected delete of old-id, got path %q", deletedPath)
	}
}

func TestPushAllConflictResolverSkip(t *testing.T) {
	createCount := 0
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			switch r.URL.Path {
			case "/api/config/v1/dashboards":
				w.Write([]byte(`{"dashboards":[{"id":"skip-id","name":"Skip Dashboard"}]}`))
			default:
				w.Write([]byte(`{"values":[],"slo":[],"monitors":[],"notebooks":[]}`))
			}
			return
		}
		if r.Method == "POST" {
			createCount++
		}
		w.WriteHeader(http.StatusCreated)
	})

	result := &ConversionResult{
		Dashboards: []Dashboard{
			{DashboardMetadata: DashboardMetadata{Name: "Skip Dashboard"}},
		},
	}

	errs := c.PushAllWithOptions(result, PushOptions{
		ConflictResolver: func(resourceType, name string) ConflictAction {
			return ConflictSkip
		},
	})
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
	}
	if createCount != 0 {
		t.Errorf("expected 0 creates when ConflictSkip, got %d", createCount)
	}
}

func TestPushAllConflictResolverReplaceNoID(t *testing.T) {
	// SLO has no ID — delete should be skipped, resource should be skipped too.
	createCount := 0
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			if r.URL.Path == "/api/v2/slo" {
				w.Write([]byte(`{"slo":[{"name":"No-ID SLO"}]}`))
				return
			}
			w.Write([]byte(`{"values":[],"dashboards":[],"monitors":[],"notebooks":[]}`))
			return
		}
		if r.Method == "POST" {
			createCount++
		}
		w.WriteHeader(http.StatusCreated)
	})

	result := &ConversionResult{
		SLOs: []SLO{{Name: "No-ID SLO"}},
	}

	errs := c.PushAllWithOptions(result, PushOptions{
		ConflictResolver: func(resourceType, name string) ConflictAction {
			return ConflictReplace
		},
	})
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
	}
	if createCount != 0 {
		t.Errorf("expected 0 creates when ID unknown for replace, got %d", createCount)
	}
}

func TestDeleteExistingUnknownType(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	err := c.deleteExisting("unknown_type", "some-id")
	if err == nil {
		t.Fatal("expected error for unknown resource type")
	}
	if !strings.Contains(err.Error(), "delete not implemented") {
		t.Errorf("expected 'delete not implemented' in error, got %q", err.Error())
	}
}

// ---------------------------------------------------------------------------
// Gen3 list functions
// ---------------------------------------------------------------------------

func TestListMaintenanceNamesGen3(t *testing.T) {
	c, _, _ := testGen3Client(t,
		func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.RawQuery, "builtin:alerting.maintenance-window") {
				w.Write([]byte(`{"items":[{"value":{"name":"MW Gen3 A"}},{"value":{"name":"MW Gen3 B"}}]}`))
				return
			}
			w.Write([]byte(`{"items":[]}`))
		},
		noopHandler,
	)
	names, err := c.ListMaintenanceNames()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}
	if names[0] != "MW Gen3 A" {
		t.Errorf("unexpected name: %q", names[0])
	}
}

func TestListNotificationNamesGen3(t *testing.T) {
	c, _, _ := testGen3Client(t,
		func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.RawQuery, "builtin:problem.notifications") {
				w.Write([]byte(`{"items":[{"value":{"name":"Notif Gen3 A"}}]}`))
				return
			}
			w.Write([]byte(`{"items":[]}`))
		},
		noopHandler,
	)
	names, err := c.ListNotificationNames()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 1 {
		t.Fatalf("expected 1 name, got %d", len(names))
	}
	if names[0] != "Notif Gen3 A" {
		t.Errorf("unexpected name: %q", names[0])
	}
}

func TestListMaintenanceNamesGen3Error(t *testing.T) {
	c, _, _ := testGen3Client(t,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
		noopHandler,
	)
	_, err := c.ListMaintenanceNames()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListNotificationNamesGen3Error(t *testing.T) {
	c, _, _ := testGen3Client(t,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
		noopHandler,
	)
	_, err := c.ListNotificationNames()
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// metricSelectorToDQL + dqlMetricName
// ---------------------------------------------------------------------------

func TestMetricSelectorToDQL(t *testing.T) {
	tests := []struct {
		selector     string
		wantContains string
	}{
		{
			// No splitBy — simple wrap with builtin → dt. prefix
			selector:     "builtin:host.cpu.user",
			wantContains: "dt.host.cpu.user",
		},
		{
			// splitBy with no dimensions — defaults to avg, no by clause
			selector:     "builtin:host.cpu.user:splitBy()",
			wantContains: "timeseries avg(dt.host.cpu.user)",
		},
		{
			// splitBy with a dimension and explicit aggregation
			selector:     `builtin:host.cpu.user:splitBy("host.name"):max`,
			wantContains: "by:{host.name}",
		},
		{
			// splitBy with multiple dimensions
			selector:     `builtin:host.disk.used:splitBy("host.name","disk.device"):sum`,
			wantContains: "by:{host.name,disk.device}",
		},
		{
			// ext: metric — no splitBy
			selector:     "ext:custom.queue.size",
			wantContains: "ext.custom.queue.size",
		},
		{
			// plain metric (no prefix) — no splitBy
			selector:     "custom.requests.total",
			wantContains: "custom.requests.total",
		},
	}

	for _, tt := range tests {
		t.Run(tt.selector, func(t *testing.T) {
			got := metricSelectorToDQL(tt.selector)
			if !strings.Contains(got, tt.wantContains) {
				t.Errorf("metricSelectorToDQL(%q) = %q; want it to contain %q", tt.selector, got, tt.wantContains)
			}
		})
	}
}

func TestDqlMetricName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"builtin:host.cpu.user", "dt.host.cpu.user"},
		{"ext:my.custom.metric", "ext.my.custom.metric"},
		{"custom.plain.metric", "custom.plain.metric"},
	}
	for _, tt := range tests {
		got := dqlMetricName(tt.input)
		if got != tt.want {
			t.Errorf("dqlMetricName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Validate Gen3 path
// ---------------------------------------------------------------------------

func TestValidateGen3Success(t *testing.T) {
	c, _, _ := testGen3Client(t,
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v2/time" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"timestamp":1234567890}`))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		},
		noopHandler,
	)
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateGen3Failure(t *testing.T) {
	c, _, _ := testGen3Client(t,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"unauthorized"}`))
		},
		noopHandler,
	)
	err := c.Validate()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected 401 in error, got %q", err.Error())
	}
}

// ---------------------------------------------------------------------------
// postPlatform error path (via Gen3 dashboard create)
// ---------------------------------------------------------------------------

func TestCreateDashboardGen3Error(t *testing.T) {
	c, _, _ := testGen3Client(t,
		noopHandler,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte(`{"error":"unprocessable"}`))
		},
	)
	if err := c.CreateDashboard(&Dashboard{DashboardMetadata: DashboardMetadata{Name: "Bad"}}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// ListDashboardsWithIDs Gen3 path
// ---------------------------------------------------------------------------

func TestListDashboardsWithIDsGen3(t *testing.T) {
	c, _, _ := testGen3Client(t,
		noopHandler,
		func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/platform/document/v1/documents") {
				w.Write([]byte(`{"documents":[{"id":"d1","name":"Gen3 Dash A"},{"id":"d2","name":"Gen3 Dash B"}]}`))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		},
	)
	resources, err := c.ListDashboardsWithIDs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}
	if resources[0].Name != "Gen3 Dash A" {
		t.Errorf("unexpected name: %q", resources[0].Name)
	}
}

// ---------------------------------------------------------------------------
// setHeaders / get / post / put / delete / getPlatform auth-error paths
// ---------------------------------------------------------------------------

func TestGetAuthError(t *testing.T) {
	c, _ := testClient(t, noopHandler)
	c.auth = &errAuth{}
	_, err := c.get("/api/v1/time")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPostAuthError(t *testing.T) {
	c, _ := testClient(t, noopHandler)
	c.auth = &errAuth{}
	_, err := c.post("/api/v1/dashboards", struct{}{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPutAuthError(t *testing.T) {
	c, _ := testClient(t, noopHandler)
	c.auth = &errAuth{}
	_, err := c.put("/api/v2/metrics/test", struct{}{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeleteAuthError(t *testing.T) {
	c, _ := testClient(t, noopHandler)
	c.auth = &errAuth{}
	err := c.delete("/api/config/v1/dashboards/x")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPlatformAuthError(t *testing.T) {
	c, srv := testClient(t, noopHandler)
	c.platformURL = srv.URL
	c.auth = &errAuth{}
	_, err := c.getPlatform("/platform/document/v1/documents")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestValidateNonGen3AuthError(t *testing.T) {
	c, _ := testClient(t, noopHandler)
	c.auth = &errAuth{}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestValidateGen3AuthError(t *testing.T) {
	c, _, _ := testGen3Client(t, noopHandler, noopHandler)
	c.auth = &errAuth{}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// postPlatform — success, 4xx error, network error, auth error
// ---------------------------------------------------------------------------

func TestPostPlatformSuccess(t *testing.T) {
	var gotMethod string
	c, srv := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"new"}`))
	})
	c.platformURL = srv.URL

	resp, err := c.postPlatform("/api/v2/test", map[string]string{"key": "val"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Error("expected non-nil response body")
	}
	if gotMethod != "POST" {
		t.Errorf("expected POST, got %s", gotMethod)
	}
}

func TestPostPlatformError(t *testing.T) {
	c, srv := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"bad request"}`))
	})
	c.platformURL = srv.URL

	_, err := c.postPlatform("/api/v2/test", struct{}{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("expected 400 in error, got: %v", err)
	}
}

func TestPostPlatformAuthError(t *testing.T) {
	c, srv := testClient(t, noopHandler)
	c.platformURL = srv.URL
	c.auth = &errAuth{}
	_, err := c.postPlatform("/api/v2/test", struct{}{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPostPlatformNetworkError(t *testing.T) {
	// Create a server, capture the URL, then close it immediately.
	closedSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	closedURL := closedSrv.URL
	closedSrv.Close()

	c := newClientWithConfig("http://localhost", "test-token", fastConfig())
	c.platformURL = closedURL

	_, err := c.postPlatform("/api/v2/test", struct{}{})
	if err == nil {
		t.Fatal("expected network error, got nil")
	}
}

// ---------------------------------------------------------------------------
// postPlatformMultipart + deletePlatform auth-error paths
// ---------------------------------------------------------------------------

func TestPostPlatformMultipartAuthError(t *testing.T) {
	c, _, _ := testGen3Client(t, noopHandler, noopHandler)
	c.auth = &errAuth{}
	// CreateDashboard on Gen3 uses postPlatformMultipart
	err := c.CreateDashboard(&Dashboard{DashboardMetadata: DashboardMetadata{Name: "Err"}})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeletePlatformAuthError(t *testing.T) {
	c, _, _ := testGen3Client(t, noopHandler, noopHandler)
	c.auth = &errAuth{}
	// DeleteDashboard on Gen3 uses deletePlatform
	err := c.DeleteDashboard("some-doc-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// delete — 4xx status (passes through limiter, not retried)
// ---------------------------------------------------------------------------

func TestDeleteDashboardForbidden(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("forbidden"))
	})
	err := c.DeleteDashboard("dash-forbidden")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("expected 403 in error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// PushAllWithOptions — SkipExisting path + notebook skip
// ---------------------------------------------------------------------------

func TestPushAllSkipExistingNotebook(t *testing.T) {
	createCount := 0
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			switch r.URL.Path {
			case "/api/config/v1/dashboards":
				w.Write([]byte(`{"dashboards":[{"id":"d1","name":"Skip Dash"}]}`))
			case "/api/v2/notebooks":
				w.Write([]byte(`{"notebooks":[{"name":"Skip Notebook"}]}`))
			default:
				w.Write([]byte(`{}`))
			}
			return
		}
		if r.Method == "POST" {
			createCount++
		}
		w.WriteHeader(http.StatusCreated)
	})

	result := &ConversionResult{
		Dashboards: []Dashboard{
			{DashboardMetadata: DashboardMetadata{Name: "Skip Dash"}},
		},
		Notebooks: []DynatraceNotebook{{Name: "Skip Notebook"}},
	}

	errs := c.PushAllWithOptions(result, PushOptions{SkipExisting: true})
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
	}
	if createCount != 0 {
		t.Errorf("expected 0 creates when SkipExisting, got %d", createCount)
	}
}

// ---------------------------------------------------------------------------
// deleteExisting — metric_event case
// ---------------------------------------------------------------------------

func TestDeleteExistingMetricEvent(t *testing.T) {
	var gotPath string
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.deleteExisting("metric_event", "me-test-id"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotPath, "me-test-id") {
		t.Errorf("expected path to contain me-test-id, got %s", gotPath)
	}
}

// ---------------------------------------------------------------------------
// OAuth setAuth error path (when getToken fails)
// ---------------------------------------------------------------------------

func TestOAuthSetAuthWhenGetTokenFails(t *testing.T) {
	// Token server returns 401
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid_client"}`))
	}))
	t.Cleanup(tokenSrv.Close)

	a := newOAuthAuth("bad-id", "bad-secret", tokenSrv.Client())
	a.tokenURL = tokenSrv.URL

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	err := a.setAuth(req)
	if err == nil {
		t.Fatal("expected error from setAuth when getToken fails")
	}
	if !strings.Contains(err.Error(), "obtaining OAuth token") {
		t.Errorf("expected 'obtaining OAuth token' in error, got: %v", err)
	}
}

func TestOAuthGetTokenNetworkError(t *testing.T) {
	a := newOAuthAuth("id", "secret", http.DefaultClient)
	a.tokenURL = "http://127.0.0.1:1" // always refuses connections

	_, err := a.getToken()
	if err == nil {
		t.Fatal("expected network error, got nil")
	}
}

// ---------------------------------------------------------------------------
// Gen3 list functions — getPlatform/get error paths
// ---------------------------------------------------------------------------

func TestListDashboardsWithIDsGen3Error(t *testing.T) {
	c, _, _ := testGen3Client(t,
		noopHandler,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error":"forbidden"}`))
		},
	)
	_, err := c.ListDashboardsWithIDs()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListMetricEventsWithIDsGen3Error(t *testing.T) {
	c, _, _ := testGen3Client(t,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		},
		noopHandler,
	)
	_, err := c.ListMetricEventsWithIDs()
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// Gen3 list functions — JSON parse error paths
// ---------------------------------------------------------------------------

func TestListDashboardsWithIDsGen3JSONError(t *testing.T) {
	c, _, _ := testGen3Client(t,
		noopHandler,
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`not valid json`))
		},
	)
	_, err := c.ListDashboardsWithIDs()
	if err == nil {
		t.Fatal("expected JSON parse error")
	}
}

func TestListMaintenanceNamesGen3JSONError(t *testing.T) {
	c, _, _ := testGen3Client(t,
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`not valid json`))
		},
		noopHandler,
	)
	_, err := c.ListMaintenanceNames()
	if err == nil {
		t.Fatal("expected JSON parse error")
	}
}

func TestListNotificationNamesGen3JSONError(t *testing.T) {
	c, _, _ := testGen3Client(t,
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`not valid json`))
		},
		noopHandler,
	)
	_, err := c.ListNotificationNames()
	if err == nil {
		t.Fatal("expected JSON parse error")
	}
}

func TestListMetricEventsWithIDsGen3JSONError(t *testing.T) {
	c, _, _ := testGen3Client(t,
		func(w http.ResponseWriter, r *http.Request) {
			// Return invalid JSON only for metric event list
			if strings.Contains(r.URL.RawQuery, "davis.anomaly-detectors") {
				w.Write([]byte(`not valid json`))
				return
			}
			w.Write([]byte(`{}`))
		},
		noopHandler,
	)
	_, err := c.ListMetricEventsWithIDs()
	if err == nil {
		t.Fatal("expected JSON parse error")
	}
}

// ---------------------------------------------------------------------------
// Gen3 create functions — error paths
// ---------------------------------------------------------------------------

func TestCreateMaintenanceWindowGen3Error(t *testing.T) {
	c, _, _ := testGen3Client(t,
		func(w http.ResponseWriter, r *http.Request) {
			// Return 400 for POST (passes through limiter)
			if r.Method == "POST" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error":"bad request"}`))
				return
			}
			w.WriteHeader(http.StatusNoContent)
		},
		noopHandler,
	)
	err := c.CreateMaintenanceWindow(&MaintenanceWindow{Name: "Test MW"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateNotificationGen3Error(t *testing.T) {
	c, _, _ := testGen3Client(t,
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error":"bad request"}`))
				return
			}
			w.WriteHeader(http.StatusNoContent)
		},
		noopHandler,
	)
	err := c.CreateNotification(&NotificationIntegration{Name: "Test Notif"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateMetricEventGen3Error(t *testing.T) {
	c, _, _ := testGen3Client(t,
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error":"bad request"}`))
				return
			}
			w.WriteHeader(http.StatusNoContent)
		},
		noopHandler,
	)
	err := c.CreateMetricEvent(&MetricEvent{
		Summary:           "Test Event",
		MonitoringStrategy: MonitoringStrategy{AlertCondition: "ABOVE"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// PushAllWithOptions — deleteExisting failure warning (line 492)
// ---------------------------------------------------------------------------

func TestPushAllConflictReplaceDeletionFailed(t *testing.T) {
	// When deletion fails, the code logs a warning but still proceeds to create.
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			if r.URL.Path == "/api/config/v1/dashboards" {
				w.Write([]byte(`{"dashboards":[{"id":"old-id","name":"Fail Delete Dash"}]}`))
				return
			}
			w.Write([]byte(`{}`))
			return
		}
		if r.Method == "DELETE" {
			// 403 passes through limiter and causes delete() to return an error.
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("forbidden"))
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	result := &ConversionResult{
		Dashboards: []Dashboard{
			{DashboardMetadata: DashboardMetadata{Name: "Fail Delete Dash"}},
		},
	}

	// Deletion failure is only logged (not returned as an error); creation still proceeds.
	errs := c.PushAllWithOptions(result, PushOptions{
		ConflictResolver: func(rt, name string) ConflictAction { return ConflictReplace },
	})
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
	}
}

// ---------------------------------------------------------------------------
// getPlatform / postPlatformMultipart / deletePlatform — network error paths
// ---------------------------------------------------------------------------

func TestGetPlatformNetworkError(t *testing.T) {
	// Close the platform server before any platform request is made.
	platSrv := httptest.NewServer(http.HandlerFunc(noopHandler))
	platURL := platSrv.URL
	platSrv.Close()

	c, _, _ := testGen3Client(t, noopHandler, noopHandler)
	c.platformURL = platURL // point to closed server

	_, err := c.getPlatform("/platform/document/v1/documents")
	if err == nil {
		t.Fatal("expected network error, got nil")
	}
}

func TestPostPlatformMultipartNetworkError(t *testing.T) {
	platSrv := httptest.NewServer(http.HandlerFunc(noopHandler))
	platURL := platSrv.URL
	platSrv.Close()

	c, _, _ := testGen3Client(t, noopHandler, noopHandler)
	c.platformURL = platURL

	err := c.CreateDashboard(&Dashboard{DashboardMetadata: DashboardMetadata{Name: "Net Err"}})
	if err == nil {
		t.Fatal("expected network error, got nil")
	}
}

// ---------------------------------------------------------------------------
// ValidateAll — empty MetricSelector paths
// ---------------------------------------------------------------------------

func TestValidateAllEmptySelectors(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("unexpected HTTP request — empty selectors should be skipped")
	})

	result := &ConversionResult{
		MetricEvents: []MetricEvent{
			{Summary: "No selector", MetricSelector: ""},
		},
		Dashboards: []Dashboard{{
			Tiles: []Tile{{
				Name:    "Empty tile",
				Queries: []DashboardQuery{{MetricSelector: ""}},
			}},
		}},
	}

	vr := c.ValidateAll(result)
	if vr.Summary.Total != 0 {
		t.Errorf("expected 0 selectors (all empty), got %d", vr.Summary.Total)
	}
}
