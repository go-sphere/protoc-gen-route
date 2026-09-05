package route

import (
	"errors"
	"strings"

	"github.com/go-sphere/protoc-gen-route/generate/internal/template"
	"google.golang.org/protobuf/compiler/protogen"
)

const (
	// DefaultOptionsKey is the default key looked up in the
	// (sphere.options.options) extension. Only methods carrying a rule with this
	// key are turned into route code.
	DefaultOptionsKey = "route"
)

// Config holds the user-facing options for the route generator. It is populated
// from command-line flags in main.go and passed to GenerateFile.
type Config struct {
	OptionsKey   string
	TemplateFile string

	RequestType      protogen.GoIdent
	ResponseType     protogen.GoIdent
	ExtraType        protogen.GoIdent
	ExtraConstructor protogen.GoIdent
}

// fileConfig holds the per-file generation state derived from Config. It is
// internal to the package and scoped to a single generated file.
type fileConfig struct {
	optionsKey  string
	packageDesc *template.PackageDesc
	// methodSets tracks the per-file duplicate count for each method GoName so
	// MethodDesc.Num stays deterministic. It is scoped to a single generated file
	// (created in generateFileContent) instead of a package global, which keeps
	// output stable across files in a single protoc invocation and across tests.
	methodSets map[string]int
}

// ParseGoIdent parses an "import/path;Ident" string into a protogen.GoIdent.
func ParseGoIdent(raw string) (protogen.GoIdent, error) {
	importPath, goName, ok := strings.Cut(raw, ";")
	if !ok || importPath == "" || goName == "" || strings.Contains(goName, ";") {
		return protogen.GoIdent{}, errors.New("invalid GoIdent format, expected 'import/path;Ident'")
	}
	return protogen.GoIdent{
		GoName:       goName,
		GoImportPath: protogen.GoImportPath(importPath),
	}, nil
}

// Validate checks required identifiers and the optional extra-data pair.
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("config is required")
	}
	if c.OptionsKey == "" {
		return errors.New("options_key is required")
	}
	if c.RequestType.GoImportPath == "" || c.RequestType.GoName == "" {
		return errors.New("request_model is required (format: 'import/path;ModelName')")
	}
	if c.ResponseType.GoImportPath == "" || c.ResponseType.GoName == "" {
		return errors.New("response_model is required (format: 'import/path;ModelName')")
	}
	hasExtraType := c.ExtraType.GoImportPath != "" || c.ExtraType.GoName != ""
	hasExtraConstructor := c.ExtraConstructor.GoImportPath != "" || c.ExtraConstructor.GoName != ""
	if hasExtraType && (c.ExtraType.GoImportPath == "" || c.ExtraType.GoName == "") {
		return errors.New("extra_data_model is required (format: 'import/path;ModelName')")
	}
	if hasExtraConstructor && (c.ExtraConstructor.GoImportPath == "" || c.ExtraConstructor.GoName == "") {
		return errors.New("extra_data_constructor is required (format: 'import/path;Ident')")
	}
	if hasExtraConstructor && !hasExtraType {
		return errors.New("extra_data_model is required when extra_data_constructor is specified")
	}
	if hasExtraType && !hasExtraConstructor {
		return errors.New("extra_data_constructor is required when extra_data_model is specified")
	}
	return nil
}

// DefaultConfig returns the plugin's real defaults. Required request and
// response models remain unset until supplied by the caller.
func DefaultConfig() *Config {
	return &Config{
		OptionsKey: DefaultOptionsKey,
	}
}
