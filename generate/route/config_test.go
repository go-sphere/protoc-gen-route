package route

import (
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
)

const (
	testRequestType      = "github.com/go-sphere/sphere/social/telegram;Update"
	testResponseType     = "github.com/go-sphere/sphere/social/telegram;Message"
	testExtraType        = "github.com/go-sphere/sphere/social/telegram;MethodExtraData"
	testExtraConstructor = "github.com/go-sphere/sphere/social/telegram;NewMethodExtraData"
)

func TestParseGoIdent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantPath string
		wantName string
		wantErr  bool
	}{
		{name: "valid", input: "github.com/example/project/pkg;Message", wantPath: "github.com/example/project/pkg", wantName: "Message"},
		{name: "empty", input: "", wantErr: true},
		{name: "missing separator", input: "github.com/example/project/pkg/Message", wantErr: true},
		{name: "multiple separators", input: "github.com/example/project/pkg;Message;Other", wantErr: true},
		{name: "empty import path", input: ";Message", wantErr: true},
		{name: "empty identifier", input: "github.com/example/project/pkg;", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGoIdent(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseGoIdent(%q) error = nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseGoIdent(%q) error = %v", tt.input, err)
			}
			if gotPath := string(got.GoImportPath); gotPath != tt.wantPath {
				t.Errorf("GoImportPath = %q, want %q", gotPath, tt.wantPath)
			}
			if got.GoName != tt.wantName {
				t.Errorf("GoName = %q, want %q", got.GoName, tt.wantName)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.OptionsKey != DefaultOptionsKey {
		t.Errorf("OptionsKey = %q, want %q", cfg.OptionsKey, DefaultOptionsKey)
	}
	if cfg.RequestType != (protogen.GoIdent{}) || cfg.ResponseType != (protogen.GoIdent{}) {
		t.Error("required models must not have test-only defaults")
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr bool
	}{
		{name: "valid", mutate: func(*Config) {}},
		{name: "missing options key", mutate: func(c *Config) { c.OptionsKey = "" }, wantErr: true},
		{name: "missing request", mutate: func(c *Config) { c.RequestType = protogen.GoIdent{} }, wantErr: true},
		{name: "missing response", mutate: func(c *Config) { c.ResponseType = protogen.GoIdent{} }, wantErr: true},
		{name: "extra type without constructor", mutate: func(c *Config) { c.ExtraConstructor = protogen.GoIdent{} }, wantErr: true},
		{name: "constructor without extra type", mutate: func(c *Config) { c.ExtraType = protogen.GoIdent{} }, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := testConfig()
			tt.mutate(cfg)
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	if err := (*Config)(nil).Validate(); err == nil {
		t.Fatal("nil Config.Validate() error = nil")
	}
}

func testConfig() *Config {
	cfg := DefaultConfig()
	cfg.RequestType = mustParseGoIdent(testRequestType)
	cfg.ResponseType = mustParseGoIdent(testResponseType)
	cfg.ExtraType = mustParseGoIdent(testExtraType)
	cfg.ExtraConstructor = mustParseGoIdent(testExtraConstructor)
	return cfg
}

func mustParseGoIdent(raw string) protogen.GoIdent {
	ident, err := ParseGoIdent(raw)
	if err != nil {
		panic(err)
	}
	return ident
}
