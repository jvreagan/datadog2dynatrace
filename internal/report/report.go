package report

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

// Report generates a Markdown migration report.
type Report struct {
	source    string
	inputDir  string
	target    string
	outputDir string
	dryRun    bool
	sections  []section
}

type section struct {
	title   string
	content string
}

// New creates a new Report.
func New() *Report {
	return &Report{}
}

func (r *Report) SetSource(source, inputDir string) {
	r.source = source
	r.inputDir = inputDir
}

func (r *Report) SetTarget(target, outputDir string) {
	r.target = target
	r.outputDir = outputDir
}

func (r *Report) SetDryRun(dryRun bool) {
	r.dryRun = dryRun
}

// AddExtractionSummary adds a summary of extracted DataDog resources.
func (r *Report) AddExtractionSummary(ext *datadog.ExtractionResult) {
	var sb strings.Builder
	sb.WriteString("| Resource Type | Count |\n")
	sb.WriteString("|---|---|\n")
	sb.WriteString(fmt.Sprintf("| Dashboards | %d |\n", len(ext.Dashboards)))
	sb.WriteString(fmt.Sprintf("| Monitors | %d |\n", len(ext.Monitors)))
	sb.WriteString(fmt.Sprintf("| SLOs | %d |\n", len(ext.SLOs)))
	sb.WriteString(fmt.Sprintf("| Synthetic Tests | %d |\n", len(ext.Synthetics)))
	sb.WriteString(fmt.Sprintf("| Log Pipelines | %d |\n", len(ext.LogPipelines)))
	sb.WriteString(fmt.Sprintf("| Metric Metadata | %d |\n", len(ext.Metrics)))
	sb.WriteString(fmt.Sprintf("| Downtimes | %d |\n", len(ext.Downtimes)))
	sb.WriteString(fmt.Sprintf("| Notification Channels | %d |\n", len(ext.Notifications)))
	sb.WriteString(fmt.Sprintf("| Notebooks | %d |\n", len(ext.Notebooks)))

	r.sections = append(r.sections, section{
		title:   "Extracted Resources (DataDog)",
		content: sb.String(),
	})
}

// AddConversionSummary adds a summary of converted Dynatrace resources.
func (r *Report) AddConversionSummary(result *dynatrace.ConversionResult) {
	var sb strings.Builder
	sb.WriteString("| Resource Type | Count |\n")
	sb.WriteString("|---|---|\n")
	sb.WriteString(fmt.Sprintf("| Dashboards | %d |\n", len(result.Dashboards)))
	sb.WriteString(fmt.Sprintf("| Metric Events | %d |\n", len(result.MetricEvents)))
	sb.WriteString(fmt.Sprintf("| SLOs | %d |\n", len(result.SLOs)))
	sb.WriteString(fmt.Sprintf("| Synthetic Monitors | %d |\n", len(result.Synthetics)))
	sb.WriteString(fmt.Sprintf("| Log Processing Rules | %d |\n", len(result.LogRules)))
	sb.WriteString(fmt.Sprintf("| Metric Descriptors | %d |\n", len(result.Metrics)))
	sb.WriteString(fmt.Sprintf("| Maintenance Windows | %d |\n", len(result.Maintenance)))
	sb.WriteString(fmt.Sprintf("| Notifications | %d |\n", len(result.Notifications)))
	sb.WriteString(fmt.Sprintf("| Notebooks | %d |\n", len(result.Notebooks)))

	r.sections = append(r.sections, section{
		title:   "Converted Resources (Dynatrace)",
		content: sb.String(),
	})
}

// AddConversionErrors adds conversion errors to the report.
func (r *Report) AddConversionErrors(errs []error) {
	if len(errs) == 0 {
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**%d conversion errors occurred:**\n\n", len(errs)))
	for _, err := range errs {
		sb.WriteString(fmt.Sprintf("- %s\n", err))
	}

	r.sections = append(r.sections, section{
		title:   "Conversion Errors",
		content: sb.String(),
	})
}

// AddPushErrors adds push errors to the report.
func (r *Report) AddPushErrors(errs []error) {
	if len(errs) == 0 {
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**%d push errors occurred:**\n\n", len(errs)))
	for _, err := range errs {
		sb.WriteString(fmt.Sprintf("- %s\n", err))
	}

	r.sections = append(r.sections, section{
		title:   "Push Errors",
		content: sb.String(),
	})
}

// WriteToFile writes the report to a Markdown file.
func (r *Report) WriteToFile(path string) error {
	var sb strings.Builder

	sb.WriteString("# DataDog to Dynatrace Migration Report\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format("2006-01-02 15:04:05 MST")))

	// Configuration
	sb.WriteString("## Configuration\n\n")
	sb.WriteString(fmt.Sprintf("- **Source:** %s", r.source))
	if r.inputDir != "" {
		sb.WriteString(fmt.Sprintf(" (`%s`)", r.inputDir))
	}
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("- **Target:** %s", r.target))
	if r.outputDir != "" {
		sb.WriteString(fmt.Sprintf(" (`%s`)", r.outputDir))
	}
	sb.WriteString("\n")
	if r.dryRun {
		sb.WriteString("- **Mode:** Dry Run (no changes applied)\n")
	}
	sb.WriteString("\n")

	// Sections
	for _, s := range r.sections {
		sb.WriteString(fmt.Sprintf("## %s\n\n", s.title))
		sb.WriteString(s.content)
		sb.WriteString("\n")
	}

	// Notes
	sb.WriteString("## Notes\n\n")
	sb.WriteString("- Some resource conversions may require manual adjustments in Dynatrace.\n")
	sb.WriteString("- DataDog query language and Dynatrace metric selectors/DQL are not 1:1 equivalent.\n")
	sb.WriteString("- Review converted dashboards and alerts for accuracy before relying on them.\n")
	sb.WriteString("- Synthetic test locations have been mapped to the nearest Dynatrace equivalent.\n")

	return os.WriteFile(path, []byte(sb.String()), 0644)
}
