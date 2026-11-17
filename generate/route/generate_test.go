package route

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/go-sphere/options/sphere/options"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// newPlugin builds a plugin from a single hand-written file descriptor (no
// imports), so plugin.Files[0] is safely the target.
func newPlugin(t *testing.T, fd *descriptorpb.FileDescriptorProto) *protogen.Plugin {
	t.Helper()
	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{fd.GetName()},
		ProtoFile:      []*descriptorpb.FileDescriptorProto{fd},
	}
	plugin, err := protogen.Options{}.New(req)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}
	return plugin
}

// TestGenerateFile_NoService verifies that a file with no services produces no output.
func TestGenerateFile_NoService(t *testing.T) {
	fd := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("noservice.proto"),
		Package: proto.String("api.v1"),
		Options: &descriptorpb.FileOptions{GoPackage: proto.String("github.com/example/api;api")},
	}
	plugin := newPlugin(t, fd)

	genFile, err := GenerateFile(plugin, plugin.Files[0], testConfig())
	if err != nil {
		t.Fatalf("GenerateFile failed: %v", err)
	}
	if genFile != nil {
		t.Error("expected nil for file with no services, got non-nil")
	}
}

// TestGenerateFile_EmptyService verifies that a service with no methods produces
// no output.
func TestGenerateFile_EmptyService(t *testing.T) {
	fd := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("empty.proto"),
		Package: proto.String("api.v1"),
		Options: &descriptorpb.FileOptions{GoPackage: proto.String("github.com/example/api;api")},
		Service: []*descriptorpb.ServiceDescriptorProto{
			{Name: proto.String("EmptyService")},
		},
	}
	plugin := newPlugin(t, fd)

	genFile, err := GenerateFile(plugin, plugin.Files[0], testConfig())
	if err != nil {
		t.Fatalf("GenerateFile failed: %v", err)
	}
	if genFile != nil {
		t.Error("expected nil for empty service, got non-nil")
	}
}

// TestGenerateFile_NoOptions verifies that a service whose methods carry no
// (sphere.options.options) rule produces no output.
func TestGenerateFile_NoOptions(t *testing.T) {
	fd := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("plain.proto"),
		Package: proto.String("api.v1"),
		Options: &descriptorpb.FileOptions{GoPackage: proto.String("github.com/example/api;api")},
		MessageType: []*descriptorpb.DescriptorProto{
			{Name: proto.String("DoRequest")},
			{Name: proto.String("DoResponse")},
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name: proto.String("PlainService"),
				Method: []*descriptorpb.MethodDescriptorProto{
					{
						Name:       proto.String("Do"),
						InputType:  proto.String(".api.v1.DoRequest"),
						OutputType: proto.String(".api.v1.DoResponse"),
					},
				},
			},
		},
	}
	plugin := newPlugin(t, fd)

	genFile, err := GenerateFile(plugin, plugin.Files[0], testConfig())
	if err != nil {
		t.Fatalf("GenerateFile failed: %v", err)
	}
	if genFile != nil {
		t.Error("expected nil for service without options, got non-nil")
	}
}

func TestNewGeneratorSnapshotsConfig(t *testing.T) {
	cfg := testConfig()
	generator, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}
	cfg.OptionsKey = "changed"
	if generator.cfg.OptionsKey != DefaultOptionsKey {
		t.Fatalf("Generator OptionsKey = %q after caller mutation", generator.cfg.OptionsKey)
	}
	if _, err := NewGenerator(nil); err == nil {
		t.Fatal("NewGenerator(nil) error = nil")
	}
}

