package route

import "google.golang.org/protobuf/compiler/protogen"

const contextPackage = protogen.GoImportPath("context")

func generateGoImport(g *protogen.GeneratedFile, cfg *Config) string {
	stdContext := contextPackage.Ident("Context")
	stdContextType := g.QualifiedGoIdent(stdContext)
	g.P("var _ = new(", stdContext, ")")
	newRefs, exprRefs := importKeepAlives(cfg)
	for _, ident := range newRefs {
		g.P("var _ = new(", ident, ")")
	}
	for _, ident := range exprRefs {
		g.P("var _ = ", ident)
	}
	g.P()
	return stdContextType
}
