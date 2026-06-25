package route

import (
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// newPlugin builds a plugin from a single hand-written file descriptor (no
// imports), so plugin.Files[0] is safely the target. Hand-written descriptors
// cannot carry the (sphere.options.options) extension, so these tests only cover
// the skip paths; the extension-driven paths are covered by the golden tests.
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

	genFile, err := GenerateFile(plugin, plugin.Files[0], DefaultConfig())
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

	genFile, err := GenerateFile(plugin, plugin.Files[0], DefaultConfig())
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

	genFile, err := GenerateFile(plugin, plugin.Files[0], DefaultConfig())
	if err != nil {
		t.Fatalf("GenerateFile failed: %v", err)
	}
	if genFile != nil {
		t.Error("expected nil for service without options, got non-nil")
	}
}
