// Package route implements the protoc-gen-route code generator: it turns a proto
// file's service methods carrying a (sphere.options.options) rule into a
// .<key>.pb.go file containing operation constants, extra-data lookups, and the
// server/codec scaffolding produced from the template.
package route

import (
	"fmt"
	"strings"

	"github.com/go-sphere/options/sphere/options"
	"github.com/go-sphere/protoc-gen-route/generate/internal/template"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	deprecationComment = "// Deprecated: Do not use."
)

const (
	contextPackage = protogen.GoImportPath("context")
)

// ReplaceTemplateIfNeed overrides the built-in code template with the file at
// path when path is non-empty. It must be called once before GenerateFile. It is
// a thin wrapper over the internal template package so that callers (e.g. main)
// need not import that internal package directly.
func ReplaceTemplateIfNeed(path string) error {
	return template.ReplaceTemplateIfNeed(path)
}

// GenerateFile generates the .<key>.pb.go file for a single proto file. It
// returns (nil, nil) when the file has no service method carrying a matching
// options rule.
func GenerateFile(gen *protogen.Plugin, file *protogen.File, conf *Config) (*protogen.GeneratedFile, error) {
	if len(file.Services) == 0 || !hasOptionsRule(file.Services, conf.OptionsKey) {
		return nil, nil
	}
	filename := file.GeneratedFilenamePrefix + fmt.Sprintf(".%s.pb.go", strings.ToLower(conf.OptionsKey))
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	generateFileHeader(gen, file, g)
	err := generateFileContent(file, g, conf)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func generateFileHeader(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	lines := formatFileHeader(
		formatProtocVersion(gen.Request.GetCompilerVersion()),
		file.Desc.Path(),
		string(file.GoPackageName),
		file.Proto.GetOptions().GetDeprecated(),
	)
	for _, line := range lines {
		g.P(line)
	}
}

func generateFileContent(file *protogen.File, g *protogen.GeneratedFile, conf *Config) error {
	if len(file.Services) == 0 {
		return nil
	}
	generateGoImport(g, conf)
	packageDesc := &template.PackageDesc{
		RequestType:  g.QualifiedGoIdent(conf.RequestType),
		ResponseType: g.QualifiedGoIdent(conf.ResponseType),
	}
	if conf.ExtraType.GoName != "" {
		packageDesc.ExtraDataType = g.QualifiedGoIdent(conf.ExtraType)
		packageDesc.NewExtraDataFunc = g.QualifiedGoIdent(conf.ExtraConstructor)
	}
	genConf := &genConfig{
		optionsKey:  conf.OptionsKey,
		packageDesc: packageDesc,
		methodSets:  make(map[string]int),
	}
	for _, service := range file.Services {
		err := generateService(g, service, genConf)
		if err != nil {
			return err
		}
	}
	return nil
}

func generateGoImport(g *protogen.GeneratedFile, conf *Config) {
	g.P("var _ = new(", contextPackage.Ident("Context"), ")")
	newRefs, exprRefs := importKeepAlives(conf)
	for _, ident := range newRefs {
		g.P("var _ = new(", ident, ")")
	}
	for _, ident := range exprRefs {
		g.P("var _ = ", ident)
	}
	g.P()
}

func generateService(g *protogen.GeneratedFile, service *protogen.Service, genConf *genConfig) error {
	if service.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated() {
		g.P("//")
		g.P(deprecationComment)
	}
	sd := &template.ServiceDesc{
		OptionsKey:  pascalCase(genConf.optionsKey),
		ServiceType: service.GoName,
		ServiceName: string(service.Desc.FullName()),
		Package:     genConf.packageDesc,
	}

	for _, method := range service.Methods {
		rule := extractOptionsRule(method, genConf.optionsKey)
		if rule == nil {
			continue
		}
		sd.Methods = append(sd.Methods, &template.MethodDesc{
			Name:         method.GoName,
			OriginalName: string(method.Desc.Name()),
			Num:          genConf.methodSets[method.GoName],
			Request:      g.QualifiedGoIdent(method.Input.GoIdent),
			Reply:        g.QualifiedGoIdent(method.Output.GoIdent),
			Comment:      formatMethodComment(string(method.Desc.Name()), string(method.Comments.Leading)),
			Extra:        rule.Extra,
		})
		genConf.methodSets[method.GoName]++
	}
	if len(sd.Methods) != 0 {
		content, err := sd.Execute()
		if err != nil {
			return err
		}
		g.P(content)
		g.P("\n\n")
	}
	return nil
}

func hasOptionsRule(services []*protogen.Service, key string) bool {
	for _, service := range services {
		for _, method := range service.Methods {
			if extractOptionsRule(method, key) != nil {
				return true
			}
		}
	}
	return false
}

func extractOptionsRule(method *protogen.Method, key string) *options.KeyValuePair {
	if method.Desc.IsStreamingClient() || method.Desc.IsStreamingServer() {
		return nil
	}
	if !proto.HasExtension(method.Desc.Options(), options.E_Options) {
		return nil
	}
	rules, ok := proto.GetExtension(method.Desc.Options(), options.E_Options).([]*options.KeyValuePair)
	if rules == nil || !ok {
		return nil
	}
	for _, rule := range rules {
		if rule.GetKey() == key {
			return rule
		}
	}
	return nil
}
