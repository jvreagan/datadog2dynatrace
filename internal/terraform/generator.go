package terraform

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

// Generator creates Terraform HCL files from Dynatrace resources.
type Generator struct {
	outputDir string
}

// NewGenerator creates a new Terraform generator.
func NewGenerator(outputDir string) *Generator {
	return &Generator{outputDir: outputDir}
}

// GenerateAll generates Terraform configs for all converted resources.
func (g *Generator) GenerateAll(result *dynatrace.ConversionResult) error {
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Write provider config
	if err := g.writeFile("provider.tf", GenerateProvider()); err != nil {
		return err
	}

	if len(result.Dashboards) > 0 {
		if err := g.writeFile("dashboards.tf", GenerateDashboards(result.Dashboards)); err != nil {
			return err
		}
	}

	if len(result.MetricEvents) > 0 {
		if err := g.writeFile("metric_events.tf", GenerateMetricEvents(result.MetricEvents)); err != nil {
			return err
		}
	}

	if len(result.SLOs) > 0 {
		if err := g.writeFile("slos.tf", GenerateSLOs(result.SLOs)); err != nil {
			return err
		}
	}

	if len(result.Synthetics) > 0 {
		if err := g.writeFile("synthetics.tf", GenerateSynthetics(result.Synthetics)); err != nil {
			return err
		}
	}

	if len(result.LogRules) > 0 {
		if err := g.writeFile("log_processing.tf", GenerateLogProcessing(result.LogRules)); err != nil {
			return err
		}
	}

	if len(result.Maintenance) > 0 {
		if err := g.writeFile("maintenance.tf", GenerateMaintenance(result.Maintenance)); err != nil {
			return err
		}
	}

	if len(result.Notifications) > 0 {
		if err := g.writeFile("notifications.tf", GenerateNotifications(result.Notifications)); err != nil {
			return err
		}
	}

	if len(result.Notebooks) > 0 {
		if err := g.writeFile("notebooks.tf", GenerateNotebooks(result.Notebooks)); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) writeFile(name, content string) error {
	path := filepath.Join(g.outputDir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", name, err)
	}
	return nil
}
