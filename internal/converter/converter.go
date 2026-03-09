package converter

import (
	"fmt"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/logging"
)

// Options configures the converter behavior.
type Options struct {
	EnableGrail bool
}

// Converter orchestrates the conversion of DataDog resources to Dynatrace equivalents.
type Converter struct {
	opts Options
}

// New creates a new Converter with the given options.
func New(opts Options) *Converter {
	return &Converter{opts: opts}
}

// ConvertAll converts all extracted DataDog resources to Dynatrace resources.
// Returns the conversion result and any errors encountered.
func (c *Converter) ConvertAll(ext *datadog.ExtractionResult) (*dynatrace.ConversionResult, []error) {
	result := &dynatrace.ConversionResult{}
	var errs []error

	logging.Info("converting %d dashboards", len(ext.Dashboards))
	for _, d := range ext.Dashboards {
		logging.Debug("converting dashboard %q", d.Title)
		dt, err := ConvertDashboard(&d, c.opts.EnableGrail)
		if err != nil {
			logging.Warn("dashboard %q: %v", d.Title, err)
			errs = append(errs, fmt.Errorf("dashboard %q: %w", d.Title, err))
			continue
		}
		result.Dashboards = append(result.Dashboards, *dt)
	}

	logging.Info("converting %d monitors", len(ext.Monitors))
	for _, m := range ext.Monitors {
		logging.Debug("converting monitor %q", m.Name)
		me, err := ConvertMonitor(&m)
		if err != nil {
			logging.Warn("monitor %q: %v", m.Name, err)
			errs = append(errs, fmt.Errorf("monitor %q: %w", m.Name, err))
			continue
		}
		result.MetricEvents = append(result.MetricEvents, *me)
	}

	logging.Info("converting %d SLOs", len(ext.SLOs))
	for _, s := range ext.SLOs {
		logging.Debug("converting SLO %q", s.Name)
		dt, err := ConvertSLO(&s)
		if err != nil {
			logging.Warn("SLO %q: %v", s.Name, err)
			errs = append(errs, fmt.Errorf("SLO %q: %w", s.Name, err))
			continue
		}
		result.SLOs = append(result.SLOs, *dt)
	}

	logging.Info("converting %d synthetics", len(ext.Synthetics))
	for _, s := range ext.Synthetics {
		logging.Debug("converting synthetic %q", s.Name)
		sm, err := ConvertSynthetic(&s)
		if err != nil {
			logging.Warn("synthetic %q: %v", s.Name, err)
			errs = append(errs, fmt.Errorf("synthetic %q: %w", s.Name, err))
			continue
		}
		result.Synthetics = append(result.Synthetics, *sm)
	}

	logging.Info("converting %d log pipelines", len(ext.LogPipelines))
	for _, l := range ext.LogPipelines {
		logging.Debug("converting log pipeline %q", l.Name)
		rule, err := ConvertLogPipeline(&l)
		if err != nil {
			logging.Warn("log pipeline %q: %v", l.Name, err)
			errs = append(errs, fmt.Errorf("log pipeline %q: %w", l.Name, err))
			continue
		}
		result.LogRules = append(result.LogRules, *rule)
	}

	logging.Info("converting %d metrics", len(ext.Metrics))
	for _, m := range ext.Metrics {
		logging.Debug("converting metric %q", m.Metric)
		md, err := ConvertMetricMetadata(&m)
		if err != nil {
			logging.Warn("metric %q: %v", m.Metric, err)
			errs = append(errs, fmt.Errorf("metric %q: %w", m.Metric, err))
			continue
		}
		result.Metrics = append(result.Metrics, *md)
	}

	logging.Info("converting %d downtimes", len(ext.Downtimes))
	for _, d := range ext.Downtimes {
		logging.Debug("converting downtime %d", d.ID)
		mw, err := ConvertDowntime(&d)
		if err != nil {
			logging.Warn("downtime %d: %v", d.ID, err)
			errs = append(errs, fmt.Errorf("downtime %d: %w", d.ID, err))
			continue
		}
		result.Maintenance = append(result.Maintenance, *mw)
	}

	logging.Info("converting %d notifications", len(ext.Notifications))
	for _, n := range ext.Notifications {
		logging.Debug("converting notification %q", n.Name)
		ni, err := ConvertNotification(&n)
		if err != nil {
			logging.Warn("notification %q: %v", n.Name, err)
			errs = append(errs, fmt.Errorf("notification %q: %w", n.Name, err))
			continue
		}
		result.Notifications = append(result.Notifications, *ni)
	}

	logging.Info("converting %d notebooks", len(ext.Notebooks))
	for _, n := range ext.Notebooks {
		logging.Debug("converting notebook %q", n.Name)
		nb, err := ConvertNotebook(&n)
		if err != nil {
			logging.Warn("notebook %q: %v", n.Name, err)
			errs = append(errs, fmt.Errorf("notebook %q: %w", n.Name, err))
			continue
		}
		result.Notebooks = append(result.Notebooks, *nb)
	}

	return result, errs
}
