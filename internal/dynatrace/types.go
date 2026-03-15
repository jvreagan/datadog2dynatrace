package dynatrace

// Dashboard represents a Dynatrace dashboard.
type Dashboard struct {
	ID                string            `json:"id,omitempty"`
	DashboardMetadata DashboardMetadata `json:"dashboardMetadata"`
	Tiles             []Tile            `json:"tiles"`
}

type DashboardMetadata struct {
	Name   string   `json:"name"`
	Owner  string   `json:"owner,omitempty"`
	Tags   []string `json:"tags,omitempty"`
	Preset bool     `json:"preset,omitempty"`
}

type Tile struct {
	Name       string      `json:"name"`
	TileType   string      `json:"tileType"`
	Configured bool        `json:"configured"`
	Bounds     TileBounds  `json:"bounds"`
	// For custom charting tiles
	FilterConfig *TileFilterConfig `json:"filterConfig,omitempty"`
	// For markdown tiles
	Markdown string `json:"markdown,omitempty"`
	// For data explorer tiles
	Queries []DashboardQuery `json:"queries,omitempty"`
	// For SLO tiles
	AssignedEntities []string `json:"assignedEntities,omitempty"`
}

type TileBounds struct {
	Top    int `json:"top"`
	Left   int `json:"left"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type TileFilterConfig struct {
	Type                 string                `json:"type"`
	CustomName           string                `json:"customName,omitempty"`
	ChartConfig          *ChartConfig          `json:"chartConfig,omitempty"`
	FiltersPerEntityType map[string]interface{} `json:"filtersPerEntityType,omitempty"`
}

type ChartConfig struct {
	Type           string        `json:"type,omitempty"`
	Series         []ChartSeries `json:"series"`
	LeftAxisCustomUnit string    `json:"leftAxisCustomUnit,omitempty"`
}

type ChartSeries struct {
	Metric      string            `json:"metric"`
	Aggregation string            `json:"aggregation"`
	Type        string            `json:"type,omitempty"`
	EntityType  string            `json:"entityType,omitempty"`
	Dimensions  []SeriesDimension `json:"dimensions,omitempty"`
}

type SeriesDimension struct {
	ID     string   `json:"id"`
	Name   string   `json:"name,omitempty"`
	Values []string `json:"values,omitempty"`
}

type DashboardQuery struct {
	ID             string `json:"id"`
	MetricSelector string `json:"metricSelector,omitempty"`
	DQL            string `json:"dql,omitempty"`
	SpaceAggregation string `json:"spaceAggregation,omitempty"`
	TimeAggregation  string `json:"timeAggregation,omitempty"`
}

// MetricEvent represents a Dynatrace metric event (alert).
type MetricEvent struct {
	ID                   string             `json:"id,omitempty"`
	Summary              string             `json:"summary"`
	Description          string             `json:"description,omitempty"`
	MetricSelector       string             `json:"metricSelector"`
	Enabled              bool               `json:"enabled"`
	EventType            string             `json:"eventType"` // "CUSTOM_ALERT", "ERROR", "INFO"
	ModelType            string             `json:"modelType"` // "STATIC_THRESHOLD", "AUTO_ADAPTIVE"
	MonitoringStrategy   MonitoringStrategy `json:"monitoringStrategy"`
	AlertCondition       string             `json:"alertCondition"` // "ABOVE", "BELOW"
	Threshold            float64            `json:"threshold,omitempty"`
	DealertingSamples    int                `json:"dealertingSamples,omitempty"`
	Samples              int                `json:"samples,omitempty"`
	ViolatingSamples     int                `json:"violatingSamples,omitempty"`
	Tags                 []METag            `json:"queryFilterTags,omitempty"`
}

type MonitoringStrategy struct {
	Type              string  `json:"type"`
	Samples           int     `json:"samples,omitempty"`
	ViolatingSamples  int     `json:"violatingSamples,omitempty"`
	DealertingSamples int     `json:"dealertingSamples,omitempty"`
	AlertCondition    string  `json:"alertCondition,omitempty"`
	Threshold         float64 `json:"threshold,omitempty"`
}

type METag struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

// SLO represents a Dynatrace SLO.
type SLO struct {
	ID             string  `json:"id,omitempty"`
	Name           string  `json:"name"`
	Description    string  `json:"description,omitempty"`
	MetricName     string  `json:"metricName,omitempty"`
	MetricExpression string `json:"metricExpression,omitempty"`
	EvaluationType string  `json:"evaluationType"` // "AGGREGATE"
	Filter         string  `json:"filter,omitempty"`
	Target         float64 `json:"target"`
	Warning        float64 `json:"warning"`
	Timeframe      string  `json:"timeframe"` // "-1d", "-1w", "-1M"
	Enabled        bool    `json:"enabled"`
}

// SyntheticMonitor represents a Dynatrace synthetic monitor.
type SyntheticMonitor struct {
	EntityID            string               `json:"entityId,omitempty"`
	Name                string               `json:"name"`
	Type                string               `json:"type"` // "HTTP", "BROWSER"
	Enabled             bool                 `json:"enabled"`
	FrequencyMin        int                  `json:"frequencyMin"`
	Locations           []string             `json:"locations"`
	Tags                []SyntheticTag       `json:"tags,omitempty"`
	AnomalyDetection    *AnomalyDetection    `json:"anomalyDetection,omitempty"`
	// For HTTP monitors
	Script              *SyntheticScript     `json:"script,omitempty"`
	// For browser monitors
	KeyPerformanceMetrics *KeyPerformanceMetrics `json:"keyPerformanceMetrics,omitempty"`
}

type SyntheticTag struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

type AnomalyDetection struct {
	OutageHandling *OutageHandling `json:"outageHandling,omitempty"`
	LoadingTimeThresholds *LoadingTimeThresholds `json:"loadingTimeThresholds,omitempty"`
}

type OutageHandling struct {
	GlobalOutage      bool `json:"globalOutage"`
	LocalOutage       bool `json:"localOutage"`
	LocalOutagePolicy *LocalOutagePolicy `json:"localOutagePolicy,omitempty"`
	RetryOnError      bool `json:"retryOnError"`
}

type LocalOutagePolicy struct {
	AffectedLocations int `json:"affectedLocations"`
	ConsecutiveRuns   int `json:"consecutiveRuns"`
}

type LoadingTimeThresholds struct {
	Enabled    bool                `json:"enabled"`
	Thresholds []LoadingThreshold  `json:"thresholds,omitempty"`
}

type LoadingThreshold struct {
	Type         string `json:"type"`
	ValueMs      int    `json:"valueMs"`
}

type SyntheticScript struct {
	Version  string          `json:"version,omitempty"`
	Type     string          `json:"type,omitempty"`
	Requests []ScriptRequest `json:"requests,omitempty"`
	// For browser monitors
	Events []ScriptEvent `json:"events,omitempty"`
}

type ScriptRequest struct {
	Description    string                 `json:"description"`
	URL            string                 `json:"url"`
	Method         string                 `json:"method"`
	RequestBody    string                 `json:"requestBody,omitempty"`
	Configuration  *RequestConfiguration  `json:"configuration,omitempty"`
	Validation     *RequestValidation     `json:"validation,omitempty"`
}

type RequestConfiguration struct {
	AcceptAnyCertificate bool              `json:"acceptAnyCertificate,omitempty"`
	FollowRedirects      bool              `json:"followRedirects,omitempty"`
	Headers              map[string]string `json:"headers,omitempty"`
}

type RequestValidation struct {
	Rules []ValidationRule `json:"rules,omitempty"`
}

type ValidationRule struct {
	Type       string `json:"type"`
	Value      string `json:"value"`
	PassIfFound bool  `json:"passIfFound"`
}

type ScriptEvent struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	URL         string                 `json:"url,omitempty"`
	Target      *EventTarget           `json:"target,omitempty"`
	Wait        *EventWait             `json:"wait,omitempty"`
}

type EventTarget struct {
	Window   string `json:"window,omitempty"`
	Locators []Locator `json:"locators,omitempty"`
}

type Locator struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type EventWait struct {
	WaitFor      string `json:"waitFor,omitempty"`
	Milliseconds int    `json:"milliseconds,omitempty"`
}

type KeyPerformanceMetrics struct {
	LoadActionKPM string `json:"loadActionKpm"`
	XHRActionKPM  string `json:"xhrActionKpm"`
}

// LogProcessingRule represents a Dynatrace log processing rule.
type LogProcessingRule struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name"`
	Enabled  bool   `json:"enabled"`
	Query    string `json:"query"`
	Processor string `json:"processor"` // DQL processor definition
}

