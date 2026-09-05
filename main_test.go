package main

import (
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/go-sphere/protoc-gen-route/generate/route"
	"google.golang.org/protobuf/compiler/protogen"
)

func TestExtractConfig(t *testing.T) {
	t.Cleanup(resetConfigFlags)
	tests := []struct {
		name             string
		request          string
		response         string
		extra            string
		extraConstructor string
		key              string
		want             *route.Config
		wantErr          string
	}{
		{name: "missing request", key: "route", wantErr: "request_model is required (format: 'import/path;ModelName')"},
		{name: "missing response", request: "net/http;Request", key: "route", wantErr: "response_model is required (format: 'import/path;ModelName')"},
		{name: "empty options key", request: "net/http;Request", response: "net/http;Response", wantErr: "options_key is required"},
		{
			name: "constructor without extra model", request: "net/http;Request", response: "net/http;Response",
			extraConstructor: "example.com/data;NewData", key: "route",
			wantErr: "extra_data_model is required when extra_data_constructor is specified",
		},
		{
			name: "extra model without constructor", request: "net/http;Request", response: "net/http;Response",
			extra: "example.com/data;Data", key: "route",
			wantErr: "extra_data_constructor is required when extra_data_model is specified",
		},
		{
			name: "minimal", request: "net/http;Request", response: "net/http;Response", key: "route",
			want: &route.Config{
				OptionsKey:   "route",
				RequestType:  mustGoIdent(t, "net/http;Request"),
				ResponseType: mustGoIdent(t, "net/http;Response"),
			},
		},
		{
			name: "with extra data", request: "net/http;Request", response: "net/http;Response",
			extra: "example.com/data;Data", extraConstructor: "example.com/data;NewData", key: "bot",
			want: &route.Config{
				OptionsKey:       "bot",
				RequestType:      mustGoIdent(t, "net/http;Request"),
				ResponseType:     mustGoIdent(t, "net/http;Response"),
				ExtraType:        mustGoIdent(t, "example.com/data;Data"),
				ExtraConstructor: mustGoIdent(t, "example.com/data;NewData"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			*requestModel = tt.request
			*responseModel = tt.response
			*extraDataModel = tt.extra
			*extraDataConstructor = tt.extraConstructor
			*optionsKey = tt.key
			*templateFile = ""

			got, err := extractConfig()
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("extractConfig() error = nil, want %q", tt.wantErr)
				}
				if gotErr := err.Error(); gotErr != tt.wantErr {
					t.Errorf("extractConfig() error = %q, want %q", gotErr, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("extractConfig() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractConfig() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestVersionFlag(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "protoc-gen-route")
	build := exec.CommandContext(t.Context(), "go", "build", "-o", binPath, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build plugin: %v\n%s", err, output)
	}

	command := exec.CommandContext(t.Context(), binPath, "-version")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("run -version: %v\n%s", err, output)
	}
	if got, want := string(output), "protoc-gen-route "+version+"\n"; got != want {
		t.Errorf("version output = %q, want %q", got, want)
	}
}

func mustGoIdent(t *testing.T, raw string) protogen.GoIdent {
	t.Helper()
	ident, err := route.ParseGoIdent(raw)
	if err != nil {
		t.Fatalf("ParseGoIdent(%q): %v", raw, err)
	}
	return ident
}

func resetConfigFlags() {
	*requestModel = ""
	*responseModel = ""
	*extraDataModel = ""
	*extraDataConstructor = ""
	*optionsKey = route.DefaultOptionsKey
	*templateFile = ""
}
