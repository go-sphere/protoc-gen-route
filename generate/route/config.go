package route

import (
	"errors"
	"fmt"
	"go/token"
	"strings"
	"unicode"

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

	// PluginVersion is recorded in the generated file header
	// ("// - protoc-gen-route <PluginVersion>"). main.go sets it from the
	// binary's version; an empty value renders as "(unknown)". It does not
	// affect anything but the header.
	PluginVersion string
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
	ident := protogen.GoIdent{
		GoName:       goName,
		GoImportPath: protogen.GoImportPath(importPath),
	}
	if err := validateGoIdent(ident); err != nil {
		return protogen.GoIdent{}, err
	}
	return ident, nil
}

// Validate checks required identifiers, options_key, and the optional extra-data pair.
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("config is required")
	}
	if err := validateOptionsKey(c.OptionsKey); err != nil {
		return err
	}
	if c.RequestType.GoImportPath == "" || c.RequestType.GoName == "" {
		return errors.New("request_model is required (format: 'import/path;ModelName')")
	}
	if err := validateGoIdent(c.RequestType); err != nil {
		return fmt.Errorf("invalid request_model: %w", err)
	}
	if c.ResponseType.GoImportPath == "" || c.ResponseType.GoName == "" {
		return errors.New("response_model is required (format: 'import/path;ModelName')")
	}
	if err := validateGoIdent(c.ResponseType); err != nil {
		return fmt.Errorf("invalid response_model: %w", err)
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
	if hasExtraType {
		if err := validateGoIdent(c.ExtraType); err != nil {
			return fmt.Errorf("invalid extra_data_model: %w", err)
		}
	}
	if hasExtraConstructor {
		if err := validateGoIdent(c.ExtraConstructor); err != nil {
			return fmt.Errorf("invalid extra_data_constructor: %w", err)
		}
	}
	return nil
}

func validateOptionsKey(key string) error {
	if key == "" {
		return errors.New("options_key is required")
	}
	for _, r := range key {
		if r == '/' || r == '\\' || r == '.' || unicode.IsSpace(r) {
			return fmt.Errorf("invalid options_key %q: must not contain '/', '\\\\', '.', or whitespace", key)
		}
	}
	suffix := pascalCase(key)
	if suffix == "" || !token.IsIdentifier("_"+suffix) {
		return fmt.Errorf("invalid options_key %q: must form a valid Go identifier suffix", key)
	}
	return nil
}

func validateGoIdent(ident protogen.GoIdent) error {
	if err := validateImportPath(string(ident.GoImportPath)); err != nil {
		return err
	}
	if !token.IsIdentifier(ident.GoName) {
		return fmt.Errorf("invalid Go identifier %q", ident.GoName)
	}
	return nil
}

func validateImportPath(importPath string) error {
	if importPath == "" {
		return errors.New("import path cannot be empty")
	}
	for seg := range strings.SplitSeq(importPath, "/") {
		if seg == "" || seg == "." || seg == ".." {
			return fmt.Errorf("invalid segment %q in import path %q", seg, importPath)
		}
	}
	for _, r := range importPath {
		if !unicode.IsGraphic(r) || unicode.IsSpace(r) || r == '\ufffd' || strings.ContainsRune("!\"#$%&'()*,:;<=>?@[\\]^`{|}", r) {
			return fmt.Errorf("illegal character %q in import path %q", r, importPath)
		}
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