// MetricDescriptor represents Dynatrace metric metadata.
type MetricDescriptor struct {
	MetricID    string   `json:"metricId"`
	DisplayName string   `json:"displayName,omitempty"`
	Description string   `json:"description,omitempty"`
	Unit        string   `json:"unit,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// MaintenanceWindow represents a Dynatrace maintenance window.
type MaintenanceWindow struct {
	ID          string             `json:"id,omitempty"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Type        string             `json:"type"` // "PLANNED", "UNPLANNED"
	Suppression string             `json:"suppression"` // "DETECT_PROBLEMS_AND_ALERT", "DETECT_PROBLEMS_DONT_ALERT", "DONT_DETECT_PROBLEMS"
	Schedule    MaintenanceSchedule `json:"schedule"`
	Scope       *MaintenanceScope  `json:"scope,omitempty"`
}

type MaintenanceSchedule struct {
	RecurrenceType string `json:"recurrenceType"` // "ONCE", "DAILY", "WEEKLY", "MONTHLY"
	Start          string `json:"start"`
	End            string `json:"end"`
	ZoneID         string `json:"zoneId"`
	Recurrence     *MaintenanceRecurrence `json:"recurrence,omitempty"`
}

type MaintenanceRecurrence struct {
	DayOfWeek  string `json:"dayOfWeek,omitempty"`
	DayOfMonth int    `json:"dayOfMonth,omitempty"`
	DurationMinutes int `json:"durationMinutes"`
	StartTime  string `json:"startTime"` // "HH:mm"
}

