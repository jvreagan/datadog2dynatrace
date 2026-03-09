package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/config"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/converter"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/importer"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/logging"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/ratelimit"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/report"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/terraform"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/ui"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "datadog2dynatrace",
		Short: "Convert DataDog monitoring configurations to Dynatrace",
		Long:  "A CLI tool that converts DataDog monitoring configurations to Dynatrace Cloud equivalents.",
	}

	rootCmd.AddCommand(convertCmd())
	rootCmd.AddCommand(validateCmd())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("datadog2dynatrace %s\n", config.Version)
		},
	}
}

func validateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate DataDog and Dynatrace credentials",
		RunE:  runValidate,
	}
	config.BindValidateFlags(cmd)
	return cmd
}

func convertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert",
		Short: "Convert DataDog resources to Dynatrace",
		RunE:  runConvert,
	}
	config.BindFlags(cmd)
	return cmd
}

func initLogging(cfg *config.Config) {
	switch {
	case cfg.Debug:
		logging.SetLevel(logging.LevelDebug)
	case cfg.Verbose:
		logging.SetLevel(logging.LevelInfo)
	}
	ratelimit.SetLogWriter(logging.Writer())
}

func runValidate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	initLogging(cfg)

	success := true

	// Validate DataDog
	fmt.Print("Validating DataDog credentials... ")
	if err := cfg.ValidateDataDog(); err != nil {
		color.Red("MISSING")
		fmt.Printf("  %s\n", err)
		success = false
	} else {
		ddClient := datadog.NewClient(cfg.DataDog.APIKey, cfg.DataDog.AppKey, cfg.DataDog.Site)
		if err := ddClient.Validate(); err != nil {
			color.Red("FAILED")
			fmt.Printf("  %s\n", err)
			success = false
		} else {
			color.Green("OK")
		}
	}

	// Validate Dynatrace
	fmt.Print("Validating Dynatrace credentials... ")
	if err := cfg.ValidateDynatrace(); err != nil {
		color.Red("MISSING")
		fmt.Printf("  %s\n", err)
		success = false
	} else {
		dtClient := dynatrace.NewClient(cfg.Dynatrace.EnvURL, cfg.Dynatrace.APIToken)
		if err := dtClient.Validate(); err != nil {
			color.Red("FAILED")
			fmt.Printf("  %s\n", err)
			success = false
		} else {
			color.Green("OK")
		}
	}

	if !success {
		return fmt.Errorf("credential validation failed")
	}
	return nil
}

