package route

import (
	"github.com/go-sphere/options/sphere/options"
	"github.com/go-sphere/protoc-gen-route/generate/internal/template"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const deprecationComment = "// Deprecated: Do not use."

type matchedMethod struct {
	method *protogen.Method
	rule   *options.KeyValuePair
}

type matchedService struct {
	service *protogen.Service
	methods []matchedMethod
}

func collectMatchedServices(services []*protogen.Service, key string) []matchedService {
	var matched []matchedService
	for _, service := range services {
		var methods []matchedMethod
		for _, method := range service.Methods {
			if rule := extractOptionsRule(method, key); rule != nil {
				methods = append(methods, matchedMethod{
					method: method,
					rule:   rule,
				})
			}
		}
		if len(methods) > 0 {
			matched = append(matched, matchedService{
				service: service,
				methods: methods,
			})
		}
	}
	return matched
}

func generateService(g *protogen.GeneratedFile, svc matchedService, cfg *fileConfig, renderer *template.Renderer) error {
	desc := &template.ServiceDesc{
		OptionsKey:  pascalCase(cfg.optionsKey),
		ServiceType: svc.service.GoName,
		ServiceName: string(svc.service.Desc.FullName()),
		Package:     cfg.packageDesc,
	}

	for _, m := range svc.methods {
		method := m.method
		desc.Methods = append(desc.Methods, &template.MethodDesc{
			Name:         method.GoName,
			OriginalName: string(method.Desc.Name()),
			Num:          cfg.methodSets[method.GoName],
			Request:      g.QualifiedGoIdent(method.Input.GoIdent),
			Reply:        g.QualifiedGoIdent(method.Output.GoIdent),
			Comment:      formatMethodComment(string(method.Desc.Name()), string(method.Comments.Leading)),
			Extra:        m.rule.Extra,
		})
		cfg.methodSets[method.GoName]++
	}
	if len(desc.Methods) == 0 {
		return nil
	}
	if opts, ok := svc.service.Desc.Options().(*descriptorpb.ServiceOptions); ok && opts.GetDeprecated() {
		g.P("//")
		g.P(deprecationComment)
	}
	desc.MethodSets = template.IndexMethods(desc.Methods)
	content, err := renderer.Execute(desc)
	if err != nil {
		return err
	}
	g.P(content)
	g.P("\n\n")
	return nil
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
