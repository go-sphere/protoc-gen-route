package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteQuotesExtra(t *testing.T) {
	renderer, err := NewRenderer("")
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	methods := []*MethodDesc{{
		Name:         "Start",
		OriginalName: "Start",
		Request:      "StartRequest",
		Reply:        "StartResponse",
		Extra: map[string]string{
			`key "quoted"`: "line1\nline2",
		},
	}}
	out, err := renderer.Execute(&ServiceDesc{
		OptionsKey:  "Bot",
		ServiceType: "Menu",
		ServiceName: "bot.v1.Menu",
		Methods:     methods,
		MethodSets:  IndexMethods(methods),
		Package: &PackageDesc{
			RequestType:      "Request",
			ResponseType:     "Response",
			ExtraDataType:    "Extra",
			NewExtraDataFunc: "NewExtra",
		},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out, `"key \"quoted\"": "line1\nline2"`) {
		t.Errorf("extra map not Go-quoted, got:\n%s", out)
	}
}

func TestNewRendererIsIsolated(t *testing.T) {
	defaultRenderer, err := NewRenderer("")
	if err != nil {
		t.Fatalf("NewRenderer(default): %v", err)
	}

	path := filepath.Join(t.TempDir(), "custom.tmpl")
	if err := os.WriteFile(path, []byte("// custom {{.ServiceType}}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	customRenderer, err := NewRenderer(path)
	if err != nil {
		t.Fatalf("NewRenderer(custom): %v", err)
	}

	desc := &ServiceDesc{ServiceType: "Menu", Package: &PackageDesc{}}
	customOut, err := customRenderer.Execute(desc)
	if err != nil {
		t.Fatalf("custom Execute: %v", err)
	}
	if !strings.Contains(customOut, "// custom Menu") {
		t.Errorf("custom template not applied, got %q", customOut)
	}
	defaultOut, err := defaultRenderer.Execute(desc)
	if err != nil {
		t.Fatalf("default Execute: %v", err)
	}
	if strings.Contains(defaultOut, "// custom") {
		t.Fatal("custom renderer must not mutate the embedded default renderer")
	}
}

func TestExecuteDoesNotMutateInput(t *testing.T) {
	renderer, err := NewRenderer("")
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	desc := &ServiceDesc{
		ServiceType: "Menu",
		Methods:     []*MethodDesc{{Name: "Start"}},
		Package:     &PackageDesc{},
	}
	if _, err := renderer.Execute(desc); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if desc.MethodSets != nil {
		t.Errorf("Execute must not populate MethodSets, got %v", desc.MethodSets)
	}
}

func TestIndexMethodsLastWins(t *testing.T) {
	first := &MethodDesc{Name: "Start", Num: 0}
	second := &MethodDesc{Name: "Start", Num: 1}
	other := &MethodDesc{Name: "Stop"}
	sets := IndexMethods([]*MethodDesc{first, second, other})
	if len(sets) != 2 || sets["Start"] != second || sets["Stop"] != other {
		t.Errorf("unexpected index: %v", sets)
	}
}
