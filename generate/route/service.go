package route

import (
	"github.com/go-sphere/options/sphere/options"
	"github.com/go-sphere/protoc-gen-route/generate/internal/template"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const deprecationComment = "// Deprecated: Do not use."

func generateService(g *protogen.GeneratedFile, service *protogen.Service, cfg *fileConfig, renderer *template.Renderer) error {
	if service.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated() {
		g.P("//")
		g.P(deprecationComment)
	}
	desc := &template.ServiceDesc{
		OptionsKey:  pascalCase(cfg.optionsKey),
		ServiceType: service.GoName,
		ServiceName: string(service.Desc.FullName()),
		Package:     cfg.packageDesc,
	}

	for _, method := range service.Methods {
		rule := extractOptionsRule(method, cfg.optionsKey)
		if rule == nil {
			continue
		}
		desc.Methods = append(desc.Methods, &template.MethodDesc{
			Name:         method.GoName,
			OriginalName: string(method.Desc.Name()),
			Num:          cfg.methodSets[method.GoName],
			Request:      g.QualifiedGoIdent(method.Input.GoIdent),
			Reply:        g.QualifiedGoIdent(method.Output.GoIdent),
			Comment:      formatMethodComment(string(method.Desc.Name()), string(method.Comments.Leading)),
			Extra:        rule.Extra,
		})
		cfg.methodSets[method.GoName]++
	}
	if len(desc.Methods) == 0 {
		return nil
	}
	content, err := renderer.Execute(desc)
	if err != nil {
		return err
	}
	g.P(content)
	g.P("\n\n")
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
