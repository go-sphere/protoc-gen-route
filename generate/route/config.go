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

	// Example GoIdent flag values, in "import/path;Ident" format, mirroring the
	// telegram bot setup documented in the README. They are not real defaults for
	// main.go (request/response models are required there) but give DefaultConfig
	// representative values so golden output matches a realistic invocation.
	exampleRequestType      = "github.com/go-sphere/sphere/social/telegram;Update"
	exampleResponseType     = "github.com/go-sphere/sphere/social/telegram;Message"
	exampleExtraType        = "github.com/go-sphere/sphere/social/telegram;MethodExtraData"
	exampleExtraConstructor = "github.com/go-sphere/sphere/social/telegram;NewMethodExtraData"
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

// genConfig holds the per-file generation state derived from Config. It is
// internal to the package and scoped to a single generated file.
type genConfig struct {
	optionsKey  string
	packageDesc *template.PackageDesc
	// methodSets tracks the per-file duplicate count for each method GoName so
	// MethodDesc.Num stays deterministic. It is scoped to a single generated file
	// (created in generateFileContent) instead of a package global, which keeps
	// output stable across files in a single protoc invocation and across tests.
	methodSets map[string]int
}

// ParseGoIdent parses a "import/path;Ident" string into a protogen.GoIdent.
func ParseGoIdent(raw string) (protogen.GoIdent, error) {
	parts := strings.Split(raw, ";")
	if len(parts) != 2 {
		return protogen.GoIdent{}, errors.New("invalid GoIdent format, expected 'path;ident'")
	}
	return protogen.GoIdent{
		GoName:       parts[1],
		GoImportPath: protogen.GoImportPath(parts[0]),
	}, nil
}

// DefaultConfig returns a Config populated with representative example values
// (the telegram bot setup from the README). main.go builds its Config from
// required flags instead; DefaultConfig exists so tests produce golden output
// that matches a realistic invocation without re-stating every field.
func DefaultConfig() *Config {
	mustIdent := func(raw string) protogen.GoIdent {
		id, err := ParseGoIdent(raw)
		if err != nil {
			panic(err)
		}
		return id
	}
	return &Config{
		OptionsKey:       DefaultOptionsKey,
		RequestType:      mustIdent(exampleRequestType),
		ResponseType:     mustIdent(exampleResponseType),
		ExtraType:        mustIdent(exampleExtraType),
		ExtraConstructor: mustIdent(exampleExtraConstructor),
	}
}
