package generators

import (
	"unicode"

	"github.com/printchard/scapi/spec"
)

func capitalize(str string) string {
	if len(str) == 0 {
		return str
	}
	if unicode.IsLower(rune(str[0])) {
		return string(unicode.ToUpper(rune(str[0]))) + str[1:]
	}
	return str
}

type GoGenerator struct {
	API      *spec.APISpec
	Resolver spec.TypeResolver
}

func (g *GoGenerator) generateObjectTypeDef(typeName string, obj *spec.ObjectType) string {
	formatter := spec.NewFormatter()
	formatter.Line("type %s struct {", typeName)
	formatter.Indent()
	for fieldName, field := range obj.Fields {
		goType := g.generateGoType(field.Ref.Name, field.Optional)
		formatter.Line("%s %s", capitalize(fieldName), goType)
	}
	formatter.Dedent()
	formatter.Line("}")
	formatter.Line("")
	return formatter.String()
}

func (g *GoGenerator) generateGoType(tName string, optional bool) string {
	if g.Resolver.IsPrimitive(spec.TypeRef{Name: tName}) {
		primType, _ := g.Resolver.PrimitiveOf(spec.TypeRef{Name: tName})
		var typ string
		switch primType {
		case spec.String:
			typ = "string"
		case spec.Integer:
			typ = "int"
		case spec.Float:
			typ = "float64"
		case spec.Boolean:
			typ = "bool"
		default:
			return "interface{}"
		}
		if optional {
			return "*" + typ
		}
		return typ
	} else if g.Resolver.IsObject(spec.TypeRef{Name: tName}) {
		if optional {
			return "*" + tName
		}
		return tName
	} else if g.Resolver.IsArray(spec.TypeRef{Name: tName}) {
		elemRef, _ := g.Resolver.ArrayElement(spec.TypeRef{Name: tName})
		return "[]" + g.generateGoType(elemRef.Name, false)
	}
	return "interface{}"
}

func (g *GoGenerator) GenerateTypeDefs() string {
	result := "package main\n\nimport \"context\"\n\n"
	for typeName, typ := range g.API.Types {
		if typ.Kind != spec.Object {
			continue
		}
		result += g.generateObjectTypeDef(typeName, typ.ObjectType)
	}
	return result
}

func (g *GoGenerator) generateParamsWrapper(endpoint spec.Endpoint) string {
	formatter := spec.NewFormatter()
	formatter.Line("type %sParams struct {", endpoint.Name)
	formatter.Indent()
	for paramName, field := range endpoint.Input.Params {
		goType := g.generateGoType(field.Ref.Name, field.Optional)
		formatter.Line("%s %s", capitalize(paramName), goType)
	}
	formatter.Dedent()
	formatter.Line("}")
	formatter.Line("")
	return formatter.String()
}

func (g *GoGenerator) generateQueryWrapper(endpoint spec.Endpoint) string {
	formatter := spec.NewFormatter()
	formatter.Line("type %sQuery struct {", endpoint.Name)
	formatter.Indent()
	for queryName, field := range endpoint.Input.Query {
		goType := g.generateGoType(field.Ref.Name, field.Optional)
		formatter.Line("%s %s", capitalize(queryName), goType)
	}
	formatter.Dedent()
	formatter.Line("}")
	formatter.Line("")
	return formatter.String()
}

func (g *GoGenerator) generateInputWrapper(endpoint spec.Endpoint) string {
	subTypeDefs := ""
	formatter := spec.NewFormatter()
	formatter.Line("type %sInput struct {", endpoint.Name)
	formatter.Indent()
	if endpoint.Input.Params != nil {
		formatter.Line("Params %sParams", endpoint.Name)
		subTypeDefs += g.generateParamsWrapper(endpoint)
	}

	if endpoint.Input.Query != nil {
		formatter.Line("Query %sQuery", endpoint.Name)
		subTypeDefs += g.generateQueryWrapper(endpoint)
	}

	if endpoint.Input.Body != nil {
		formatter.Line("Body %s", endpoint.Input.Body.Name)

	}

	formatter.Dedent()
	formatter.Line("}")
	formatter.Line("")
	return subTypeDefs + formatter.String()
}

func (g *GoGenerator) GenerateEndpointFunc(endpoint spec.Endpoint) string {
	defs := ""
	formatter := spec.NewFormatter()
	formatter.Line("func %s(", endpoint.Name)
	formatter.Indent()
	formatter.Line("ctx context.Context,")
	if endpoint.Input != nil {
		defs += g.generateInputWrapper(endpoint)
		formatter.Line("input %sInput,", endpoint.Name)
	}
	formatter.Dedent()
	formatter.Line(") (")
	formatter.Indent()
	successResp := g.Resolver.ResolveSuccessResponse(endpoint)
	formatter.Line("*%s,", successResp.Ref.Name)
	formatter.Line("error,")
	formatter.Dedent()
	formatter.Line(") {")
	formatter.Indent()
	formatter.Line("// TODO: Implement the function logic")
	formatter.Line("return nil, nil")
	formatter.Dedent()
	formatter.Line("}")
	formatter.Line("")

	return defs + formatter.String()
}

func (g *GoGenerator) GenerateEndpoints() string {
	result := ""
	for _, endpoint := range g.API.Endpoints {
		result += g.GenerateEndpointFunc(endpoint)
	}
	return result
}

func NewGoGenerator(api *spec.APISpec) *GoGenerator {
	return &GoGenerator{
		API:      api,
		Resolver: spec.NewTypeResolver(api),
	}
}
