package api

import (
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/converter"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

// API provides an internal API layer for the conversion pipeline.
// This is designed for future integration with a web UI.
type API struct {
	ddClient *datadog.Client
	dtClient *dynatrace.Client
	conv     *converter.Converter
}

// New creates a new API instance.
func New(ddClient *datadog.Client, dtClient *dynatrace.Client) *API {
	return &API{
		ddClient: ddClient,
		dtClient: dtClient,
		conv:     converter.New(),
	}
}

// Extract extracts all resources from DataDog.
func (a *API) Extract() (*datadog.ExtractionResult, error) {
	return a.ddClient.ExtractAll()
}

// Convert converts extracted DataDog resources to Dynatrace resources.
func (a *API) Convert(ext *datadog.ExtractionResult) (*dynatrace.ConversionResult, []error) {
	return a.conv.ConvertAll(ext)
}

// Push pushes converted resources to Dynatrace.
func (a *API) Push(result *dynatrace.ConversionResult) []error {
	return a.dtClient.PushAll(result)
}

// ValidateDataDog validates DataDog API credentials.
func (a *API) ValidateDataDog() error {
	return a.ddClient.Validate()
}

// ValidateDynatrace validates Dynatrace API credentials.
func (a *API) ValidateDynatrace() error {
	return a.dtClient.Validate()
}
