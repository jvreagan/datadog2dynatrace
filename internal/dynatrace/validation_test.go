package dynatrace

import (
	"net/http"
	"strings"
	"sync"
	"testing"
)

func TestValidateMetricSelectorValid(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/v2/metrics/query") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":[]}`))
	})

	err := c.ValidateMetricSelector("builtin:host.cpu.usage")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestValidateMetricSelectorInvalid(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"message":"invalid metric selector"}}`))
	})

	err := c.ValidateMetricSelector("not.a.real.metric")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("expected error to contain 400, got %q", err.Error())
	}
}

func TestValidateMetricSelectorServerError(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`internal server error`))
	})

	err := c.ValidateMetricSelector("builtin:host.cpu.usage")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected error to contain 500, got %q", err.Error())
	}
}

func TestValidateAllDeduplicates(t *testing.T) {
	var mu sync.Mutex
	callCount := 0

	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":[]}`))
	})

	result := &ConversionResult{
		MetricEvents: []MetricEvent{
			{Summary: "High CPU on hosts", MetricSelector: "builtin:host.cpu.usage"},
			{Summary: "CPU alert", MetricSelector: "builtin:host.cpu.usage"},
		},
	}

	vr := c.ValidateAll(result)

	mu.Lock()
	defer mu.Unlock()

	if callCount != 1 {
		t.Errorf("expected 1 API call (deduplicated), got %d", callCount)
	}
	if len(vr.Selectors) != 1 {
		t.Errorf("expected 1 selector result, got %d", len(vr.Selectors))
	}
	if len(vr.Selectors[0].Sources) != 2 {
		t.Errorf("expected 2 sources for deduplicated selector, got %d", len(vr.Selectors[0].Sources))
	}
}

func TestValidateAllSkipsPlaceholders(t *testing.T) {
	callCount := 0

	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":[]}`))
	})

	result := &ConversionResult{
		MetricEvents: []MetricEvent{
			{
				Summary:        "Error Log Volume",
				MetricSelector: "builtin:host.availability",
				Description:    "Migration Note: This alert uses DQL. fetch logs | filter ...",
			},
		},
	}

	vr := c.ValidateAll(result)

	if callCount != 0 {
		t.Errorf("expected 0 API calls for placeholder, got %d", callCount)
	}
	if len(vr.Selectors) != 1 {
		t.Fatalf("expected 1 selector result, got %d", len(vr.Selectors))
	}
	if !vr.Selectors[0].Skipped {
		t.Error("expected placeholder selector to be skipped")
	}
	if vr.Summary.Skipped != 1 {
		t.Errorf("expected summary skipped=1, got %d", vr.Summary.Skipped)
	}
}

func TestValidateAllCollectsFromDashboards(t *testing.T) {
	var mu sync.Mutex
	queriedSelectors := []string{}

	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		queriedSelectors = append(queriedSelectors, r.URL.Query().Get("metricSelector"))
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":[]}`))
	})

	result := &ConversionResult{
		Dashboards: []Dashboard{
			{
				DashboardMetadata: DashboardMetadata{Name: "Infra"},
				Tiles: []Tile{
					{
						Name: "CPU Usage",
						Queries: []DashboardQuery{
							{MetricSelector: "builtin:host.cpu.usage"},
						},
					},
					{
						Name: "Memory Usage",
						Queries: []DashboardQuery{
							{MetricSelector: "builtin:host.mem.usage"},
						},
					},
				},
			},
		},
	}

	vr := c.ValidateAll(result)

	if len(vr.Selectors) != 2 {
		t.Fatalf("expected 2 selectors from dashboard tiles, got %d", len(vr.Selectors))
	}
	if vr.Summary.Valid != 2 {
		t.Errorf("expected 2 valid, got %d", vr.Summary.Valid)
	}
}

func TestValidateAllMixedResults(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		selector := r.URL.Query().Get("metricSelector")
		if strings.Contains(selector, "bad") {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":{"message":"unknown metric"}}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":[]}`))
	})

	result := &ConversionResult{
		MetricEvents: []MetricEvent{
			{Summary: "Good metric", MetricSelector: "builtin:host.cpu.usage"},
			{Summary: "Bad metric", MetricSelector: "bad.custom.metric"},
			{
				Summary:        "Placeholder",
				MetricSelector: "builtin:host.availability",
				Description:    "Migration Note: composite monitor",
			},
		},
	}

	vr := c.ValidateAll(result)

	if vr.Summary.Total != 3 {
		t.Errorf("expected total=3, got %d", vr.Summary.Total)
	}
	if vr.Summary.Valid != 1 {
		t.Errorf("expected valid=1, got %d", vr.Summary.Valid)
	}
	if vr.Summary.Invalid != 1 {
		t.Errorf("expected invalid=1, got %d", vr.Summary.Invalid)
	}
	if vr.Summary.Skipped != 1 {
		t.Errorf("expected skipped=1, got %d", vr.Summary.Skipped)
	}
}
