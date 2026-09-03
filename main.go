package main

import (
	"flag"
	"fmt"

	"github.com/go-sphere/protoc-gen-route/generate/route"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

var (
	showVersion = flag.Bool("version", false, "print the version and exit")

	optionsKey   = flag.String("options_key", "route", "options key in proto")
	templateFile = flag.String("template_file", "", "template file, if not set, use default template")

	requestModel   = flag.String("request_model", "", "request model")
	responseModel  = flag.String("response_model", "", "response model")
	extraDataModel = flag.String("extra_data_model", "", "extra data model")

	extraDataConstructor = flag.String("extra_data_constructor", "", "extra data constructor, and return a pointer of extra data")
)

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-route %v\n", "0.0.1")
		return
	}
	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(gen *protogen.Plugin) error {
		conf, err := extractConfig()
		if err != nil {
			return err
		}
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		err = route.ReplaceTemplateIfNeed(conf.TemplateFile)
		if err != nil {
			return err
		}
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			_, gErr := route.GenerateFile(gen, f, conf)
			if gErr != nil {
				return gErr
			}
		}
		return nil
	})
}

func extractConfig() (*route.Config, error) {
	if *requestModel == "" {
		return nil, fmt.Errorf("request_model is required (format: 'import/path;ModelName')")
	}
	_requestModel, err := route.ParseGoIdent(*requestModel)
	if err != nil {
		return nil, err
	}
	if *responseModel == "" {
		return nil, fmt.Errorf("response_model is required (format: 'import/path;ModelName')")
	}
	_responseModel, err := route.ParseGoIdent(*responseModel)
	if err != nil {
		return nil, err
	}

	conf := &route.Config{
		OptionsKey:   *optionsKey,
		TemplateFile: *templateFile,

		RequestType:  _requestModel,
		ResponseType: _responseModel,
	}

	if *extraDataModel == "" {
		if *extraDataConstructor != "" {
			return nil, fmt.Errorf("extra_data_model is required when extra_data_constructor is specified")
		}
		return conf, nil
	}
	if *extraDataConstructor == "" {
		return nil, fmt.Errorf("extra_data_constructor is required when extra_data_model is specified")
	}

	_extraDataModel, err := route.ParseGoIdent(*extraDataModel)
	if err != nil {
		return nil, err
	}
	_extraDataConstructor, err := route.ParseGoIdent(*extraDataConstructor)
	if err != nil {
		return nil, err
	}
	conf.ExtraType = _extraDataModel
	conf.ExtraConstructor = _extraDataConstructor

	return conf, nil
}
