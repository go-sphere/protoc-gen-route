package main

import (
	"strings"
	"testing"
)

func TestExtractConfig_RequiredFlags(t *testing.T) {
	// Reset flags before each test
	reset := func() {
		*requestModel = ""
		*responseModel = ""
		*extraDataModel = ""
		*extraDataConstructor = ""
	}

	t.Run("missing request_model", func(t *testing.T) {
		reset()
		_, err := extractConfig()
		if err == nil || !strings.Contains(err.Error(), "request_model is required") {
			t.Fatalf("expected request_model is required error, got: %v", err)
		}
	})

	t.Run("missing response_model", func(t *testing.T) {
		reset()
		*requestModel = "net/http;Request"
		_, err := extractConfig()
		if err == nil || !strings.Contains(err.Error(), "response_model is required") {
			t.Fatalf("expected response_model is required error, got: %v", err)
		}
	})

	t.Run("valid minimal config", func(t *testing.T) {
		reset()
		*requestModel = "net/http;Request"
		*responseModel = "net/http;Response"
		cfg, err := extractConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.RequestType.GoName != "Request" || cfg.ResponseType.GoName != "Response" {
			t.Fatalf("unexpected config models: %+v", cfg)
		}
	})

	t.Run("extra_data_constructor without model", func(t *testing.T) {
		reset()
		*requestModel = "net/http;Request"
		*responseModel = "net/http;Response"
		*extraDataConstructor = "pkg/data;NewData"
		_, err := extractConfig()
		if err == nil || !strings.Contains(err.Error(), "extra_data_model is required") {
			t.Fatalf("expected extra_data_model is required error, got: %v", err)
		}
	})

	t.Run("extra_data_model without constructor", func(t *testing.T) {
		reset()
		*requestModel = "net/http;Request"
		*responseModel = "net/http;Response"
		*extraDataModel = "pkg/data;Data"
		_, err := extractConfig()
		if err == nil || !strings.Contains(err.Error(), "extra_data_constructor is required") {
			t.Fatalf("expected extra_data_constructor is required error, got: %v", err)
		}
	})

	t.Run("valid full config", func(t *testing.T) {
		reset()
		*requestModel = "net/http;Request"
		*responseModel = "net/http;Response"
		*extraDataModel = "pkg/data;Data"
		*extraDataConstructor = "pkg/data;NewData"
		cfg, err := extractConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ExtraType.GoName != "Data" || cfg.ExtraConstructor.GoName != "NewData" {
			t.Fatalf("unexpected extra config: %+v", cfg)
		}
	})
}
