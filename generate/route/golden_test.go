package route

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-sphere/protoc-gen-route/generate/internal/testutil"
	"google.golang.org/protobuf/compiler/protogen"
)

var updateGolden = flag.Bool("update-golden", false, "update golden files")

type goldenCase struct {
	name       string
	pbFile     string // testdata/pb/<name>.pb
	protoName  string // proto path inside the descriptor set
	wantFile   bool   // whether a generated file is expected
	goldenFile string // testdata/golden/<name>.route.pb.go (empty when wantFile is false)
	config     func() *Config
}

func goldenCases() []goldenCase {
	return []goldenCase{
		{
			name:       "basic",
			pbFile:     "testdata/pb/basic.pb",
			protoName:  "basic.proto",
			wantFile:   true,
			goldenFile: "testdata/golden/basic.route.pb.go",
		},
		{
			name:       "complex",
			pbFile:     "testdata/pb/complex.pb",
			protoName:  "complex.proto",
			wantFile:   true,
			goldenFile: "testdata/golden/complex.route.pb.go",
		},
		{
			// Same proto as basic, but without an extra-data type/constructor so
			// the no-extra import path and template branch are exercised.
			name:       "basic_no_extra",
			pbFile:     "testdata/pb/basic.pb",
			protoName:  "basic.proto",
			wantFile:   true,
			goldenFile: "testdata/golden/basic_no_extra.route.pb.go",
			config: func() *Config {
				c := DefaultConfig()
				c.ExtraType = protogen.GoIdent{}
				c.ExtraConstructor = protogen.GoIdent{}
				return c
			},
		},
		{
			// Generating the "bot" key against complex.proto: only
			// UserService.Delete carries that key.
			name:       "custom_key",
			pbFile:     "testdata/pb/complex.pb",
			protoName:  "complex.proto",
			wantFile:   true,
			goldenFile: "testdata/golden/custom_key.route.pb.go",
			config: func() *Config {
				c := DefaultConfig()
				c.OptionsKey = "bot"
				return c
			},
		},
		{
			name:      "no_options",
			pbFile:    "testdata/pb/no_options.pb",
			protoName: "no_options.proto",
			wantFile:  false,
		},
	}
}

// generate runs the plugin for a single case and returns the formatted content,
// or nil when no file was generated.
func (tt goldenCase) generate(t *testing.T) []byte {
	t.Helper()
	set := testutil.LoadDescriptorSet(t, tt.pbFile)
	plugin := testutil.MustCreatePlugin(t, set, tt.protoName)
	file := testutil.FileToGenerate(t, plugin)

	cfg := DefaultConfig()
	if tt.config != nil {
		cfg = tt.config()
	}

	genFile, err := GenerateFile(plugin, file, cfg)
	if err != nil {
		t.Fatalf("GenerateFile(%s) failed: %v", tt.name, err)
	}
	if genFile == nil {
		return nil
	}
	content, err := genFile.Content()
	if err != nil {
		t.Fatalf("Content(%s) failed: %v", tt.name, err)
	}
	return content
}

func TestGolden(t *testing.T) {
	for _, tt := range goldenCases() {
		t.Run(tt.name, func(t *testing.T) {
			content := tt.generate(t)

			if !tt.wantFile {
				if content != nil {
					t.Fatalf("expected no generated file, got %d bytes", len(content))
				}
				return
			}
			if content == nil {
				t.Fatal("expected a generated file, got nil")
			}

			if *updateGolden {
				if err := os.MkdirAll(filepath.Dir(tt.goldenFile), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(tt.goldenFile, content, 0o644); err != nil {
					t.Fatal(err)
				}
				t.Logf("updated golden file: %s", tt.goldenFile)
				return
			}

			expected, err := os.ReadFile(tt.goldenFile)
			if err != nil {
				t.Fatalf("failed to read golden file (run `make update-golden` to create): %v", err)
			}
			if diff := firstDiff(string(expected), string(content)); diff != "" {
				t.Errorf("generated content mismatch for %s (run `make update-golden` to refresh):\n%s", tt.name, diff)
			}
		})
	}
}

// TestGoldenDeterministic guards against non-deterministic output (e.g. map
// iteration order leaking into the generated file) by generating twice and
// comparing bytes.
func TestGoldenDeterministic(t *testing.T) {
	for _, tt := range goldenCases() {
		if !tt.wantFile {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			first := tt.generate(t)
			second := tt.generate(t)
			if string(first) != string(second) {
				t.Errorf("non-deterministic output for %s:\n%s", tt.name, firstDiff(string(first), string(second)))
			}
		})
	}
}

// firstDiff returns a human-readable description of the first differing line
// between want and got, or "" when they are equal. It avoids pulling in
// github.com/google/go-cmp as a module dependency.
func firstDiff(want, got string) string {
	if want == got {
		return ""
	}
	wl := strings.Split(want, "\n")
	gl := strings.Split(got, "\n")
	for i := 0; i < len(wl) && i < len(gl); i++ {
		if wl[i] != gl[i] {
			return fmt.Sprintf("first difference at line %d:\n  want: %q\n  got:  %q", i+1, wl[i], gl[i])
		}
	}
	return fmt.Sprintf("line count differs: want %d lines, got %d lines", len(wl), len(gl))
}
