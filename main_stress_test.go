package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-sphere/protoc-gen-route/generate/route"
)

// TestExtractConfig_AdversarialInputs stress-tests extractConfig with empty strings,
// malformed GoIdent formats, and mismatched extra_data flags.
func TestExtractConfig_AdversarialInputs(t *testing.T) {
	reset := func() {
		*requestModel = ""
		*responseModel = ""
		*extraDataModel = ""
		*extraDataConstructor = ""
		*optionsKey = "route"
		*templateFile = ""
	}

	t.Run("empty string request_model", func(t *testing.T) {
		reset()
		*requestModel = ""
		*responseModel = "net/http;Response"
		_, err := extractConfig()
		if err == nil || !strings.Contains(err.Error(), "request_model is required") {
			t.Fatalf("expected request_model required error, got %v", err)
		}
	})

	t.Run("empty string response_model", func(t *testing.T) {
		reset()
		*requestModel = "net/http;Request"
		*responseModel = ""
		_, err := extractConfig()
		if err == nil || !strings.Contains(err.Error(), "response_model is required") {
			t.Fatalf("expected response_model required error, got %v", err)
		}
	})

	t.Run("malformed request_model no semicolon", func(t *testing.T) {
		reset()
		*requestModel = "net/http/Request"
		*responseModel = "net/http;Response"
		_, err := extractConfig()
		if err == nil || !strings.Contains(err.Error(), "invalid GoIdent format") {
			t.Fatalf("expected invalid GoIdent format error, got %v", err)
		}
	})

	t.Run("malformed request_model multiple semicolons", func(t *testing.T) {
		reset()
		*requestModel = "net/http;Request;Extra"
		*responseModel = "net/http;Response"
		_, err := extractConfig()
		if err == nil || !strings.Contains(err.Error(), "invalid GoIdent format") {
			t.Fatalf("expected invalid GoIdent format error, got %v", err)
		}
	})

	t.Run("malformed response_model no semicolon", func(t *testing.T) {
		reset()
		*requestModel = "net/http;Request"
		*responseModel = "net/http/Response"
		_, err := extractConfig()
		if err == nil || !strings.Contains(err.Error(), "invalid GoIdent format") {
			t.Fatalf("expected invalid GoIdent format error, got %v", err)
		}
	})

	t.Run("malformed response_model multiple semicolons", func(t *testing.T) {
		reset()
		*requestModel = "net/http;Request"
		*responseModel = "net/http;Response;Extra"
		_, err := extractConfig()
		if err == nil || !strings.Contains(err.Error(), "invalid GoIdent format") {
			t.Fatalf("expected invalid GoIdent format error, got %v", err)
		}
	})

	t.Run("mismatched extra_data: constructor without model", func(t *testing.T) {
		reset()
		*requestModel = "net/http;Request"
		*responseModel = "net/http;Response"
		*extraDataModel = ""
		*extraDataConstructor = "pkg/data;NewData"
		_, err := extractConfig()
		if err == nil || !strings.Contains(err.Error(), "extra_data_model is required") {
			t.Fatalf("expected extra_data_model is required error, got %v", err)
		}
	})

	t.Run("mismatched extra_data: model without constructor", func(t *testing.T) {
		reset()
		*requestModel = "net/http;Request"
		*responseModel = "net/http;Response"
		*extraDataModel = "pkg/data;Data"
		*extraDataConstructor = ""
		_, err := extractConfig()
		if err == nil || !strings.Contains(err.Error(), "extra_data_constructor is required") {
			t.Fatalf("expected extra_data_constructor is required error, got %v", err)
		}
	})

	t.Run("mismatched extra_data: malformed model with valid constructor", func(t *testing.T) {
		reset()
		*requestModel = "net/http;Request"
		*responseModel = "net/http;Response"
		*extraDataModel = "invalid_model_ident"
		*extraDataConstructor = "pkg/data;NewData"
		_, err := extractConfig()
		if err == nil || !strings.Contains(err.Error(), "invalid GoIdent format") {
			t.Fatalf("expected invalid GoIdent format for model, got %v", err)
		}
	})

	t.Run("mismatched extra_data: valid model with malformed constructor", func(t *testing.T) {
		reset()
		*requestModel = "net/http;Request"
		*responseModel = "net/http;Response"
		*extraDataModel = "pkg/data;Data"
		*extraDataConstructor = "invalid_constructor_ident"
		_, err := extractConfig()
		if err == nil || !strings.Contains(err.Error(), "invalid GoIdent format") {
			t.Fatalf("expected invalid GoIdent format for constructor, got %v", err)
		}
	})

	t.Run("empty options_key preserves empty string", func(t *testing.T) {
		reset()
		*requestModel = "net/http;Request"
		*responseModel = "net/http;Response"
		*optionsKey = ""
		cfg, err := extractConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.OptionsKey != "" {
			t.Errorf("expected empty options_key, got %q", cfg.OptionsKey)
		}
	})
}

