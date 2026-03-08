package config

import (
	"testing"
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

func TestValidateDynatraceMissingToken(t *testing.T) {
	cfg := &Config{Dynatrace: DynatraceConfig{EnvURL: "https://test.dynatrace.com"}}
	if err := cfg.ValidateDynatrace(); err == nil {
		t.Error("expected error for missing token")
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
