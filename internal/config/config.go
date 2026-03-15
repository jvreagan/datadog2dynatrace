package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Version = "dev"

type Config struct {
	DataDog   DataDogConfig   `mapstructure:"datadog"`
	Dynatrace DynatraceConfig `mapstructure:"dynatrace"`
	Source    string          `mapstructure:"source"`
	InputDir  string          `mapstructure:"input_dir"`
	Target    string          `mapstructure:"target"`
	OutputDir string          `mapstructure:"output_dir"`
	DryRun       bool            `mapstructure:"dry_run"`
	SkipExisting bool            `mapstructure:"skip_existing"`
	FailFast  bool            `mapstructure:"fail_fast"`
	All        bool            `mapstructure:"all"`
	ReportFile string          `mapstructure:"report_file"`
	Verbose    bool            `mapstructure:"verbose"`
	Debug      bool            `mapstructure:"debug"`
	EnableGrail bool           `mapstructure:"enable_grail"`
	Validate    bool           `mapstructure:"validate"`
}

type DataDogConfig struct {
	APIKey string `mapstructure:"api_key"`
	AppKey string `mapstructure:"app_key"`
	Site   string `mapstructure:"site"`
}

type DynatraceConfig struct {
	EnvURL       string `mapstructure:"env_url"`
	APIToken     string `mapstructure:"api_token"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

func (c *Config) ValidateDataDog() error {
	if c.DataDog.APIKey == "" {
		return fmt.Errorf("DataDog API key is required (--dd-api-key, config file, or DD_API_KEY env var)")
	}
	if c.DataDog.AppKey == "" {
		return fmt.Errorf("DataDog Application key is required (--dd-app-key, config file, or DD_APP_KEY env var)")
	}
	return nil
}

func (c *Config) ValidateDynatrace() error {
	if c.Dynatrace.EnvURL == "" {
		return fmt.Errorf("Dynatrace environment URL is required (--dt-env-url, config file, or DT_ENV_URL env var)")
	}
	// Check for partial OAuth creds first (more specific error)
	if c.Dynatrace.ClientID != "" && c.Dynatrace.ClientSecret == "" {
		return fmt.Errorf("Dynatrace client secret is required when client ID is set (--dt-client-secret or DT_CLIENT_SECRET)")
	}
	if c.Dynatrace.ClientSecret != "" && c.Dynatrace.ClientID == "" {
		return fmt.Errorf("Dynatrace client ID is required when client secret is set (--dt-client-id or DT_CLIENT_ID)")
	}
	hasToken := c.Dynatrace.APIToken != ""
	hasOAuth := c.Dynatrace.ClientID != "" && c.Dynatrace.ClientSecret != ""
	if !hasToken && !hasOAuth {
		return fmt.Errorf("Dynatrace auth is required: set --dt-api-token (or DT_API_TOKEN) for token auth, or --dt-client-id and --dt-client-secret (or DT_CLIENT_ID/DT_CLIENT_SECRET) for OAuth")
	}
	return nil
}

func BindFlags(cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.String("source", "api", "Input source: api or file")
	flags.String("input-dir", "", "Directory with DD export files (when source=file)")
	flags.String("target", "terraform", "Output target: api, terraform, or json")
	flags.String("output-dir", "./dynatrace-terraform/", "Terraform output directory")
	flags.Bool("dry-run", false, "Preview without pushing")
	flags.Bool("skip-existing", true, "Skip resources that already exist in Dynatrace")
	flags.Bool("fail-fast", false, "Stop on first error")
	flags.Bool("all", false, "Convert all resources (skip selection)")
	flags.String("report-file", "./migration-report.md", "Migration report path")
	flags.Bool("verbose", false, "Enable verbose output (info-level logging)")
	flags.Bool("debug", false, "Enable debug output (debug-level logging)")
	flags.Bool("enable-grail", false, "Emit native DQL tiles for Grail-powered dashboards")
	flags.Bool("validate", false, "Validate metric selectors against Dynatrace API")

	flags.String("dd-api-key", "", "DataDog API key")
	flags.String("dd-app-key", "", "DataDog Application key")
	flags.String("dd-site", "datadoghq.com", "DataDog site")
	flags.String("dt-env-url", "", "Dynatrace environment URL")
	flags.String("dt-api-token", "", "Dynatrace API token")
	flags.String("dt-client-id", "", "Dynatrace OAuth client ID (for Gen3/Platform)")
	flags.String("dt-client-secret", "", "Dynatrace OAuth client secret (for Gen3/Platform)")

	viper.BindPFlag("source", flags.Lookup("source"))
	viper.BindPFlag("input_dir", flags.Lookup("input-dir"))
	viper.BindPFlag("target", flags.Lookup("target"))
	viper.BindPFlag("output_dir", flags.Lookup("output-dir"))
	viper.BindPFlag("dry_run", flags.Lookup("dry-run"))
	viper.BindPFlag("skip_existing", flags.Lookup("skip-existing"))
	viper.BindPFlag("fail_fast", flags.Lookup("fail-fast"))
	viper.BindPFlag("all", flags.Lookup("all"))
	viper.BindPFlag("report_file", flags.Lookup("report-file"))
	viper.BindPFlag("verbose", flags.Lookup("verbose"))
	viper.BindPFlag("debug", flags.Lookup("debug"))
	viper.BindPFlag("enable_grail", flags.Lookup("enable-grail"))
	viper.BindPFlag("validate", flags.Lookup("validate"))

	viper.BindPFlag("datadog.api_key", flags.Lookup("dd-api-key"))
	viper.BindPFlag("datadog.app_key", flags.Lookup("dd-app-key"))
	viper.BindPFlag("datadog.site", flags.Lookup("dd-site"))
	viper.BindPFlag("dynatrace.env_url", flags.Lookup("dt-env-url"))
	viper.BindPFlag("dynatrace.api_token", flags.Lookup("dt-api-token"))
	viper.BindPFlag("dynatrace.client_id", flags.Lookup("dt-client-id"))
	viper.BindPFlag("dynatrace.client_secret", flags.Lookup("dt-client-secret"))
}

func BindValidateFlags(cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.String("dd-api-key", "", "DataDog API key")
	flags.String("dd-app-key", "", "DataDog Application key")
	flags.String("dd-site", "datadoghq.com", "DataDog site")
	flags.String("dt-env-url", "", "Dynatrace environment URL")
	flags.String("dt-api-token", "", "Dynatrace API token")
	flags.String("dt-client-id", "", "Dynatrace OAuth client ID (for Gen3/Platform)")
	flags.String("dt-client-secret", "", "Dynatrace OAuth client secret (for Gen3/Platform)")
	flags.Bool("verbose", false, "Enable verbose output (info-level logging)")
	flags.Bool("debug", false, "Enable debug output (debug-level logging)")

	viper.BindPFlag("datadog.api_key", flags.Lookup("dd-api-key"))
	viper.BindPFlag("datadog.app_key", flags.Lookup("dd-app-key"))
	viper.BindPFlag("datadog.site", flags.Lookup("dd-site"))
	viper.BindPFlag("dynatrace.env_url", flags.Lookup("dt-env-url"))
	viper.BindPFlag("dynatrace.api_token", flags.Lookup("dt-api-token"))
	viper.BindPFlag("dynatrace.client_id", flags.Lookup("dt-client-id"))
	viper.BindPFlag("dynatrace.client_secret", flags.Lookup("dt-client-secret"))
	viper.BindPFlag("verbose", flags.Lookup("verbose"))
	viper.BindPFlag("debug", flags.Lookup("debug"))
}

// envFallbacks applies environment variables only for keys not already set
// by CLI flags or the config file. This enforces the documented precedence:
// CLI flags > config file > environment variables > defaults.
var envFallbacks = []struct {
	key    string
	envVar string
}{
	{"datadog.api_key", "DD_API_KEY"},
	{"datadog.app_key", "DD_APP_KEY"},
	{"datadog.site", "DD_SITE"},
	{"dynatrace.env_url", "DT_ENV_URL"},
	{"dynatrace.api_token", "DT_API_TOKEN"},
	{"dynatrace.client_id", "DT_CLIENT_ID"},
	{"dynatrace.client_secret", "DT_CLIENT_SECRET"},
}

func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("finding home directory: %w", err)
	}

	viper.SetConfigName(".datadog2dynatrace")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(home)

	// Defaults
	viper.SetDefault("datadog.site", "datadoghq.com")
	viper.SetDefault("source", "api")
	viper.SetDefault("target", "terraform")
	viper.SetDefault("output_dir", "./dynatrace-terraform/")
	viper.SetDefault("report_file", "./migration-report.md")
	viper.SetDefault("skip_existing", true)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Only warn if config exists but can't be read
			if _, statErr := os.Stat(filepath.Join(home, ".datadog2dynatrace.yaml")); statErr == nil {
				return nil, fmt.Errorf("reading config file: %w", err)
			}
		}
	}

	// Apply env vars only as fallback (when not set by flags or config file).
	for _, fb := range envFallbacks {
		if !viper.IsSet(fb.key) {
			if val := os.Getenv(fb.envVar); val != "" {
				viper.Set(fb.key, val)
			}
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}
