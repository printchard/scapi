package golang

import (
	"github.com/printchard/scapi/spec"
)

func (g *GoGenerator) generateQueryCreation(f *spec.Formatter, endpoint spec.Endpoint) {
	f.Line("query := url.Values{}")
	for name, field := range endpoint.Input.Query {
		f.Line("if input.Query.%s {", capitalize(name))
		f.Indent()
		typ, _ := g.Resolver.PrimitiveOf(field.Ref)
		switch typ {
		case spec.Boolean:
			f.Line(`query.Add("%s", fmt.Sprintf("%%t", input.Query.%s))`, name, capitalize(name))
		case spec.Integer:
			f.Line(`query.Add("%s", fmt.Sprintf("%%d", input.Query.%s))`, name, capitalize(name))
		case spec.Float:
			f.Line(`query.Add("%s", fmt.Sprintf("%%f", input.Query.%s))`, name, capitalize(name))
		case spec.String:
			f.Line(`query.Add("%s", input.Query.%s)`, name, capitalize(name))
		}
		f.Dedent()
		f.Line("}")
	}
}

func (g *GoGenerator) generatePathCreation(f *spec.Formatter, endpoint spec.Endpoint) {
	if endpoint.Input == nil || endpoint.Input.Params == nil {
		f.Line(`path := "%s"`, endpoint.Path.String())
		return
	}

	f.Line("path := fmt.Sprintf(\"%s\",", endpoint.Path.FormatString())
	f.Indent()
	for paramName, field := range endpoint.Input.Params {
		typ, _ := g.Resolver.PrimitiveOf(field.Ref)
		switch typ {
		case spec.String:
			f.Line("input.Params.%s,", capitalize(paramName))
		case spec.Integer:
			f.Line("fmt.Sprintf(\"%%d\", input.Params.%s)", capitalize(paramName))
		case spec.Float:
			f.Line("fmt.Sprintf(\"%%f\", input.Params.%s)", capitalize(paramName))
		case spec.Boolean:
			f.Line("fmt.Sprintf(\"%%t\", input.Params.%s)", capitalize(paramName))
		}
	}
	f.Dedent()
	f.Line(")")
}

func (g *GoGenerator) generateFetch(f *spec.Formatter, endpoint spec.Endpoint) {

	g.generateQueryCreation(f, endpoint)
	g.generatePathCreation(f, endpoint)

	if endpoint.Input != nil && endpoint.Input.Query != nil {
		f.Line(`reqURL := fmt.Sprintf("%s%%s?%%s", path, query.Encode())`, g.API.BaseURL)
	} else {
		f.Line(`reqURL := fmt.Sprintf("%s%%s", path)`, g.API.BaseURL)
	}
	f.Line(`response, err := http.Get(reqURL)`)
	f.Line("if err != nil {")
	f.Indent()
	f.Line("return nil, err")
	f.Dedent()
	f.Line("}")

	f.Line("defer response.Body.Close()")
	f.Line("decoder := json.NewDecoder(response.Body)")

	f.Line("switch response.StatusCode {")
	successResp := g.Resolver.ResolveSuccessResponse(endpoint)
	f.Line("case %d:", successResp.Code)
	f.Indent()
	f.Line("var successResp %s", g.generateGoType(*successResp.Ref, false))
	f.Line("if err := decoder.Decode(&successResp); err != nil {")
	f.Indent()
	f.Line("return nil, err")
	f.Dedent()
	f.Line("}")
	f.Line("return &successResp, nil")
	f.Dedent()

	for _, resp := range endpoint.Responses {
		if resp.Code == successResp.Code {
			continue
		}
		f.Line("case %d:", resp.Code)
		f.Indent()
		f.Line("var errorResp %s", g.generateGoType(*resp.Ref, false))
		f.Line("if err := decoder.Decode(&errorResp); err != nil {")
		f.Indent()
		f.Line("return nil, err")
		f.Dedent()
		f.Line("}")
		f.Line("return nil, &HTTPError{Code: %d, Body: errorResp}", resp.Code)
		f.Dedent()
	}
	f.Line("default:")
	f.Indent()
	f.Line("return nil, fmt.Errorf(\"unexpected status code: %%d\", response.StatusCode)")
	f.Dedent()
	f.Line("}")
}

func (g *GoGenerator) generateClientMethod(endpoint spec.Endpoint) string {
	defs := ""
	formatter := spec.NewFormatter()
	formatter.Partial("func (c *Client) %s(", endpoint.Name)
	if endpoint.Input != nil {
		formatter.Partial("ctx context.Context, ")
		defs += g.generateInputWrapper(endpoint)
		formatter.Partial("input %sInput", endpoint.Name)
	} else {
		formatter.Partial("ctx context.Context")
	}
	successResp := g.Resolver.ResolveSuccessResponse(endpoint)
	formatter.Partial(") (*%s, error) {\n", g.generateGoType(*successResp.Ref, false))
	formatter.Flush()
	formatter.Indent()
	if endpoint.Input != nil && endpoint.Input.Query != nil {
		formatter.Line("") // keep a blank line after query creation if desired
	}
	g.generateFetch(formatter, endpoint)
	formatter.Line("return nil, nil")
	formatter.Dedent()
	formatter.Line("}")

	return defs + formatter.String()
}

func (g *GoGenerator) GenerateClientMethods() string {
	formatter := spec.NewFormatter()
	formatter.Line("type Client struct {}")
	formatter.Line("")
	gen := ""
	for _, endpoint := range g.API.Endpoints {
		gen += g.generateClientMethod(endpoint)
	}

	return formatter.String() + gen
}
