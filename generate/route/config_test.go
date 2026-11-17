package route

import (
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
)

const (
	testRequestType      = "github.com/go-sphere/telegram-bot/telegram;Update"
	testResponseType     = "github.com/go-sphere/telegram-bot/telegram;Message"
	testExtraType        = "github.com/go-sphere/telegram-bot/telegram;MethodExtraData"
	testExtraConstructor = "github.com/go-sphere/telegram-bot/telegram;NewMethodExtraData"
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
		{name: "stdlib single segment", input: "context;Context", wantPath: "context", wantName: "Context"},
		{name: "empty", input: "", wantErr: true},
		{name: "missing separator", input: "github.com/example/project/pkg/Message", wantErr: true},
		{name: "multiple separators", input: "github.com/example/project/pkg;Message;Other", wantErr: true},
		{name: "empty import path", input: ";Message", wantErr: true},
		{name: "empty identifier", input: "github.com/example/project/pkg;", wantErr: true},
		{name: "identifier starts with digit", input: "github.com/example/project/pkg;123Message", wantErr: true},
		{name: "identifier with space", input: "github.com/example/project/pkg;Bad Name", wantErr: true},
		{name: "identifier with dot", input: "github.com/example/project/pkg;pkg.Message", wantErr: true},
		{name: "keyword identifier", input: "github.com/example/project/pkg;type", wantErr: true},
		{name: "interface literal identifier", input: "github.com/example/project/pkg;interface{}", wantErr: true},
		{name: "leading slash import path", input: "/github.com/example/pkg;Message", wantErr: true},
		{name: "trailing slash import path", input: "github.com/example/pkg/;Message", wantErr: true},
		{name: "double slash import path", input: "github.com//example/pkg;Message", wantErr: true},
		{name: "relative dot segment", input: "./pkg;Message", wantErr: true},
		{name: "parent dot segment", input: "github.com/../pkg;Message", wantErr: true},
		{name: "space in import path", input: "github.com/example/bad pkg;Message", wantErr: true},
		{name: "version suffix in import path", input: "github.com/example/pkg@v1;Message", wantErr: true},
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
		{name: "underscore options key", mutate: func(c *Config) { c.OptionsKey = "callback_query" }},
		{name: "hyphen options key", mutate: func(c *Config) { c.OptionsKey = "bot-command" }},
		{name: "digit prefix options key", mutate: func(c *Config) { c.OptionsKey = "2fa" }},
		{name: "missing options key", mutate: func(c *Config) { c.OptionsKey = "" }, wantErr: true},
		{name: "options key with slash", mutate: func(c *Config) { c.OptionsKey = "bot/cmd" }, wantErr: true},
		{name: "options key with backslash", mutate: func(c *Config) { c.OptionsKey = `bot\cmd` }, wantErr: true},
		{name: "options key with dot", mutate: func(c *Config) { c.OptionsKey = "bot.cmd" }, wantErr: true},
		{name: "options key with space", mutate: func(c *Config) { c.OptionsKey = "bot cmd" }, wantErr: true},
		{name: "options key with tab", mutate: func(c *Config) { c.OptionsKey = "bot\tcmd" }, wantErr: true},
		{name: "options key punctuation only", mutate: func(c *Config) { c.OptionsKey = "___" }, wantErr: true},
		{name: "options key invalid identifier char", mutate: func(c *Config) { c.OptionsKey = "bot:cmd" }, wantErr: true},
		{name: "missing request", mutate: func(c *Config) { c.RequestType = protogen.GoIdent{} }, wantErr: true},
		{name: "invalid request ident", mutate: func(c *Config) { c.RequestType.GoName = "123Bad" }, wantErr: true},
		{name: "invalid request path", mutate: func(c *Config) { c.RequestType.GoImportPath = "bad path" }, wantErr: true},
		{name: "missing response", mutate: func(c *Config) { c.ResponseType = protogen.GoIdent{} }, wantErr: true},
		{name: "invalid response ident", mutate: func(c *Config) { c.ResponseType.GoName = "bad ident" }, wantErr: true},
		{name: "extra type without constructor", mutate: func(c *Config) { c.ExtraConstructor = protogen.GoIdent{} }, wantErr: true},
		{name: "constructor without extra type", mutate: func(c *Config) { c.ExtraType = protogen.GoIdent{} }, wantErr: true},
		{name: "invalid extra type ident", mutate: func(c *Config) { c.ExtraType.GoName = "bad.Name" }, wantErr: true},
		{name: "invalid extra constructor path", mutate: func(c *Config) { c.ExtraConstructor.GoImportPath = "./bad" }, wantErr: true},
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
