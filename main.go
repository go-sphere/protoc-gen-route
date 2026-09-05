package main

import (
	"flag"
	"fmt"

	"github.com/go-sphere/protoc-gen-route/generate/route"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

const version = "0.0.1"

var (
	showVersion = flag.Bool("version", false, "print the version and exit")

	optionsKey   = flag.String("options_key", route.DefaultOptionsKey, "options key in proto")
	templateFile = flag.String("template_file", "", "template file, if not set, use default template")

	requestModel   = flag.String("request_model", "", "request model")
	responseModel  = flag.String("response_model", "", "response model")
	extraDataModel = flag.String("extra_data_model", "", "extra data model")

	extraDataConstructor = flag.String("extra_data_constructor", "", "extra data constructor, and return a pointer of extra data")
)

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-route %s\n", version)
		return
	}
	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(run)
}

func run(plugin *protogen.Plugin) error {
	plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
	cfg, err := extractConfig()
	if err != nil {
		return err
	}
	generator, err := route.NewGenerator(cfg)
	if err != nil {
		return err
	}
	for _, file := range plugin.Files {
		if !file.Generate {
			continue
		}
		if _, err := generator.GenerateFile(plugin, file); err != nil {
			return err
		}
	}
	return nil
}

func extractConfig() (*route.Config, error) {
	cfg := route.DefaultConfig()
	cfg.OptionsKey = *optionsKey
	cfg.TemplateFile = *templateFile

	if *requestModel != "" {
		ident, err := route.ParseGoIdent(*requestModel)
		if err != nil {
			return nil, err
		}
		cfg.RequestType = ident
	}
	if *responseModel != "" {
		ident, err := route.ParseGoIdent(*responseModel)
		if err != nil {
			return nil, err
		}
		cfg.ResponseType = ident
	}
	if *extraDataModel != "" {
		ident, err := route.ParseGoIdent(*extraDataModel)
		if err != nil {
			return nil, err
		}
		cfg.ExtraType = ident
	}
	if *extraDataConstructor != "" {
		ident, err := route.ParseGoIdent(*extraDataConstructor)
		if err != nil {
			return nil, err
		}
		cfg.ExtraConstructor = ident
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}
