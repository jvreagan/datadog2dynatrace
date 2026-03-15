package config

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestValidateDataDogMissingAPIKey(t *testing.T) {
	cfg := &Config{DataDog: DataDogConfig{AppKey: "test"}}
	if err := cfg.ValidateDataDog(); err == nil {
		t.Error("expected error for missing API key")
	}
}

func TestValidateDataDogMissingAppKey(t *testing.T) {
	cfg := &Config{DataDog: DataDogConfig{APIKey: "test"}}
	if err := cfg.ValidateDataDog(); err == nil {
		t.Error("expected error for missing app key")
	}
}

func TestValidateDynatraceMissingURL(t *testing.T) {
	cfg := &Config{Dynatrace: DynatraceConfig{APIToken: "test"}}
	if err := cfg.ValidateDynatrace(); err == nil {
		t.Error("expected error for missing URL")
	}
}

func TestValidateDynatraceMissingAuth(t *testing.T) {
	cfg := &Config{Dynatrace: DynatraceConfig{EnvURL: "https://test.dynatrace.com"}}
	if err := cfg.ValidateDynatrace(); err == nil {
		t.Error("expected error for missing auth")
	}
}

func TestValidateDataDogOK(t *testing.T) {
	cfg := &Config{DataDog: DataDogConfig{APIKey: "key", AppKey: "app"}}
	if err := cfg.ValidateDataDog(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateDynatraceOK(t *testing.T) {
	cfg := &Config{Dynatrace: DynatraceConfig{EnvURL: "https://test.dynatrace.com", APIToken: "tok"}}
	if err := cfg.ValidateDynatrace(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateDataDogErrorMessages(t *testing.T) {
	cfg := &Config{}
	err := cfg.ValidateDataDog()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "API key") {
		t.Errorf("expected 'API key' in error, got %q", err.Error())
	}

	cfg.DataDog.APIKey = "key"
	err = cfg.ValidateDataDog()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Application key") {
		t.Errorf("expected 'Application key' in error, got %q", err.Error())
	}
}

func TestValidateDynatraceErrorMessages(t *testing.T) {
	cfg := &Config{}
	err := cfg.ValidateDynatrace()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "environment URL") {
		t.Errorf("expected 'environment URL' in error, got %q", err.Error())
	}

	cfg.Dynatrace.EnvURL = "https://test.dynatrace.com"
	err = cfg.ValidateDynatrace()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "auth is required") {
		t.Errorf("expected 'auth is required' in error, got %q", err.Error())
	}
}

func TestBindFlags(t *testing.T) {
	// Reset viper state
	viper.Reset()

	cmd := &cobra.Command{Use: "test"}
	BindFlags(cmd)

	expectedFlags := []string{
		"source", "input-dir", "target", "output-dir",
		"dry-run", "skip-existing", "fail-fast", "all",
		"report-file", "verbose", "debug", "enable-grail", "validate",
		"dd-api-key", "dd-app-key", "dd-site",
		"dt-env-url", "dt-api-token",
	}
	for _, name := range expectedFlags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to be registered", name)
		}
	}

	// Check defaults
	if v, _ := cmd.Flags().GetString("source"); v != "api" {
		t.Errorf("source default: got %q, want %q", v, "api")
	}
	if v, _ := cmd.Flags().GetString("target"); v != "terraform" {
		t.Errorf("target default: got %q, want %q", v, "terraform")
	}
	if v, _ := cmd.Flags().GetString("dd-site"); v != "datadoghq.com" {
		t.Errorf("dd-site default: got %q, want %q", v, "datadoghq.com")
	}
	if v, _ := cmd.Flags().GetBool("dry-run"); v != false {
		t.Error("dry-run should default to false")
	}
	if v, _ := cmd.Flags().GetBool("skip-existing"); v != true {
		t.Error("skip-existing should default to true")
	}
}

func TestBindValidateFlags(t *testing.T) {
	viper.Reset()

	cmd := &cobra.Command{Use: "test"}
	BindValidateFlags(cmd)

	expectedFlags := []string{
		"dd-api-key", "dd-app-key", "dd-site",
		"dt-env-url", "dt-api-token",
		"verbose", "debug",
	}
	for _, name := range expectedFlags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to be registered", name)
		}
	}

	// Validate flags should NOT include migrate-specific flags
	for _, name := range []string{"source", "target", "output-dir", "dry-run"} {
		if cmd.Flags().Lookup(name) != nil {
			t.Errorf("validate command should not have flag %q", name)
		}
	}
}

