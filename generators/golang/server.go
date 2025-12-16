package golang

import (
	"github.com/printchard/scapi/spec"
)

func (g *GoGenerator) GenerateEndpointFunc(endpoint spec.Endpoint) (string, string) {
	defs := ""
	formatter := spec.NewFormatter()
	formatter.Partial("%s(", capitalize(endpoint.Name))
	if endpoint.Input != nil {
		formatter.Partial("ctx context.Context, ")
		defs += g.generateInputWrapper(endpoint)
		formatter.Partial("input %sInput", endpoint.Name)
	} else {
		formatter.Partial("ctx context.Context")
	}
	successResp := g.Resolver.ResolveSuccessResponse(endpoint)
	formatter.Partial(") (*%s, error)", successResp.Ref.Name)
	formatter.Flush()

	return defs, formatter.String()
}

func (g *GoGenerator) GenerateEndpoints() string {
	formatter := spec.NewFormatter()
	formatter.Line("type %sServer interface {", g.API.Name)
	formatter.Indent()
	defs := ""
	for _, endpoint := range g.API.Endpoints {
		def, funcStr := g.GenerateEndpointFunc(endpoint)
		defs += def
		formatter.Line("%s", funcStr)
	}
	formatter.Dedent()
	formatter.Line("}")
	return defs + formatter.String()
}

func NewGoGenerator(api *spec.APISpec) *GoGenerator {
	return &GoGenerator{
		API:      api,
		Resolver: spec.NewTypeResolver(api),
	}
}
