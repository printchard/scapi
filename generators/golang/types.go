package golang

import (
	"fmt"
	"unicode"

	"github.com/printchard/scapi/spec"
)

type GoGenerator struct {
	API      *spec.APISpec
	Resolver spec.TypeResolver
	Name     string
}

func capitalize(str string) string {
	if len(str) == 0 {
		return str
	}
	if unicode.IsLower(rune(str[0])) {
		return string(unicode.ToUpper(rune(str[0]))) + str[1:]
	}
	return str
}

func (g *GoGenerator) generateErrorTypeDef() string {
	formatter := spec.NewFormatter()
	formatter.Line("type HTTPError struct {")
	formatter.Indent()
	formatter.Line("Code int `json:\"code\"`")
	formatter.Line("Body any `json:\"body\"`")
	formatter.Dedent()
	formatter.Line("}")
	formatter.Line("")
	formatter.Line("func (e *HTTPError) Error() string {")
	formatter.Indent()
	formatter.Line("return fmt.Sprintf(\"HTTP %%d: %%v\", e.Code, e.Body)")
	formatter.Dedent()
	formatter.Line("}")
	formatter.Line("")
	return formatter.String()
}

func (g *GoGenerator) generateObjectTypeDef(typeName string, obj *spec.ObjectType) string {
	formatter := spec.NewFormatter()
	formatter.Line("type %s struct {", typeName)
	formatter.Indent()
	for fieldName, field := range obj.Fields {
		tags := fmt.Sprintf("`json:\"%s", fieldName)
		goType := g.generateGoType(field.Ref, field.Optional)
		if field.Optional {
			tags += ",omitempty"
		}
		tags += "\"`"

		formatter.Line("%s %s %s", capitalize(fieldName), goType, tags)
	}
	formatter.Dedent()
	formatter.Line("}")
	formatter.Line("")
	return formatter.String()
}

func (g *GoGenerator) generateGoType(tRef spec.TypeRef, optional bool) string {
	if g.Resolver.IsPrimitive(tRef) {
		primType, _ := g.Resolver.PrimitiveOf(tRef)
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
			return "any"
		}
		if optional {
			return "*" + typ
		}
		return typ
	} else if g.Resolver.IsObject(tRef) {
		if optional {
			return "*" + tRef.Name
		}
		return tRef.Name
	} else if g.Resolver.IsArray(tRef) {
		elemRef, _ := g.Resolver.ArrayElement(tRef)
		return "[]" + g.generateGoType(elemRef, false)
	}
	return "any"
}

func (g *GoGenerator) GenerateTypeDefs() string {
	result := "package main\n\nimport (\n  \"context\"\n  \"fmt\"\n  \"net/url\"\n  \"encoding/json\"\n  \"net/http\"\n)\n\n"
	result += g.generateErrorTypeDef()
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
		goType := g.generateGoType(field.Ref, field.Optional)
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
		goType := g.generateGoType(field.Ref, field.Optional)
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