// TestParseGoIdent_AdversarialMatrix tests ParseGoIdent edge cases.
func TestParseGoIdent_AdversarialMatrix(t *testing.T) {
	cases := []struct {
		input   string
		wantErr bool
		expPkg  string
		expName string
	}{
		{"", true, "", ""},
		{"single_token", true, "", ""},
		{"a;b;c", true, "", ""},
		{";;;", true, "", ""},
		{"a;b", false, "a", "b"},
		{";b", false, "", "b"},
		{"a;", false, "a", ""},
		{";", false, "", ""},
		{"github.com/foo/bar;Baz", false, "github.com/foo/bar", "Baz"},
	}

	for _, tc := range cases {
		t.Run("ident_"+tc.input, func(t *testing.T) {
			id, err := route.ParseGoIdent(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("input %q: got err %v, wantErr %v", tc.input, err, tc.wantErr)
			}
			if !tc.wantErr {
				if string(id.GoImportPath) != tc.expPkg || id.GoName != tc.expName {
					t.Errorf("input %q: got (%q, %q), want (%q, %q)",
						tc.input, id.GoImportPath, id.GoName, tc.expPkg, tc.expName)
				}
			}
		})
	}
}

// TestCLI_CommandInvocations builds and executes the protoc-gen-route binary
// directly with various command-line arguments and input streams.
func TestCLI_CommandInvocations(t *testing.T) {
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "protoc-gen-route")

	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build protoc-gen-route binary: %v, output: %s", err, string(out))
	}

	// 1. -version flag
	t.Run("version flag", func(t *testing.T) {
		cmd := exec.Command(binPath, "-version")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("-version failed: %v, output: %s", err, string(out))
		}
		if !strings.Contains(string(out), "protoc-gen-route") {
			t.Errorf("version output unexpected: %s", string(out))
		}
	})

	// 2. Empty stdin invocation (protogen expects CodeGeneratorRequest)
	t.Run("empty stdin without flags", func(t *testing.T) {
		cmd := exec.Command(binPath)
		cmd.Stdin = bytes.NewReader([]byte{})
		out, _ := cmd.CombinedOutput()
		// Should exit cleanly or with an unmarshal error, but must never crash/panic
		if strings.Contains(string(out), "panic:") {
			t.Fatalf("unexpected panic on empty stdin: %s", string(out))
		}
	})

	// 3. Garbage / malformed stdin
	t.Run("malformed stdin", func(t *testing.T) {
		cmd := exec.Command(binPath)
		cmd.Stdin = bytes.NewReader([]byte("not a valid protobuf stream \x00\xff\xfe"))
		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Logf("completed: %s", string(out))
		}
		if strings.Contains(string(out), "panic:") {
			t.Fatalf("binary panicked on malformed stdin: %s", string(out))
		}
	})
}
