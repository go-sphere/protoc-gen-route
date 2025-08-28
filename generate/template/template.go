package template

import (
	_ "embed"
	"os"
	"strings"
	"text/template"
)

//go:embed template.tmpl
var routeTemplate string

/*
service MenuService {
  // test comment line1
  // test comment line2
  // test comment line3
  rpc UpdateCount(UpdateCountRequest) returns (UpdateCountResponse) {
    option (sphere.options.options) = {
      key: "bot"
      extra: [
        {
          key: "command"
          value: "start"
        },
        {
          key: "callback_query"
          value: "start"
        }
      ]
    };
  }
}
*/

type ServiceDesc struct {
	OptionsKey string // bot

	ServiceType string // MenuService
	ServiceName string // bot.v1.MenuService

	Methods    []*MethodDesc
	MethodSets map[string]*MethodDesc

	Package *PackageDesc
}

type MethodDesc struct {
	Name         string // rpc method name: UpdateCount
	OriginalName string // service and method name: MenuServiceUpdateCount
	Num          int    // duplicate method number, used for generating unique method names

	Request string // rpc request type: UpdateCountRequest
	Reply   string // rpc reply type: UpdateCountResponse
	Comment string

	Extra map[string]string
}

type PackageDesc struct {
	RequestType      string
	ResponseType     string
	ExtraDataType    string
	NewExtraDataFunc string
}

func (s *ServiceDesc) Execute() (string, error) {
	s.MethodSets = make(map[string]*MethodDesc)
	for _, m := range s.Methods {
		s.MethodSets[m.Name] = m
	}
	var buf strings.Builder
	tmpl, err := template.New("route").Parse(routeTemplate)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(&buf, s)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func ReplaceTemplateIfNeed(path string) error {
	if path != "" {
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		routeTemplate = string(raw)
	}
	return nil
}