func TestGenerateFile_SkipsDeprecatedServiceWithoutMatchingMethods(t *testing.T) {
	routeMethodOpts := &descriptorpb.MethodOptions{}
	proto.SetExtension(routeMethodOpts, options.E_Options, []*options.KeyValuePair{
		{Key: "route"},
	})

	fd := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("deprecated.proto"),
		Package: proto.String("api.v1"),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{GoPackage: proto.String("github.com/example/api;api")},
		MessageType: []*descriptorpb.DescriptorProto{
			{Name: proto.String("DoRequest")},
			{Name: proto.String("DoResponse")},
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name:    proto.String("DeprecatedSkippedService"),
				Options: &descriptorpb.ServiceOptions{Deprecated: proto.Bool(true)},
				Method: []*descriptorpb.MethodDescriptorProto{
					{
						Name:       proto.String("Ignored"),
						InputType:  proto.String(".api.v1.DoRequest"),
						OutputType: proto.String(".api.v1.DoResponse"),
					},
				},
			},
			{
				Name: proto.String("ActiveService"),
				Method: []*descriptorpb.MethodDescriptorProto{
					{
						Name:       proto.String("Do"),
						InputType:  proto.String(".api.v1.DoRequest"),
						OutputType: proto.String(".api.v1.DoResponse"),
						Options:    routeMethodOpts,
					},
				},
			},
		},
	}

	plugin := newPlugin(t, fd)
	genFile, err := GenerateFile(plugin, plugin.Files[0], testConfig())
	if err != nil {
		t.Fatalf("GenerateFile failed: %v", err)
	}
	if genFile == nil {
		t.Fatal("expected generated file, got nil")
	}
	raw, err := genFile.Content()
	if err != nil {
		t.Fatalf("genFile.Content: %v", err)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "deprecated.route.pb.go", raw, parser.ParseComments); err != nil {
		t.Fatalf("generated code is not valid Go: %v\n%s", err, raw)
	}
	content := string(raw)
	if strings.Contains(content, "Deprecated: Do not use.") {
		t.Errorf("stray deprecation comment emitted for skipped service:\n%s", content)
	}
	if strings.Contains(content, "DeprecatedSkippedService") {
		t.Errorf("skipped deprecated service should not appear in output:\n%s", content)
	}
	if !strings.Contains(content, "ActiveServiceRouteServer") {
		t.Errorf("active service missing from output:\n%s", content)
	}

	t.Run("deprecated service with matching method keeps notice", func(t *testing.T) {
		fd2 := &descriptorpb.FileDescriptorProto{
			Name:    proto.String("deprecated_active.proto"),
			Package: proto.String("api.v1"),
			Syntax:  proto.String("proto3"),
			Options: &descriptorpb.FileOptions{GoPackage: proto.String("github.com/example/api;api")},
			MessageType: []*descriptorpb.DescriptorProto{
				{Name: proto.String("DoRequest")},
				{Name: proto.String("DoResponse")},
			},
			Service: []*descriptorpb.ServiceDescriptorProto{
				{
					Name:    proto.String("OldService"),
					Options: &descriptorpb.ServiceOptions{Deprecated: proto.Bool(true)},
					Method: []*descriptorpb.MethodDescriptorProto{
						{
							Name:       proto.String("Do"),
							InputType:  proto.String(".api.v1.DoRequest"),
							OutputType: proto.String(".api.v1.DoResponse"),
							Options:    routeMethodOpts,
						},
					},
				},
			},
		}
		p2 := newPlugin(t, fd2)
		gf2, err := GenerateFile(p2, p2.Files[0], testConfig())
		if err != nil {
			t.Fatalf("GenerateFile failed: %v", err)
		}
		if gf2 == nil {
			t.Fatal("expected generated file, got nil")
		}
		raw2, err := gf2.Content()
		if err != nil {
			t.Fatalf("gf2.Content: %v", err)
		}
		if !strings.Contains(string(raw2), deprecationComment) {
			t.Errorf("expected deprecation comment for active deprecated service:\n%s", raw2)
		}
	})
}

func TestGenerateFile_GoPackageNamedContext(t *testing.T) {
	routeMethodOpts := &descriptorpb.MethodOptions{}
	proto.SetExtension(routeMethodOpts, options.E_Options, []*options.KeyValuePair{
		{Key: "route"},
	})

	fd := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("ctx_pkg.proto"),
		Package: proto.String("ctx.v1"),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{GoPackage: proto.String("example.com/context;context")},
		MessageType: []*descriptorpb.DescriptorProto{
			{Name: proto.String("PingRequest")},
			{Name: proto.String("PingResponse")},
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name: proto.String("CtxService"),
				Method: []*descriptorpb.MethodDescriptorProto{
					{
						Name:       proto.String("Ping"),
						InputType:  proto.String(".ctx.v1.PingRequest"),
						OutputType: proto.String(".ctx.v1.PingResponse"),
						Options:    routeMethodOpts,
					},
				},
			},
		},
	}

	cfg := DefaultConfig()
	cfg.RequestType = mustParseGoIdent("example.com/other/context;Req")
	cfg.ResponseType = mustParseGoIdent("example.com/context;Resp")

	plugin := newPlugin(t, fd)
	genFile, err := GenerateFile(plugin, plugin.Files[0], cfg)
	if err != nil {
		t.Fatalf("GenerateFile failed: %v", err)
	}
	if genFile == nil {
		t.Fatal("expected generated file, got nil")
	}
	raw, err := genFile.Content()
	if err != nil {
		t.Fatalf("genFile.Content: %v", err)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "ctx_pkg.route.pb.go", raw, parser.ParseComments); err != nil {
		t.Fatalf("generated code is not valid Go: %v\n%s", err, raw)
	}
	content := string(raw)
	for _, want := range []string{
		"package context",
		`context "context"`,
		`context1 "example.com/other/context"`,
		"var _ = new(context.Context)",
		"var _ = new(context1.Req)",
		"var _ = new(Resp)",
		"Ping(context.Context, *PingRequest) (*PingResponse, error)",
		"DecodePingRequest(ctx context.Context, request *context1.Req) (*PingRequest, error)",
		"EncodePingResponse(ctx context.Context, response *PingResponse) (*Resp, error)",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("generated content missing %q:\n%s", want, content)
		}
	}
}
