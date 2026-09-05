// Package route implements the protoc-gen-route code generator: it turns a proto
// file's service methods carrying a (sphere.options.options) rule into a
// .<key>.pb.go file containing operation constants, extra-data lookups, and the
// server/codec scaffolding produced from the template.
package route

import (
	"errors"
	"strings"

	"github.com/go-sphere/protoc-gen-route/generate/internal/template"
	"google.golang.org/protobuf/compiler/protogen"
)

// Generator owns validated configuration and an immutable parsed template.
type Generator struct {
	cfg      *Config
	renderer *template.Renderer
}

// NewGenerator validates cfg and loads its template once for reuse across all
// files in a protoc invocation.
func NewGenerator(cfg *Config) (*Generator, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	renderer, err := template.NewRenderer(cfg.TemplateFile)
	if err != nil {
		return nil, err
	}
	return &Generator{cfg: new(*cfg), renderer: renderer}, nil
}

// GenerateFile is a convenience wrapper for generating one file. Callers that
// generate multiple files should construct a Generator and reuse it.
func GenerateFile(plugin *protogen.Plugin, file *protogen.File, cfg *Config) (*protogen.GeneratedFile, error) {
	generator, err := NewGenerator(cfg)
	if err != nil {
		return nil, err
	}
	return generator.GenerateFile(plugin, file)
}

// GenerateFile generates the .<key>.pb.go file for a single proto file. It
// returns (nil, nil) when the file has no service method carrying a matching
// options rule.
func (g *Generator) GenerateFile(plugin *protogen.Plugin, file *protogen.File) (*protogen.GeneratedFile, error) {
	if len(file.Services) == 0 || !hasOptionsRule(file.Services, g.cfg.OptionsKey) {
		return nil, nil
	}
	filename := file.GeneratedFilenamePrefix + "." + strings.ToLower(g.cfg.OptionsKey) + ".pb.go"
	generated := plugin.NewGeneratedFile(filename, file.GoImportPath)
	generateFileHeader(plugin, file, generated)
	if err := generateFileContent(file, generated, g.cfg, g.renderer); err != nil {
		return nil, err
	}
	return generated, nil
}

func generateFileContent(file *protogen.File, g *protogen.GeneratedFile, cfg *Config, renderer *template.Renderer) error {
	if len(file.Services) == 0 {
		return nil
	}
	generateGoImport(g, cfg)
	packageDesc := &template.PackageDesc{
		RequestType:  g.QualifiedGoIdent(cfg.RequestType),
		ResponseType: g.QualifiedGoIdent(cfg.ResponseType),
	}
	if cfg.ExtraType.GoName != "" {
		packageDesc.ExtraDataType = g.QualifiedGoIdent(cfg.ExtraType)
		packageDesc.NewExtraDataFunc = g.QualifiedGoIdent(cfg.ExtraConstructor)
	}
	fileCfg := &fileConfig{
		optionsKey:  cfg.OptionsKey,
		packageDesc: packageDesc,
		methodSets:  make(map[string]int),
	}
	for _, service := range file.Services {
		if err := generateService(g, service, fileCfg, renderer); err != nil {
			return err
		}
	}
	return nil
}
