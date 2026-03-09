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
	sb.WriteString("| Resource Type | Count | Names |\n")
	sb.WriteString("|---|---|---|\n")

	dashNames := make([]string, len(ext.Dashboards))
	for i, d := range ext.Dashboards {
		dashNames[i] = d.Title
	}
	sb.WriteString(fmt.Sprintf("| Dashboards | %d | %s |\n", len(ext.Dashboards), joinResourceNames(dashNames)))

	monNames := make([]string, len(ext.Monitors))
	for i, m := range ext.Monitors {
		monNames[i] = m.Name
	}
	sb.WriteString(fmt.Sprintf("| Monitors | %d | %s |\n", len(ext.Monitors), joinResourceNames(monNames)))

	sloNames := make([]string, len(ext.SLOs))
	for i, s := range ext.SLOs {
		sloNames[i] = s.Name
	}
	sb.WriteString(fmt.Sprintf("| SLOs | %d | %s |\n", len(ext.SLOs), joinResourceNames(sloNames)))

	synNames := make([]string, len(ext.Synthetics))
	for i, s := range ext.Synthetics {
		synNames[i] = s.Name
	}
	sb.WriteString(fmt.Sprintf("| Synthetic Tests | %d | %s |\n", len(ext.Synthetics), joinResourceNames(synNames)))

	pipeNames := make([]string, len(ext.LogPipelines))
	for i, p := range ext.LogPipelines {
		pipeNames[i] = p.Name
	}
	sb.WriteString(fmt.Sprintf("| Log Pipelines | %d | %s |\n", len(ext.LogPipelines), joinResourceNames(pipeNames)))
	sb.WriteString(fmt.Sprintf("| Metric Metadata | %d | |\n", len(ext.Metrics)))

	dtNames := make([]string, len(ext.Downtimes))
	for i, d := range ext.Downtimes {
		dtNames[i] = d.Message
	}
	sb.WriteString(fmt.Sprintf("| Downtimes | %d | %s |\n", len(ext.Downtimes), joinResourceNames(dtNames)))

	notifNames := make([]string, len(ext.Notifications))
	for i, n := range ext.Notifications {
		notifNames[i] = n.Name
	}
	sb.WriteString(fmt.Sprintf("| Notification Channels | %d | %s |\n", len(ext.Notifications), joinResourceNames(notifNames)))

	nbNames := make([]string, len(ext.Notebooks))
	for i, n := range ext.Notebooks {
		nbNames[i] = n.Name
	}
	sb.WriteString(fmt.Sprintf("| Notebooks | %d | %s |\n", len(ext.Notebooks), joinResourceNames(nbNames)))

	r.sections = append(r.sections, section{
		title:   "Extracted Resources (DataDog)",
		content: sb.String(),
	})
}

