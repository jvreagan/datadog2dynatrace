package datadog

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// FlexTime.UnmarshalJSON
// ---------------------------------------------------------------------------

func TestFlexTimeRFC3339(t *testing.T) {
	var ft FlexTime
	input := `"2024-03-15T10:30:00Z"`
	if err := json.Unmarshal([]byte(input), &ft); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ft.Year() != 2024 || ft.Month() != time.March || ft.Day() != 15 {
		t.Errorf("unexpected time: %v", ft.Time)
	}
}

func TestFlexTimeAlternateFormat(t *testing.T) {
	// "2006-01-02T15:04:05.000000-07:00" format (DataDog API response format)
	var ft FlexTime
	input := `"2024-09-01T00:00:00.000000+00:00"`
	if err := json.Unmarshal([]byte(input), &ft); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ft.Year() != 2024 || ft.Month() != time.September || ft.Day() != 1 {
		t.Errorf("unexpected time: %v", ft.Time)
	}
}

func TestFlexTimeInvalidString(t *testing.T) {
	var ft FlexTime
	// A string that is neither RFC3339Nano nor the alternate format
	input := `"not-a-date"`
	// Should return an error (both formats fail)
	if err := json.Unmarshal([]byte(input), &ft); err == nil {
		t.Fatal("expected error for unparseable date string, got nil")
	}
}

func TestFlexTimeUnixSeconds(t *testing.T) {
	var ft FlexTime
	// Unix timestamp in seconds (≤ 1e12)
	input := `1700000000`
	if err := json.Unmarshal([]byte(input), &ft); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Unix(1700000000, 0)
	if !ft.Time.Equal(want) {
		t.Errorf("got %v, want %v", ft.Time, want)
	}
}

func TestFlexTimeUnixMilliseconds(t *testing.T) {
	var ft FlexTime
	// Unix timestamp in milliseconds (> 1e12)
	ms := int64(1700000000000)
	input := `1700000000000`
	if err := json.Unmarshal([]byte(input), &ft); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Unix(0, ms*int64(time.Millisecond))
	if !ft.Time.Equal(want) {
		t.Errorf("got %v, want %v", ft.Time, want)
	}
}

// ---------------------------------------------------------------------------
// GetDashboards — error fetching individual dashboard
// ---------------------------------------------------------------------------

func TestGetDashboardsDetailError(t *testing.T) {
	callCount := 0
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.URL.Path == "/api/v1/dashboard" {
			// List returns one dashboard
			w.Write([]byte(`{"dashboards":[{"id":"d1","title":"Test"}]}`))
			return
		}
		// Fetching the individual dashboard fails
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	})

	_, err := c.GetDashboards()
	if err == nil {
		t.Fatal("expected error when fetching dashboard detail fails")
	}
}

func TestGetDashboardsSuccess(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/dashboard" {
			w.Write([]byte(`{"dashboards":[{"id":"d1","title":"Dash"}]}`))
			return
		}
		if r.URL.Path == "/api/v1/dashboard/d1" {
			w.Write([]byte(`{"id":"d1","title":"Full Dash","widgets":[]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	dashboards, err := c.GetDashboards()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dashboards) != 1 {
		t.Fatalf("expected 1 dashboard, got %d", len(dashboards))
	}
	if dashboards[0].Title != "Full Dash" {
		t.Errorf("unexpected title: %q", dashboards[0].Title)
	}
}

// ---------------------------------------------------------------------------
// Validate — network error path
// ---------------------------------------------------------------------------

func TestValidateNetworkError(t *testing.T) {
	// Point at a port that refuses connections to trigger a network error.
	c := newClientWithConfig("test-api-key", "test-app-key", "http://127.0.0.1:1", fastConfig())
	err := c.Validate()
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}

// ---------------------------------------------------------------------------
// get — network error path
// ---------------------------------------------------------------------------

func TestGetNetworkError(t *testing.T) {
	c := newClientWithConfig("test-api-key", "test-app-key", "http://127.0.0.1:1", fastConfig())
	_, err := c.GetMonitors()
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}
