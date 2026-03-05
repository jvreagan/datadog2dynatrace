package datadog

import "time"

// Dashboard represents a DataDog dashboard.
type Dashboard struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	LayoutType  string           `json:"layout_type"` // "ordered" or "free"
	Widgets     []Widget         `json:"widgets"`
	TemplateVars []TemplateVar   `json:"template_variables,omitempty"`
	CreatedAt   time.Time        `json:"created_at,omitempty"`
	ModifiedAt  time.Time        `json:"modified_at,omitempty"`
}

type Widget struct {
	ID         int64           `json:"id,omitempty"`
	Definition WidgetDefinition `json:"definition"`
	Layout     *WidgetLayout   `json:"layout,omitempty"`
}

type WidgetDefinition struct {
	Type     string          `json:"type"`
	Title    string          `json:"title,omitempty"`
	Requests []WidgetRequest `json:"requests,omitempty"`
	// For group widgets
	Widgets []Widget `json:"widgets,omitempty"`
	// For note/free text widgets
	Content   string `json:"content,omitempty"`
	FontSize  string `json:"font_size,omitempty"`
	TextAlign string `json:"text_align,omitempty"`
	// For query value widgets
	Precision int    `json:"precision,omitempty"`
	Autoscale bool   `json:"autoscale,omitempty"`
	// For timeseries
	Yaxis    *Axis  `json:"yaxis,omitempty"`
	Markers  []Marker `json:"markers,omitempty"`
}

type WidgetRequest struct {
	Query        string        `json:"q,omitempty"`
	Queries      []QueryDef    `json:"queries,omitempty"`
	Formulas     []Formula     `json:"formulas,omitempty"`
	DisplayType  string        `json:"display_type,omitempty"`
	Style        *WidgetStyle  `json:"style,omitempty"`
	ResponseFormat string      `json:"response_format,omitempty"`
	// For log/APM queries
	LogQuery  *LogQuery  `json:"log_query,omitempty"`
	ApmQuery  *ApmQuery  `json:"apm_query,omitempty"`
}

type QueryDef struct {
	Name       string `json:"name"`
	DataSource string `json:"data_source"`
	Query      string `json:"query"`
}

type Formula struct {
	Formula string `json:"formula"`
	Alias   string `json:"alias,omitempty"`
}

type LogQuery struct {
	Index   string       `json:"index"`
	Search  *SearchQuery `json:"search,omitempty"`
	Compute *Compute     `json:"compute,omitempty"`
	GroupBy []GroupBy    `json:"group_by,omitempty"`
}

type ApmQuery struct {
	Index   string       `json:"index"`
	Search  *SearchQuery `json:"search,omitempty"`
	Compute *Compute     `json:"compute,omitempty"`
	GroupBy []GroupBy    `json:"group_by,omitempty"`
}

type SearchQuery struct {
	Query string `json:"query"`
}

type Compute struct {
	Aggregation string `json:"aggregation"`
	Facet       string `json:"facet,omitempty"`
	Interval    int    `json:"interval,omitempty"`
}

type GroupBy struct {
	Facet string `json:"facet"`
	Limit int    `json:"limit,omitempty"`
	Sort  *Sort  `json:"sort,omitempty"`
}

type Sort struct {
	Aggregation string `json:"aggregation"`
	Order       string `json:"order"`
}

type WidgetLayout struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type WidgetStyle struct {
	Palette   string `json:"palette,omitempty"`
	LineType  string `json:"line_type,omitempty"`
	LineWidth string `json:"line_width,omitempty"`
}

type Axis struct {
	Min       string `json:"min,omitempty"`
	Max       string `json:"max,omitempty"`
	Scale     string `json:"scale,omitempty"`
	Label     string `json:"label,omitempty"`
	IncludeZero bool `json:"include_zero,omitempty"`
}

type Marker struct {
	Value       string `json:"value"`
	DisplayType string `json:"display_type,omitempty"`
	Label       string `json:"label,omitempty"`
}

