package converter

import (
	"fmt"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

// Converter orchestrates the conversion of DataDog resources to Dynatrace equivalents.
type Converter struct{}

// New creates a new Converter.
func New() *Converter {
	return &Converter{}
}

// ConvertAll converts all extracted DataDog resources to Dynatrace resources.
// Returns the conversion result and any errors encountered.
func (c *Converter) ConvertAll(ext *datadog.ExtractionResult) (*dynatrace.ConversionResult, []error) {
	result := &dynatrace.ConversionResult{}
	var errs []error

	for _, d := range ext.Dashboards {
		dt, err := ConvertDashboard(&d)
		if err != nil {
			errs = append(errs, fmt.Errorf("dashboard %q: %w", d.Title, err))
			continue
		}
		result.Dashboards = append(result.Dashboards, *dt)
	}

	for _, m := range ext.Monitors {
		me, err := ConvertMonitor(&m)
		if err != nil {
			errs = append(errs, fmt.Errorf("monitor %q: %w", m.Name, err))
			continue
		}
		result.MetricEvents = append(result.MetricEvents, *me)
	}

	for _, s := range ext.SLOs {
		dt, err := ConvertSLO(&s)
		if err != nil {
			errs = append(errs, fmt.Errorf("SLO %q: %w", s.Name, err))
			continue
		}
		result.SLOs = append(result.SLOs, *dt)
	}

	for _, s := range ext.Synthetics {
		sm, err := ConvertSynthetic(&s)
		if err != nil {
			errs = append(errs, fmt.Errorf("synthetic %q: %w", s.Name, err))
			continue
		}
		result.Synthetics = append(result.Synthetics, *sm)
	}

	for _, l := range ext.LogPipelines {
		rule, err := ConvertLogPipeline(&l)
		if err != nil {
			errs = append(errs, fmt.Errorf("log pipeline %q: %w", l.Name, err))
			continue
		}
		result.LogRules = append(result.LogRules, *rule)
	}

	for _, m := range ext.Metrics {
		md, err := ConvertMetricMetadata(&m)
		if err != nil {
			errs = append(errs, fmt.Errorf("metric %q: %w", m.Metric, err))
			continue
		}
		result.Metrics = append(result.Metrics, *md)
	}

	for _, d := range ext.Downtimes {
		mw, err := ConvertDowntime(&d)
		if err != nil {
			errs = append(errs, fmt.Errorf("downtime %d: %w", d.ID, err))
			continue
		}
		result.Maintenance = append(result.Maintenance, *mw)
	}

	for _, n := range ext.Notifications {
		ni, err := ConvertNotification(&n)
		if err != nil {
			errs = append(errs, fmt.Errorf("notification %q: %w", n.Name, err))
			continue
		}
		result.Notifications = append(result.Notifications, *ni)
	}

	for _, n := range ext.Notebooks {
		nb, err := ConvertNotebook(&n)
		if err != nil {
			errs = append(errs, fmt.Errorf("notebook %q: %w", n.Name, err))
			continue
		}
		result.Notebooks = append(result.Notebooks, *nb)
	}

	return result, errs
}
