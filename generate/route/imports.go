package route

import "google.golang.org/protobuf/compiler/protogen"

const contextPackage = protogen.GoImportPath("context")

func generateGoImport(g *protogen.GeneratedFile, cfg *Config) {
	g.P("var _ = new(", contextPackage.Ident("Context"), ")")
	newRefs, exprRefs := importKeepAlives(cfg)
	for _, ident := range newRefs {
		g.P("var _ = new(", ident, ")")
	}
	for _, ident := range exprRefs {
		g.P("var _ = ", ident)
	}
	g.P()
}