// AddConversionSummary adds a summary of converted Dynatrace resources.
func (r *Report) AddConversionSummary(result *dynatrace.ConversionResult) {
	var sb strings.Builder
	sb.WriteString("| Resource Type | Count | Names |\n")
	sb.WriteString("|---|---|---|\n")

	dashNames := make([]string, len(result.Dashboards))
	for i, d := range result.Dashboards {
		dashNames[i] = d.DashboardMetadata.Name
	}
	sb.WriteString(fmt.Sprintf("| Dashboards | %d | %s |\n", len(result.Dashboards), joinResourceNames(dashNames)))

	meNames := make([]string, len(result.MetricEvents))
	for i, me := range result.MetricEvents {
		meNames[i] = me.Summary
	}
	sb.WriteString(fmt.Sprintf("| Metric Events | %d | %s |\n", len(result.MetricEvents), joinResourceNames(meNames)))

	sloNames := make([]string, len(result.SLOs))
	for i, s := range result.SLOs {
		sloNames[i] = s.Name
	}
	sb.WriteString(fmt.Sprintf("| SLOs | %d | %s |\n", len(result.SLOs), joinResourceNames(sloNames)))

	synNames := make([]string, len(result.Synthetics))
	for i, s := range result.Synthetics {
		synNames[i] = s.Name
	}
	sb.WriteString(fmt.Sprintf("| Synthetic Monitors | %d | %s |\n", len(result.Synthetics), joinResourceNames(synNames)))

	lrNames := make([]string, len(result.LogRules))
	for i, lr := range result.LogRules {
		lrNames[i] = lr.Name
	}
	sb.WriteString(fmt.Sprintf("| Log Processing Rules | %d | %s |\n", len(result.LogRules), joinResourceNames(lrNames)))
	sb.WriteString(fmt.Sprintf("| Metric Descriptors | %d | |\n", len(result.Metrics)))

	mwNames := make([]string, len(result.Maintenance))
	for i, mw := range result.Maintenance {
		mwNames[i] = mw.Name
	}
	sb.WriteString(fmt.Sprintf("| Maintenance Windows | %d | %s |\n", len(result.Maintenance), joinResourceNames(mwNames)))

	nNames := make([]string, len(result.Notifications))
	for i, n := range result.Notifications {
		nNames[i] = n.Name
	}
	sb.WriteString(fmt.Sprintf("| Notifications | %d | %s |\n", len(result.Notifications), joinResourceNames(nNames)))

	nbNames := make([]string, len(result.Notebooks))
	for i, nb := range result.Notebooks {
		nbNames[i] = nb.Name
	}
	sb.WriteString(fmt.Sprintf("| Notebooks | %d | %s |\n", len(result.Notebooks), joinResourceNames(nbNames)))

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

// AddDashboardDetails adds per-dashboard tile type breakdown.
func (r *Report) AddDashboardDetails(dashboards []dynatrace.Dashboard) {
	if len(dashboards) == 0 {
		return
	}

	var sb strings.Builder
	for _, d := range dashboards {
		tileCounts := map[string]int{}
		for _, t := range d.Tiles {
			tileCounts[t.TileType]++
		}
		sb.WriteString(fmt.Sprintf("### %s\n\n", d.DashboardMetadata.Name))
		sb.WriteString(fmt.Sprintf("Total tiles: %d\n\n", len(d.Tiles)))
		for tileType, count := range tileCounts {
			sb.WriteString(fmt.Sprintf("- %s: %d\n", tileType, count))
		}
		sb.WriteString("\n")
	}

	r.sections = append(r.sections, section{
		title:   "Dashboard Details",
		content: sb.String(),
	})
}

// AddDQLQueryNotes scans dashboard tiles for DQL patterns and lists queries
// that may need attention. In grail mode, DQL tiles are natively embedded;
// in classic mode, they fall back to MARKDOWN tiles requiring a Notebook.
func (r *Report) AddDQLQueryNotes(dashboards []dynatrace.Dashboard) {
	var markdownNotes, grailNotes []string
	for _, d := range dashboards {
		for _, t := range d.Tiles {
			if t.TileType == "MARKDOWN" && strings.Contains(t.Markdown, "fetch logs") {
				markdownNotes = append(markdownNotes, fmt.Sprintf("- Dashboard %q, tile %q contains a DQL query that requires a Dynatrace Notebook", d.DashboardMetadata.Name, t.Name))
			}
			if t.TileType == "DATA_EXPLORER" {
				for _, q := range t.Queries {
					if q.DQL != "" {
						grailNotes = append(grailNotes, fmt.Sprintf("- Dashboard %q, tile %q has a native DQL query (Grail)", d.DashboardMetadata.Name, t.Name))
					}
				}
			}
		}
	}

	if len(markdownNotes) == 0 && len(grailNotes) == 0 {
		return
	}

	var sb strings.Builder
	if len(grailNotes) > 0 {
		sb.WriteString("The following tiles contain native DQL queries (Grail-powered dashboard):\n\n")
		for _, note := range grailNotes {
			sb.WriteString(note + "\n")
		}
		sb.WriteString("\n")
	}
	if len(markdownNotes) > 0 {
		sb.WriteString("The following tiles contain DQL queries that cannot be displayed in classic Dynatrace dashboards.\nUse **Dynatrace Notebooks** to run these queries:\n\n")
		for _, note := range markdownNotes {
			sb.WriteString(note + "\n")
		}
	}

	r.sections = append(r.sections, section{
		title:   "DQL Query Notes",
		content: sb.String(),
	})
}

// AddValidationResults adds a metric selector validation section to the report.
func (r *Report) AddValidationResults(val *dynatrace.ValidationResult) {
	if val == nil {
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Validated **%d** selectors: %d valid, %d invalid, %d skipped\n\n",
		val.Summary.Total, val.Summary.Valid, val.Summary.Invalid, val.Summary.Skipped))

	sb.WriteString("| Selector | Source(s) | Status | Error |\n")
	sb.WriteString("|---|---|---|---|\n")

	for _, sv := range val.Selectors {
		status := "Valid"
		if sv.Skipped {
			status = "Skipped (placeholder)"
		} else if !sv.Valid {
			status = "Invalid"
		}
		sources := strings.Join(sv.Sources, "; ")
		errMsg := sv.Error
		// Escape pipes in error messages for markdown table
		errMsg = strings.ReplaceAll(errMsg, "|", "\\|")
		sb.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s |\n", sv.Selector, sources, status, errMsg))
	}

	r.sections = append(r.sections, section{
		title:   "Metric Selector Validation",
		content: sb.String(),
	})
}

// joinResourceNames joins up to 5 names, appending "+N more" if truncated.
func joinResourceNames(names []string) string {
	if len(names) == 0 {
		return ""
	}
	limit := 5
	if len(names) <= limit {
		return strings.Join(names, ", ")
	}
	return strings.Join(names[:limit], ", ") + fmt.Sprintf(" +%d more", len(names)-limit)
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