type TemplateVar struct {
	Name    string   `json:"name"`
	Prefix  string   `json:"prefix,omitempty"`
	Default string   `json:"default,omitempty"`
	AvailableValues []string `json:"available_values,omitempty"`
}

// Monitor represents a DataDog monitor.
type Monitor struct {
	ID              int64          `json:"id"`
	Name            string         `json:"name"`
	Type            string         `json:"type"`
	Query           string         `json:"query"`
	Message         string         `json:"message"`
	Tags            []string       `json:"tags"`
	Options         MonitorOptions `json:"options"`
	OverallState    string         `json:"overall_state,omitempty"`
	CreatedAt       time.Time      `json:"created_at,omitempty"`
	ModifiedAt      time.Time      `json:"modified_at,omitempty"`
}

type MonitorOptions struct {
	Thresholds        *Thresholds `json:"thresholds,omitempty"`
	NotifyNoData      bool        `json:"notify_no_data,omitempty"`
	NoDataTimeframe   int         `json:"no_data_timeframe,omitempty"`
	NotifyAudit       bool        `json:"notify_audit,omitempty"`
	RenotifyInterval  int         `json:"renotify_interval,omitempty"`
	EscalationMessage string      `json:"escalation_message,omitempty"`
	IncludeTags       bool        `json:"include_tags,omitempty"`
	EvaluationDelay   int         `json:"evaluation_delay,omitempty"`
	NewGroupDelay     int         `json:"new_group_delay,omitempty"`
}

type Thresholds struct {
	Critical         *float64 `json:"critical,omitempty"`
	Warning          *float64 `json:"warning,omitempty"`
	OK               *float64 `json:"ok,omitempty"`
	CriticalRecovery *float64 `json:"critical_recovery,omitempty"`
	WarningRecovery  *float64 `json:"warning_recovery,omitempty"`
}

// SLO represents a DataDog Service Level Objective.
type SLO struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Type        string     `json:"type"` // "metric" or "monitor"
	Tags        []string   `json:"tags"`
	Query       *SLOQuery  `json:"query,omitempty"`
	MonitorIDs  []int64    `json:"monitor_ids,omitempty"`
	Thresholds  []SLOThreshold `json:"thresholds"`
	CreatedAt   int64      `json:"created_at,omitempty"`
	ModifiedAt  int64      `json:"modified_at,omitempty"`
}

type SLOQuery struct {
	Numerator   string `json:"numerator"`
	Denominator string `json:"denominator"`
}

type SLOThreshold struct {
	Timeframe string  `json:"timeframe"`
	Target    float64 `json:"target"`
	Warning   float64 `json:"warning,omitempty"`
}

// SyntheticTest represents a DataDog synthetic test.
type SyntheticTest struct {
	PublicID  string          `json:"public_id"`
	Name     string          `json:"name"`
	Type     string          `json:"type"` // "api" or "browser"
	SubType  string          `json:"subtype,omitempty"` // "http", "ssl", "dns", "tcp", "multi"
	Config   SyntheticConfig `json:"config"`
	Options  SyntheticOptions `json:"options"`
	Message  string          `json:"message,omitempty"`
	Tags     []string        `json:"tags"`
	Status   string          `json:"status"`
	Locations []string       `json:"locations"`
}

type SyntheticConfig struct {
	Request    *SyntheticRequest    `json:"request,omitempty"`
	Assertions []SyntheticAssertion `json:"assertions,omitempty"`
	// For browser tests
	Steps []BrowserStep `json:"steps,omitempty"`
}