func runConvert(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	initLogging(cfg)

	// Step 1: Extract from DataDog
	var extraction *datadog.ExtractionResult

	switch cfg.Source {
	case "api":
		if err := cfg.ValidateDataDog(); err != nil {
			return err
		}
		ddClient := datadog.NewClient(cfg.DataDog.APIKey, cfg.DataDog.AppKey, cfg.DataDog.Site)
		fmt.Println("Extracting resources from DataDog API...")
		extraction, err = ddClient.ExtractAll()
		if err != nil && cfg.FailFast {
			return fmt.Errorf("extraction failed: %w", err)
		}
		if err != nil {
			color.Yellow("Warning: some extraction errors occurred: %v", err)
		}
	case "file":
		if cfg.InputDir == "" {
			return fmt.Errorf("--input-dir is required when --source=file")
		}
		fmt.Printf("Importing resources from %s...\n", cfg.InputDir)
		extraction, err = importer.ImportFromDirectory(cfg.InputDir)
		if err != nil {
			return fmt.Errorf("import failed: %w", err)
		}
	default:
		return fmt.Errorf("invalid source: %s (must be 'api' or 'file')", cfg.Source)
	}

	// Step 2: Interactive selection (unless --all)
	resources := buildResourceList(extraction)
	if !cfg.All {
		selected, err := ui.SelectResources(resources)
		if err != nil {
			return fmt.Errorf("resource selection: %w", err)
		}
		extraction = filterExtraction(extraction, selected)
	}

	// Step 3: Convert
	fmt.Println("Converting resources...")
	conv := converter.New(converter.Options{EnableGrail: cfg.EnableGrail})
	result, convErrors := conv.ConvertAll(extraction)

	if len(convErrors) > 0 {
		if cfg.FailFast {
			return fmt.Errorf("conversion failed: %v", convErrors[0])
		}
		color.Yellow("Warning: %d conversion errors occurred", len(convErrors))
		for _, e := range convErrors {
			fmt.Printf("  - %s\n", e)
		}
	}

	// Step 3.5: Validate metric selectors (optional)
	var valResult *dynatrace.ValidationResult
	if cfg.Validate {
		if err := cfg.ValidateDynatrace(); err != nil {
			return fmt.Errorf("--validate requires Dynatrace credentials: %w", err)
		}
		dtValidateClient := dynatrace.NewClient(cfg.Dynatrace.EnvURL, cfg.Dynatrace.APIToken)
		fmt.Println("Validating metric selectors against Dynatrace API...")
		valResult = dtValidateClient.ValidateAll(result)
		printValidationSummary(valResult)
	}

	// Step 4: Output
	var pushErrors []error

	switch cfg.Target {
	case "terraform":
		if cfg.DryRun {
			fmt.Println("\n--- DRY RUN (Terraform) ---")
			printDryRunSummary(result)
			printTerraformFileList(result, cfg.OutputDir)
			fmt.Println("No files were written.")
		} else {
			fmt.Printf("Generating Terraform configs in %s...\n", cfg.OutputDir)
			gen := terraform.NewGenerator(cfg.OutputDir)
			if err := gen.GenerateAll(result); err != nil {
				return fmt.Errorf("terraform generation failed: %w", err)
			}
			color.Green("Terraform configs written to %s", cfg.OutputDir)
		}

	case "api":
		if err := cfg.ValidateDynatrace(); err != nil {
			return err
		}
		dtClient := dynatrace.NewClient(cfg.Dynatrace.EnvURL, cfg.Dynatrace.APIToken)

		if cfg.DryRun {
			fmt.Println("\n--- DRY RUN ---")
			printDryRunSummary(result)
			fmt.Println("No changes were made.")
		} else {
			fmt.Println("Pushing resources to Dynatrace...")
			pushErrors = dtClient.PushAllWithOptions(result, dynatrace.PushOptions{
				SkipExisting: cfg.SkipExisting,
			})
		}

	case "json":
		if cfg.DryRun {
			fmt.Println("\n--- DRY RUN (JSON) ---")
			printDryRunSummary(result)
			fmt.Println("No files were written.")
		} else {
			if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
				return fmt.Errorf("creating output directory: %w", err)
			}
			data, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return fmt.Errorf("marshalling JSON: %w", err)
			}
			outPath := filepath.Join(cfg.OutputDir, "dynatrace-resources.json")
			if err := os.WriteFile(outPath, data, 0644); err != nil {
				return fmt.Errorf("writing JSON output: %w", err)
			}
			color.Green("JSON output written to %s", outPath)
		}

	default:
		return fmt.Errorf("invalid target: %s (must be 'api', 'terraform', or 'json')", cfg.Target)
	}

	// Step 5: Generate report
	rpt := report.New()
	rpt.SetSource(cfg.Source, cfg.InputDir)
	rpt.SetTarget(cfg.Target, cfg.OutputDir)
	rpt.SetDryRun(cfg.DryRun)
	rpt.AddExtractionSummary(extraction)
	rpt.AddConversionErrors(convErrors)
	rpt.AddPushErrors(pushErrors)
	rpt.AddConversionSummary(result)
	rpt.AddDashboardDetails(result.Dashboards)
	rpt.AddDQLQueryNotes(result.Dashboards)
	rpt.AddValidationResults(valResult)

	if err := rpt.WriteToFile(cfg.ReportFile); err != nil {
		color.Yellow("Warning: could not write migration report: %v", err)
	} else {
		fmt.Printf("Migration report written to %s\n", cfg.ReportFile)
	}

	if len(pushErrors) > 0 {
		color.Yellow("\n%d push errors occurred:", len(pushErrors))
		for _, e := range pushErrors {
			fmt.Printf("  - %s\n", e)
		}
		if cfg.FailFast {
			return fmt.Errorf("push failed")
		}
	}

	color.Green("\nConversion complete!")
	return nil
}

type resourceItem struct {
	Type  string
	Name  string
	Count int
}

