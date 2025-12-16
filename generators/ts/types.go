package ts

import "github.com/printchard/scapi/spec"

type TsGenerator struct {
	API      *spec.APISpec
	Resolver spec.TypeResolver
	Name     string
}

func NewTsGenerator(api *spec.APISpec) *TsGenerator {
	return &TsGenerator{
		API:      api,
		Resolver: spec.NewTypeResolver(api),
	}
}

func (g *TsGenerator) generateErrorTypeDef() string {
	formatter := spec.NewFormatter()
	formatter.Line("export class HTTPError extends Error {")
	formatter.Indent()
	formatter.Line("code: number;")
	formatter.Line("body: unknown;")
	formatter.Line("constructor(code: number, body: unknown) {")
	formatter.Indent()
	formatter.Line("super(`HTTP ${code}: ${body}`);")
	formatter.Line("this.code = code;")
	formatter.Line("this.body = body;")
	formatter.Dedent()
	formatter.Line("}")
	formatter.Dedent()
	formatter.Line("}")
	formatter.Line("")
	return formatter.String()
}

func (g *TsGenerator) generateTsType(typeRef spec.TypeRef) string {
	if g.Resolver.IsPrimitive(typeRef) {
		primType, _ := g.Resolver.PrimitiveOf(typeRef)
		switch primType {
		case spec.String:
			return "string"
		case spec.Integer, spec.Float:
			return "number"
		case spec.Boolean:
			return "boolean"
		default:
			return "any"
		}
	} else if g.Resolver.IsObject(typeRef) {
		return typeRef.Name
	} else if g.Resolver.IsArray(typeRef) {
		elemRef, _ := g.Resolver.ArrayElement(typeRef)
		return g.generateTsType(elemRef) + "[]"
	}
	return "any"
}

func (g *TsGenerator) generateObjectTypeDef(typeName string, obj *spec.ObjectType) string {
	formatter := spec.NewFormatter()
	formatter.Line("export interface %s {", typeName)
	formatter.Indent()
	for fieldName, field := range obj.Fields {
		tsType := g.generateTsType(field.Ref)
		optionalMark := ""
		if field.Optional {
			optionalMark = "?"
		}
		nullableMark := ""
		if field.Nullable {
			nullableMark = " | null"
		}
		formatter.Line("%s%s: %s%s;", fieldName, optionalMark, tsType, nullableMark)
	}
	formatter.Dedent()
	formatter.Line("}")
	return formatter.String()
}

func (g *TsGenerator) GenerateTypeDefs() string {
	result := ""
	for typeName, typ := range g.API.Types {
		if typ.Kind != spec.Object {
			continue
		}
		result += g.generateObjectTypeDef(typeName, typ.ObjectType)
	}
	result += g.generateErrorTypeDef()
	return result
}