func TestLoadDefaults(t *testing.T) {
	viper.Reset()

	// Load with no config file — should succeed with defaults
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Source != "api" {
		t.Errorf("source default: got %q, want %q", cfg.Source, "api")
	}
	if cfg.Target != "terraform" {
		t.Errorf("target default: got %q, want %q", cfg.Target, "terraform")
	}
	if cfg.OutputDir != "./dynatrace-terraform/" {
		t.Errorf("output_dir default: got %q, want %q", cfg.OutputDir, "./dynatrace-terraform/")
	}
	if cfg.ReportFile != "./migration-report.md" {
		t.Errorf("report_file default: got %q, want %q", cfg.ReportFile, "./migration-report.md")
	}
	if !cfg.SkipExisting {
		t.Error("skip_existing should default to true")
	}
	if cfg.DataDog.Site != "datadoghq.com" {
		t.Errorf("datadog.site default: got %q, want %q", cfg.DataDog.Site, "datadoghq.com")
	}
}

func TestLoadFromEnv(t *testing.T) {
	viper.Reset()

	os.Setenv("DD_API_KEY", "env-api-key")
	os.Setenv("DD_APP_KEY", "env-app-key")
	os.Setenv("DT_ENV_URL", "https://env.dynatrace.com")
	os.Setenv("DT_API_TOKEN", "env-token")
	defer func() {
		os.Unsetenv("DD_API_KEY")
		os.Unsetenv("DD_APP_KEY")
		os.Unsetenv("DT_ENV_URL")
		os.Unsetenv("DT_API_TOKEN")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DataDog.APIKey != "env-api-key" {
		t.Errorf("DD_API_KEY: got %q, want %q", cfg.DataDog.APIKey, "env-api-key")
	}
	if cfg.DataDog.AppKey != "env-app-key" {
		t.Errorf("DD_APP_KEY: got %q, want %q", cfg.DataDog.AppKey, "env-app-key")
	}
	if cfg.Dynatrace.EnvURL != "https://env.dynatrace.com" {
		t.Errorf("DT_ENV_URL: got %q, want %q", cfg.Dynatrace.EnvURL, "https://env.dynatrace.com")
	}
	if cfg.Dynatrace.APIToken != "env-token" {
		t.Errorf("DT_API_TOKEN: got %q, want %q", cfg.Dynatrace.APIToken, "env-token")
	}
}

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
	if Version != "dev" {
		t.Logf("Version is %q (overridden from default 'dev')", Version)
	}
}

func TestValidateDynatraceOAuthOK(t *testing.T) {
	cfg := &Config{Dynatrace: DynatraceConfig{
		EnvURL:       "https://test.live.dynatrace.com",
		ClientID:     "dt0s02.client",
		ClientSecret: "dt0s02.client.secret",
	}}
	if err := cfg.ValidateDynatrace(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateDynatraceOAuthMissingSecret(t *testing.T) {
	cfg := &Config{Dynatrace: DynatraceConfig{
		EnvURL:   "https://test.live.dynatrace.com",
		ClientID: "dt0s02.client",
	}}
	err := cfg.ValidateDynatrace()
	if err == nil {
		t.Fatal("expected error for missing client secret")
	}
	if !strings.Contains(err.Error(), "client secret") {
		t.Errorf("expected 'client secret' in error, got %q", err.Error())
	}
}

func TestValidateDynatraceOAuthMissingID(t *testing.T) {
	cfg := &Config{Dynatrace: DynatraceConfig{
		EnvURL:       "https://test.live.dynatrace.com",
		ClientSecret: "dt0s02.client.secret",
	}}
	err := cfg.ValidateDynatrace()
	if err == nil {
		t.Fatal("expected error for missing client ID")
	}
	if !strings.Contains(err.Error(), "client ID") {
		t.Errorf("expected 'client ID' in error, got %q", err.Error())
	}
}

func TestValidateDynatraceBothTokenAndOAuth(t *testing.T) {
	cfg := &Config{Dynatrace: DynatraceConfig{
		EnvURL:       "https://test.live.dynatrace.com",
		APIToken:     "dt0c01.token",
		ClientID:     "dt0s02.client",
		ClientSecret: "dt0s02.client.secret",
	}}
	// Both are acceptable, OAuth takes priority in code but validation passes
	if err := cfg.ValidateDynatrace(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBindFlagsIncludesOAuth(t *testing.T) {
	viper.Reset()
	cmd := &cobra.Command{Use: "test"}
	BindFlags(cmd)

	for _, name := range []string{"dt-client-id", "dt-client-secret"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to be registered", name)
		}
	}
}

func TestBindValidateFlagsIncludesOAuth(t *testing.T) {
	viper.Reset()
	cmd := &cobra.Command{Use: "test"}
	BindValidateFlags(cmd)

	for _, name := range []string{"dt-client-id", "dt-client-secret"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to be registered", name)
		}
	}
}