func buildResourceList(ext *datadog.ExtractionResult) []ui.ResourceGroup {
	var groups []ui.ResourceGroup

	if len(ext.Dashboards) > 0 {
		items := make([]ui.ResourceItem, len(ext.Dashboards))
		for i, d := range ext.Dashboards {
			items[i] = ui.ResourceItem{ID: d.ID, Name: d.Title}
		}
		groups = append(groups, ui.ResourceGroup{Type: "dashboards", Label: "Dashboards", Items: items})
	}

	if len(ext.Monitors) > 0 {
		items := make([]ui.ResourceItem, len(ext.Monitors))
		for i, m := range ext.Monitors {
			items[i] = ui.ResourceItem{ID: fmt.Sprintf("%d", m.ID), Name: m.Name}
		}
		groups = append(groups, ui.ResourceGroup{Type: "monitors", Label: "Monitors", Items: items})
	}

	if len(ext.SLOs) > 0 {
		items := make([]ui.ResourceItem, len(ext.SLOs))
		for i, s := range ext.SLOs {
			items[i] = ui.ResourceItem{ID: s.ID, Name: s.Name}
		}
		groups = append(groups, ui.ResourceGroup{Type: "slos", Label: "SLOs", Items: items})
	}

	if len(ext.Synthetics) > 0 {
		items := make([]ui.ResourceItem, len(ext.Synthetics))
		for i, s := range ext.Synthetics {
			items[i] = ui.ResourceItem{ID: s.PublicID, Name: s.Name}
		}
		groups = append(groups, ui.ResourceGroup{Type: "synthetics", Label: "Synthetic Tests", Items: items})
	}

	if len(ext.LogPipelines) > 0 {
		items := make([]ui.ResourceItem, len(ext.LogPipelines))
		for i, l := range ext.LogPipelines {
			items[i] = ui.ResourceItem{ID: l.ID, Name: l.Name}
		}
		groups = append(groups, ui.ResourceGroup{Type: "logs", Label: "Log Pipelines", Items: items})
	}

	if len(ext.Metrics) > 0 {
		items := make([]ui.ResourceItem, len(ext.Metrics))
		for i, m := range ext.Metrics {
			items[i] = ui.ResourceItem{ID: m.Metric, Name: m.Metric}
		}
		groups = append(groups, ui.ResourceGroup{Type: "metrics", Label: "Metric Metadata", Items: items})
	}

	if len(ext.Downtimes) > 0 {
		items := make([]ui.ResourceItem, len(ext.Downtimes))
		for i, d := range ext.Downtimes {
			items[i] = ui.ResourceItem{ID: fmt.Sprintf("%d", d.ID), Name: d.Message}
		}
		groups = append(groups, ui.ResourceGroup{Type: "downtimes", Label: "Downtimes", Items: items})
	}

	if len(ext.Notifications) > 0 {
		items := make([]ui.ResourceItem, len(ext.Notifications))
		for i, n := range ext.Notifications {
			items[i] = ui.ResourceItem{ID: fmt.Sprintf("%d", n.ID), Name: n.Name}
		}
		groups = append(groups, ui.ResourceGroup{Type: "notifications", Label: "Notification Channels", Items: items})
	}

	if len(ext.Notebooks) > 0 {
		items := make([]ui.ResourceItem, len(ext.Notebooks))
		for i, n := range ext.Notebooks {
			items[i] = ui.ResourceItem{ID: fmt.Sprintf("%d", n.ID), Name: n.Name}
		}
		groups = append(groups, ui.ResourceGroup{Type: "notebooks", Label: "Notebooks", Items: items})
	}

	return groups
}

func filterExtraction(ext *datadog.ExtractionResult, selected map[string][]string) *datadog.ExtractionResult {
	result := &datadog.ExtractionResult{}

	if ids, ok := selected["dashboards"]; ok {
		idSet := toSet(ids)
		for _, d := range ext.Dashboards {
			if idSet[d.ID] {
				result.Dashboards = append(result.Dashboards, d)
			}
		}
	}

	if ids, ok := selected["monitors"]; ok {
		idSet := toSet(ids)
		for _, m := range ext.Monitors {
			if idSet[fmt.Sprintf("%d", m.ID)] {
				result.Monitors = append(result.Monitors, m)
			}
		}
	}

	if ids, ok := selected["slos"]; ok {
		idSet := toSet(ids)
		for _, s := range ext.SLOs {
			if idSet[s.ID] {
				result.SLOs = append(result.SLOs, s)
			}
		}
	}

	if ids, ok := selected["synthetics"]; ok {
		idSet := toSet(ids)
		for _, s := range ext.Synthetics {
			if idSet[s.PublicID] {
				result.Synthetics = append(result.Synthetics, s)
			}
		}
	}

	if ids, ok := selected["logs"]; ok {
		idSet := toSet(ids)
		for _, l := range ext.LogPipelines {
			if idSet[l.ID] {
				result.LogPipelines = append(result.LogPipelines, l)
			}
		}
	}

	if ids, ok := selected["metrics"]; ok {
		idSet := toSet(ids)
		for _, m := range ext.Metrics {
			if idSet[m.Metric] {
				result.Metrics = append(result.Metrics, m)
			}
		}
	}

	if ids, ok := selected["downtimes"]; ok {
		idSet := toSet(ids)
		for _, d := range ext.Downtimes {
			if idSet[fmt.Sprintf("%d", d.ID)] {
				result.Downtimes = append(result.Downtimes, d)
			}
		}
	}

	if ids, ok := selected["notifications"]; ok {
		idSet := toSet(ids)
		for _, n := range ext.Notifications {
			if idSet[fmt.Sprintf("%d", n.ID)] {
				result.Notifications = append(result.Notifications, n)
			}
		}
	}

	if ids, ok := selected["notebooks"]; ok {
		idSet := toSet(ids)
		for _, n := range ext.Notebooks {
			if idSet[fmt.Sprintf("%d", n.ID)] {
				result.Notebooks = append(result.Notebooks, n)
			}
		}
	}

	return result
}