type SyntheticRequest struct {
	Method  string            `json:"method,omitempty"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
	Timeout int               `json:"timeout,omitempty"`
}

type SyntheticAssertion struct {
	Type     string      `json:"type"`
	Operator string      `json:"operator"`
	Target   interface{} `json:"target"`
	Property string      `json:"property,omitempty"`
}

type BrowserStep struct {
	Name   string                 `json:"name"`
	Type   string                 `json:"type"`
	Params map[string]interface{} `json:"params,omitempty"`
}

type SyntheticOptions struct {
	TickEvery    int  `json:"tick_every,omitempty"`
	MinFailure   int  `json:"min_failure_duration,omitempty"`
	MinLocFailed int  `json:"min_location_failed,omitempty"`
	FollowRedirects bool `json:"follow_redirects,omitempty"`
	Retry        *RetryConfig `json:"retry,omitempty"`
}

type RetryConfig struct {
	Count    int `json:"count,omitempty"`
	Interval int `json:"interval,omitempty"`
}

// LogPipeline represents a DataDog log pipeline.
type LogPipeline struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	IsEnabled  bool            `json:"is_enabled"`
	Filter     *LogFilter      `json:"filter,omitempty"`
	Processors []LogProcessor  `json:"processors,omitempty"`
}

type LogFilter struct {
	Query string `json:"query"`
}

type LogProcessor struct {
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	IsEnabled  bool                   `json:"is_enabled"`
	Sources    []string               `json:"sources,omitempty"`
	Target     string                 `json:"target,omitempty"`
	Grok       *GrokRule              `json:"grok,omitempty"`
	// Generic storage for processor-specific config
	Config     map[string]interface{} `json:"-"`
}

type GrokRule struct {
	SupportRules string `json:"support_rules,omitempty"`
	MatchRules   string `json:"match_rules"`
}

// MetricMetadata represents DataDog metric metadata.
type MetricMetadata struct {
	Metric      string `json:"metric"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Unit        string `json:"unit,omitempty"`
	PerUnit     string `json:"per_unit,omitempty"`
	ShortName   string `json:"short_name,omitempty"`
}

// Downtime represents a DataDog downtime.
type Downtime struct {
	ID          int64    `json:"id"`
	Message     string   `json:"message,omitempty"`
	MonitorID   *int64   `json:"monitor_id,omitempty"`
	MonitorTags []string `json:"monitor_tags,omitempty"`
	Scope       []string `json:"scope"`
	Start       int64    `json:"start,omitempty"`
	End         *int64   `json:"end,omitempty"`
	Timezone    string   `json:"timezone,omitempty"`
	Recurrence  *DowntimeRecurrence `json:"recurrence,omitempty"`
	Disabled    bool     `json:"disabled,omitempty"`
}

type DowntimeRecurrence struct {
	Type        string   `json:"type"` // "days", "weeks", "months", "years"
	Period      int      `json:"period"`
	WeekDays    []string `json:"week_days,omitempty"`
	UntilDate   *int64   `json:"until_date,omitempty"`
	UntilOccurrences *int `json:"until_occurrences,omitempty"`
}

// NotificationChannel represents a DataDog notification integration.
type NotificationChannel struct {
	ID     int64                  `json:"id"`
	Name   string                 `json:"name"`
	Type   string                 `json:"type"` // "slack", "pagerduty", "email", "webhook", etc.
	Config map[string]interface{} `json:"config"`
}

// Notebook represents a DataDog notebook.
type Notebook struct {
	ID       int64          `json:"id"`
	Name     string         `json:"name"`
	Author   NotebookAuthor `json:"author,omitempty"`
	Cells    []NotebookCell `json:"cells"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Created  time.Time      `json:"created,omitempty"`
	Modified time.Time      `json:"modified,omitempty"`
}

type NotebookAuthor struct {
	Handle string `json:"handle,omitempty"`
	Name   string `json:"name,omitempty"`
}

type NotebookCell struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // "markdown", "timeseries", "toplist", etc.
	Attributes NotebookCellAttributes `json:"attributes"`
}

type NotebookCellAttributes struct {
	Definition map[string]interface{} `json:"definition"`
}

// ExtractionResult holds the results of extracting resources from DataDog.
type ExtractionResult struct {
	Dashboards    []Dashboard
	Monitors      []Monitor
	SLOs          []SLO
	Synthetics    []SyntheticTest
	LogPipelines  []LogPipeline
	Metrics       []MetricMetadata
	Downtimes     []Downtime
	Notifications []NotificationChannel
	Notebooks     []Notebook
}
