package ts

import "github.com/printchard/scapi/spec"

func (g *TsGenerator) generateQueryCreation(f *spec.Formatter, endpoint spec.Endpoint) {
	f.Line("const queryParams = new URLSearchParams();")
	for name, field := range endpoint.Input.Query {
		f.Line("if (input.query.%s !== undefined) {", name)
		f.Indent()
		primType, _ := g.Resolver.PrimitiveOf(field.Ref)
		switch primType {
		case spec.Boolean:
			f.Line(`queryParams.append("%s", String(input.query.%s));`, name, name)
		case spec.Integer, spec.Float:
			f.Line(`queryParams.append("%s", String(input.query.%s));`, name, name)
		case spec.String:
			f.Line(`queryParams.append("%s", input.query.%s);`, name, name)
		}
		f.Dedent()
		f.Line("}")
	}
}

func (g *TsGenerator) generatePathCreation(f *spec.Formatter, endpoint spec.Endpoint) {
	comps := endpoint.Path.Components()
	if len(comps) == 0 {
		f.Line(`const path = "%s";`, endpoint.Path.String())
		return
	}

	f.Partial("const path = `")
	for _, comp := range comps {
		if comp.IsParam {
			f.Partial("${input.params.%s}", comp.Literal)
		} else {
			f.Partial("%s", comp.Literal)
		}
	}
	f.Partial("`;")
	f.Flush()
	f.Line("")
}

func (g *TsGenerator) generateClientMethod(f *spec.Formatter, endpoint spec.Endpoint) {
	methodName := endpoint.Name
	f.Partial("async %s(", methodName)

	if endpoint.Input != nil {
		f.Partial("input: {\n")
		f.Flush()
		f.Indent()
		if len(endpoint.Input.Params) > 0 {
			f.Line("params: {")
			f.Indent()
			for paramName, field := range endpoint.Input.Params {
				tsType := g.generateTsType(field.Ref)
				optionalMark := ""
				if field.Optional {
					optionalMark = "?"
				}
				f.Line("%s%s: %s;", paramName, optionalMark, tsType)
			}
			f.Dedent()
			f.Line("};")
		}
		if len(endpoint.Input.Query) > 0 {
			f.Line("query: {")
			f.Indent()
			for queryName, field := range endpoint.Input.Query {
				tsType := g.generateTsType(field.Ref)
				optionalMark := ""
				if field.Optional {
					optionalMark = "?"
				}
				f.Line("%s%s: %s;", queryName, optionalMark, tsType)
			}
			f.Dedent()
			f.Line("};")
		}
		if endpoint.Input.Body != nil {
			tsType := g.generateTsType(*endpoint.Input.Body)
			f.Line("body: %s;", tsType)
		}
		f.Dedent()
	}
	successResp := g.Resolver.ResolveSuccessResponse(endpoint)
	f.Line("}): Promise<%s> {", g.generateTsType(*successResp.Ref))
	f.Indent()
	g.generateQueryCreation(f, endpoint)
	g.generatePathCreation(f, endpoint)

	if endpoint.Input != nil && endpoint.Input.Query != nil {
		f.Line("const reqURL = `${this.baseURL}${path}?${queryParams.toString()}`;")
	} else {
		f.Line("const reqURL = `${this.baseURL}${path}`;")
	}

	f.Line("const response = await fetch(reqURL);")
	f.Line("if (!response.ok) {")
	f.Indent()
	f.Line("const errorBody = await response.json();")
	f.Line("throw new HTTPError(response.status, errorBody);")
	f.Dedent()
	f.Line("}")

	f.Line("const responseBody = await response.json();")
	f.Line("return responseBody as %s;", g.generateTsType(*successResp.Ref))
	f.Dedent()
	f.Line("}")
	f.Line("")
}

func (g *TsGenerator) GenerateClient() string {
	formatter := spec.NewFormatter()
	formatter.Line("export class APIClient {")
	formatter.Indent()
	formatter.Line("baseURL: string;")
	formatter.Line("constructor(baseURL: string) {")
	formatter.Indent()
	formatter.Line("this.baseURL = baseURL;")
	formatter.Dedent()
	formatter.Line("}")
	formatter.Line("")

	for _, endpoint := range g.API.Endpoints {
		g.generateClientMethod(formatter, endpoint)
	}

	formatter.Dedent()
	formatter.Line("}")
	formatter.Line("")
	return formatter.String()
}