func toSet(items []string) map[string]bool {
	s := make(map[string]bool, len(items))
	for _, item := range items {
		s[item] = true
	}
	return s
}

func printDryRunSummary(result *dynatrace.ConversionResult) {
	fmt.Println("\nResources to be created:")

	dashNames := make([]string, len(result.Dashboards))
	for i, d := range result.Dashboards {
		dashNames[i] = d.DashboardMetadata.Name
	}
	printResourceGroup("Dashboards", dashNames)

	meNames := make([]string, len(result.MetricEvents))
	for i, me := range result.MetricEvents {
		meNames[i] = me.Summary
	}
	printResourceGroup("Metric Events", meNames)

	sloNames := make([]string, len(result.SLOs))
	for i, s := range result.SLOs {
		sloNames[i] = s.Name
	}
	printResourceGroup("SLOs", sloNames)

	synNames := make([]string, len(result.Synthetics))
	for i, s := range result.Synthetics {
		synNames[i] = s.Name
	}
	printResourceGroup("Synthetic Monitors", synNames)

	lrNames := make([]string, len(result.LogRules))
	for i, r := range result.LogRules {
		lrNames[i] = r.Name
	}
	printResourceGroup("Log Processing Rules", lrNames)

	mdNames := make([]string, len(result.Metrics))
	for i, m := range result.Metrics {
		mdNames[i] = m.MetricID
	}
	printResourceGroup("Metric Descriptors", mdNames)

	mwNames := make([]string, len(result.Maintenance))
	for i, mw := range result.Maintenance {
		mwNames[i] = mw.Name
	}
	printResourceGroup("Maintenance Windows", mwNames)

	nNames := make([]string, len(result.Notifications))
	for i, n := range result.Notifications {
		nNames[i] = n.Name
	}
	printResourceGroup("Notifications", nNames)

	nbNames := make([]string, len(result.Notebooks))
	for i, nb := range result.Notebooks {
		nbNames[i] = nb.Name
	}
	printResourceGroup("Notebooks", nbNames)
}

func printResourceGroup(label string, names []string) {
	fmt.Printf("  %-24s %d\n", label+":", len(names))
	for _, name := range names {
		fmt.Printf("    - %s\n", name)
	}
}

func printValidationSummary(val *dynatrace.ValidationResult) {
	fmt.Printf("Validated %d selectors: %d valid, %d invalid, %d skipped\n",
		val.Summary.Total, val.Summary.Valid, val.Summary.Invalid, val.Summary.Skipped)
	for _, sv := range val.Selectors {
		if !sv.Valid && !sv.Skipped {
			color.Yellow("  INVALID: %s", sv.Selector)
			for _, src := range sv.Sources {
				fmt.Printf("    source: %s\n", src)
			}
			fmt.Printf("    error: %s\n", sv.Error)
		}
	}
}

func printTerraformFileList(result *dynatrace.ConversionResult, outputDir string) {
	fmt.Println("\nTerraform files that would be generated:")
	fmt.Printf("  %s/provider.tf\n", outputDir)
	if len(result.Dashboards) > 0 {
		fmt.Printf("  %s/dashboards.tf\n", outputDir)
	}
	if len(result.MetricEvents) > 0 {
		fmt.Printf("  %s/metric_events.tf\n", outputDir)
	}
	if len(result.SLOs) > 0 {
		fmt.Printf("  %s/slos.tf\n", outputDir)
	}
	if len(result.Synthetics) > 0 {
		fmt.Printf("  %s/synthetics.tf\n", outputDir)
	}
	if len(result.LogRules) > 0 {
		fmt.Printf("  %s/log_processing.tf\n", outputDir)
	}
	if len(result.Maintenance) > 0 {
		fmt.Printf("  %s/maintenance.tf\n", outputDir)
	}
	if len(result.Notifications) > 0 {
		fmt.Printf("  %s/notifications.tf\n", outputDir)
	}
	if len(result.Notebooks) > 0 {
		fmt.Printf("  %s/notebooks.tf\n", outputDir)
	}
}
