// Package template renders the routing scaffolding emitted by protoc-gen-route.
package template

import (
	_ "embed"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"
)

//go:embed template.tmpl
var defaultTemplate string

/*
service MenuService {
  // test comment line1
  // test comment line2
  // test comment line3
  rpc UpdateCount(UpdateCountRequest) returns (UpdateCountResponse) {
    option (sphere.options.options) = {
      key: "bot"
      extra: {
        key: "command"
        value: "start"
      }
      extra: {
        key: "callback_query"
        value: "start"
      }
    };
  }
}
*/

// ServiceDesc is the template model for one generated route service.
type ServiceDesc struct {
	OptionsKey string // bot

	ServiceType string // MenuService
	ServiceName string // bot.v1.MenuService

	// Methods holds one entry per rpc with a route rule, in declaration
	// order. Names are unique within a service.
	Methods []*MethodDesc
	// MethodSets indexes Methods by Name (see IndexMethods). Ranging over it
	// visits names in sorted order; kept for custom templates.
	MethodSets map[string]*MethodDesc

	Package *PackageDesc
}

// MethodDesc is the template model for one generated route method.
type MethodDesc struct {
	Name         string // rpc method name: UpdateCount
	OriginalName string // service and method name: MenuServiceUpdateCount
	Num          int    // duplicate method number, used for generating unique method names

	Request string // rpc request type: UpdateCountRequest
	Reply   string // rpc reply type: UpdateCountResponse
	Comment string

	Extra map[string]string
}

// PackageDesc contains the qualified package-level identifiers used by a
// generated route service.
type PackageDesc struct {
	RequestType      string
	ResponseType     string
	ExtraDataType    string
	NewExtraDataFunc string
}

// Renderer owns a parsed route generation template. It is immutable after
// construction and safe to reuse for every file in one plugin invocation.
type Renderer struct {
	template *template.Template
}

// NewRenderer loads and parses the embedded template, or the file at path when
// path is non-empty.
func NewRenderer(path string) (*Renderer, error) {
	source := defaultTemplate
	if path != "" {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read template %q: %w", path, err)
		}
		source = string(raw)
	}
	tmpl, err := template.New("route").Funcs(template.FuncMap{
		"goString": strconv.Quote,
	}).Parse(source)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}
	return &Renderer{template: tmpl}, nil
}

// IndexMethods returns the MethodSets index for methods, keyed by method Name.
// Descriptor builders call it so Execute stays pure rendering.
func IndexMethods(methods []*MethodDesc) map[string]*MethodDesc {
	sets := make(map[string]*MethodDesc, len(methods))
	for _, m := range methods {
		sets[m.Name] = m
	}
	return sets
}

// Execute renders a service descriptor. It does not modify s; callers must
// populate MethodSets (see IndexMethods) before rendering.
func (r *Renderer) Execute(s *ServiceDesc) (string, error) {
	var buf strings.Builder
	if err := r.template.Execute(&buf, s); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}