type MaintenanceScope struct {
	Entities []string                `json:"entities,omitempty"`
	Matches  []MaintenanceScopeMatch `json:"matches,omitempty"`
}

type MaintenanceScopeMatch struct {
	Type       string            `json:"type"`
	MzID       string            `json:"mzId,omitempty"`
	Tags       []METag           `json:"tags,omitempty"`
	TagCombination string        `json:"tagCombination,omitempty"` // "AND", "OR"
}

// NotificationIntegration represents a Dynatrace alerting profile / notification.
type NotificationIntegration struct {
	ID       string                 `json:"id,omitempty"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"` // "SLACK", "PAGER_DUTY", "EMAIL", "WEBHOOK"
	Active   bool                   `json:"active"`
	Config   map[string]interface{} `json:"config"`
}

// DynatraceNotebook represents a Dynatrace notebook.
type DynatraceNotebook struct {
	ID       string           `json:"id,omitempty"`
	Name     string           `json:"name"`
	Sections []NotebookSection `json:"sections"`
}

type NotebookSection struct {
	ID          string `json:"id,omitempty"`
	Type        string `json:"type"` // "markdown", "code", "chart"
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Content     string `json:"content,omitempty"`
	// For code/DQL sections
	Query string `json:"query,omitempty"`
	// For chart sections
	Visualization string `json:"visualization,omitempty"`
}

// ConversionResult holds the output of converting DD resources to DT.
type ConversionResult struct {
	Dashboards     []Dashboard
	MetricEvents   []MetricEvent
	SLOs           []SLO
	Synthetics     []SyntheticMonitor
	LogRules       []LogProcessingRule
	Metrics        []MetricDescriptor
	Maintenance    []MaintenanceWindow
	Notifications  []NotificationIntegration
	Notebooks      []DynatraceNotebook
}

// DocumentRequest is the payload for the Gen3 Documents API.
type DocumentRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	IsPrivate bool `json:"isPrivate"`
}

// SettingsObjectCreate is the wrapper for Settings 2.0 API objects.
type SettingsObjectCreate struct {
	SchemaID      string      `json:"schemaId"`
	SchemaVersion string      `json:"schemaVersion,omitempty"`
	Scope         string      `json:"scope"`
	Value         interface{} `json:"value"`
}

// DavisAnomalyDetector maps to the builtin:davis.anomaly-detectors value schema.
type DavisAnomalyDetector struct {
	Title           string                   `json:"title"`
	Description     string                   `json:"description,omitempty"`
	Enabled         bool                     `json:"enabled"`
	EventTemplate   DavisEventTemplate       `json:"eventTemplate"`
	Analyzer        DavisAnalyzerConfig      `json:"analyzer"`
	QueryDefinition DavisQueryDefinition     `json:"queryDefinition"`
}

// DavisEventTemplate defines the event properties for an anomaly detector.
type DavisEventTemplate struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	EventType   string `json:"eventType"`
	DavisMerge  bool   `json:"davisMerge"`
}

// DavisAnalyzerConfig defines the analyzer for an anomaly detector.
type DavisAnalyzerConfig struct {
	Input []DavisAnalyzerInput `json:"input"`
}

// DavisAnalyzerInput defines input for the analyzer.
type DavisAnalyzerInput struct {
	Name        string                    `json:"name"`
	AnalyzerDef DavisAnalyzerDefinition   `json:"analyzerDef,omitempty"`
}

// DavisAnalyzerDefinition specifies the threshold type and parameters.
type DavisAnalyzerDefinition struct {
	Type                string  `json:"type"`
	Threshold           float64 `json:"threshold,omitempty"`
	AlertCondition      string  `json:"alertCondition,omitempty"`
	Samples             int     `json:"samples,omitempty"`
	ViolatingSamples    int     `json:"violatingSamples,omitempty"`
	DealertingSamples   int     `json:"dealertingSamples,omitempty"`
}

// DavisQueryDefinition defines the metric query for an anomaly detector.
type DavisQueryDefinition struct {
	Type           string `json:"type"`
	MetricSelector string `json:"metricSelector,omitempty"`
}

// MaintenanceWindowSetting maps to builtin:alerting.maintenance-window value schema.
type MaintenanceWindowSetting struct {
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	Enabled         bool   `json:"enabled"`
	Type            string `json:"type"`
	Suppression     string `json:"suppression"`
	GeneralProperties MaintenanceGeneralProperties `json:"generalProperties"`
}

// MaintenanceGeneralProperties holds schedule info for a maintenance window setting.
type MaintenanceGeneralProperties struct {
	DisableSyntheticMonitoring bool   `json:"disableSyntheticMonitoring"`
	MaintenanceType            string `json:"maintenanceType"`
	Recurrence                 string `json:"recurrence"`
	StartTime                  string `json:"startTime,omitempty"`
	EndTime                    string `json:"endTime,omitempty"`
	ZoneID                     string `json:"zoneId,omitempty"`
}

// NotificationSetting maps to builtin:problem.notifications value schema.
type NotificationSetting struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Active  bool   `json:"active"`
	Config  map[string]interface{} `json:"config"`
}
