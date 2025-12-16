package spec

import (
	"fmt"
	"net/url"
)

type InputShape struct {
	Params map[string]Field
	Query  map[string]Field
	Body   *TypeRef
}

type Endpoint struct {
	Name      string
	Method    HTTPMethod
	Path      *PathTemplate
	Input     *InputShape
	Responses []Response
}

type APISpec struct {
	Endpoints []Endpoint
	Types     map[string]*Type
	Name      string
	BaseURL   string
}

func (s *APISpec) ResolveTypeRef(typeRef TypeRef) (*Type, bool) {
	t, ok := s.Types[typeRef.Name]
	if !ok {
		return nil, false
	}
	return t, true
}

func (s *APISpec) ResolveTypeRefOrPanic(typeRef TypeRef) *Type {
	t, ok := s.ResolveTypeRef(typeRef)
	if !ok {
		panic(fmt.Sprintf("unresolved type reference: %s", typeRef.Name))
	}
	return t
}

func validateObject(obj *ObjectType, api *APISpec) error {
	for fieldName, field := range obj.Fields {
		typ, ok := api.ResolveTypeRef(field.Ref)
		if !ok {
			return fmt.Errorf("unresolved type reference: %s in field %s", field.Ref.Name, fieldName)
		}
		switch typ.Kind {
		case Object:
			if err := validateObject(typ.ObjectType, api); err != nil {
				return err
			}
		case Array:
			if err := validateArray(typ.ArrayType, api); err != nil {
				return err
			}
		case Primitive:
			// Primitive types are always valid
		default:
			return fmt.Errorf("unknown type kind: %s in field %s", typ.Kind, fieldName)
		}
	}
	return nil
}

func validateArray(arr *ArrayType, api *APISpec) error {
	typ, ok := api.ResolveTypeRef(arr.ElementType)
	if !ok {
		return fmt.Errorf("unresolved type reference: %s in array element type", arr.ElementType.Name)
	}
	switch typ.Kind {
	case Object:
		return validateObject(typ.ObjectType, api)
	case Array:
		return validateArray(typ.ArrayType, api)
	case Primitive:
		return nil
	}
	return fmt.Errorf("unknown type kind: %s in array element type", typ.Kind)
}

func (api *APISpec) ValidateTypes() error {
	for typeName, typ := range api.Types {
		switch typ.Kind {
		case Object:
			if err := validateObject(typ.ObjectType, api); err != nil {
				return err
			}
		case Array:
			if err := validateArray(typ.ArrayType, api); err != nil {
				return err
			}
		case Primitive:
			// Primitive types are always valid
		default:
			return fmt.Errorf("unknown type kind: %s for type %s", typ.Kind, typeName)
		}
	}
	return nil
}

func (api *APISpec) ValidateEndpoints() error {
	for _, endpoint := range api.Endpoints {
		if endpoint.Input != nil {
			for paramName, field := range endpoint.Input.Params {
				if field.Optional {
					return fmt.Errorf("endpoint %s param %s cannot be optional", endpoint.Name, paramName)
				}
				if field.Nullable {
					return fmt.Errorf("endpoint %s param %s cannot be nullable", endpoint.Name, paramName)
				}
				if _, ok := api.ResolveTypeRef(field.Ref); !ok {
					return fmt.Errorf("unresolved type reference: %s in endpoint %s param %s", field.Ref.Name, endpoint.Name, paramName)
				}
			}
			for queryName, field := range endpoint.Input.Query {
				if _, ok := api.ResolveTypeRef(field.Ref); !ok {
					return fmt.Errorf("unresolved type reference: %s in endpoint %s query %s", field.Ref.Name, endpoint.Name, queryName)
				}
			}
			if endpoint.Input.Body != nil {
				if _, ok := api.ResolveTypeRef(*endpoint.Input.Body); !ok {
					return fmt.Errorf("unresolved type reference: %s in endpoint %s body", endpoint.Input.Body.Name, endpoint.Name)
				}
			}
		}

		if len(endpoint.Responses) == 0 {
			return fmt.Errorf("endpoint %s has no responses defined", endpoint.Name)
		}

		duplicateCheck := make(map[int]bool)
		for _, resp := range endpoint.Responses {
			if resp.Code < 100 || resp.Code > 599 {
				return fmt.Errorf("endpoint %s has invalid response code: %d", endpoint.Name, resp.Code)
			}
			if _, exists := duplicateCheck[resp.Code]; exists {
				return fmt.Errorf("endpoint %s has duplicate response code: %d", endpoint.Name, resp.Code)
			}
			duplicateCheck[resp.Code] = true

			if resp.Ref != nil {
				if _, ok := api.ResolveTypeRef(*resp.Ref); !ok {
					return fmt.Errorf("unresolved type reference: %s in endpoint %s response %d", resp.Ref.Name, endpoint.Name, resp.Code)
				}
			}
		}
	}
	return nil
}

func (api *APISpec) ValidatePaths() error {
	for _, endpoint := range api.Endpoints {
		if endpoint.Path == nil || endpoint.Path.template == "" {
			return fmt.Errorf("endpoint %s has invalid path template", endpoint.Name)
		}

		if endpoint.Input != nil && len(endpoint.Path.Params()) > 0 {
			if len(endpoint.Path.Params()) != len(endpoint.Input.Params) {
				return fmt.Errorf("endpoint %s path parameters do not match input parameters", endpoint.Name)
			}
			for _, param := range endpoint.Path.Params() {
				if _, ok := endpoint.Input.Params[param]; !ok {
					return fmt.Errorf("endpoint %s path parameter %s not defined in input parameters", endpoint.Name, param)
				}
			}
		}

		if len(endpoint.Path.params) > 0 && (endpoint.Input == nil || len(endpoint.Input.Params) == 0) {
			return fmt.Errorf("endpoint %s specifies path parameters not defined in template", endpoint.Name)
		}
	}
	return nil
}

func (api *APISpec) Validate() error {
	if api.Name == "" {
		return fmt.Errorf("API name cannot be empty")
	}

	if api.BaseURL == "" {
		return fmt.Errorf("API base URL cannot be empty")
	}

	if _, err := url.Parse(api.BaseURL); err != nil {
		return fmt.Errorf("invalid API base URL: %v", err)
	}

	if err := api.ValidateTypes(); err != nil {
		return err
	}
	if err := api.ValidateEndpoints(); err != nil {
		return err
	}
	if err := api.ValidatePaths(); err != nil {
		return err
	}
	return nil
}

func (api *APISpec) Responses() []Response {
	var responses []Response
	for _, endpoint := range api.Endpoints {
		responses = append(responses, endpoint.Responses...)
	}
	return responses
}

func (api *APISpec) String() string {
	f := NewFormatter()

	f.Line("API Specification:")
	f.Indent()

	f.Line("Endpoints:")
	f.Indent()
	for _, endpoint := range api.Endpoints {
		f.Line("Endpoint: %s", endpoint.Name)
		f.Indent()
		f.Line("Method: %s", endpoint.Method)
		f.Line("Path: %s", endpoint.Path.String())
		if endpoint.Input != nil {
			f.Line("Input:")
			f.Indent()
			if len(endpoint.Input.Params) > 0 {
				f.Line("Params:")
				f.Indent()
				for paramName, field := range endpoint.Input.Params {
					typ := api.Types[field.Ref.Name]
					f.Line("- %s: %s", paramName, typ.Kind)
				}
				f.Dedent()
			}
			if len(endpoint.Input.Query) > 0 {
				f.Line("Query:")
				f.Indent()
				for queryName, field := range endpoint.Input.Query {
					typ := api.Types[field.Ref.Name]
					f.Line("- %s: %s", queryName, typ.Kind)
				}
				f.Dedent()
			}
			if endpoint.Input.Body != nil {
				f.Line("Body:")
				f.Indent()
				typ := api.Types[endpoint.Input.Body.Name]
				switch typ.Kind {
				case Object:
					for fieldName, field := range typ.ObjectType.Fields {
						fieldType := api.Types[field.Ref.Name]
						optionalStr := ""
						if field.Optional {
							optionalStr = " (optional)"
						}
						f.Line("- %s: %s%s", fieldName, fieldType.Kind, optionalStr)
					}
				default:
					f.Line("- %s", typ.Kind)
				}
				f.Dedent()
			}
			f.Dedent()
		}
		f.Line("Responses:")
		f.Indent()
		for _, resp := range endpoint.Responses {
			if resp.Ref != nil {
				typ := api.Types[resp.Ref.Name]
				f.Line("- %d: %s", resp.Code, typ.Kind)
			} else {
				f.Line("- %d: no body", resp.Code)
			}
		}
		f.Dedent()
	}

	return f.String()
}

func NewAPISpec(name string, baseURL string, endpoints []Endpoint, types map[string]*Type) (*APISpec, error) {
	types["string"] = &Type{Kind: Primitive, PrimitiveType: String}
	types["integer"] = &Type{Kind: Primitive, PrimitiveType: Integer}
	types["float"] = &Type{Kind: Primitive, PrimitiveType: Float}
	types["boolean"] = &Type{Kind: Primitive, PrimitiveType: Boolean}
	normalizedUrl := baseURL
	if len(normalizedUrl) > 0 && normalizedUrl[len(normalizedUrl)-1] != '/' {
		normalizedUrl += "/"
	}

	api := &APISpec{
		Name:      name,
		BaseURL:   normalizedUrl,
		Endpoints: endpoints,
		Types:     types,
	}

	if err := api.Validate(); err != nil {
		return nil, fmt.Errorf("invalid API specification: %v", err)
	}
	return api, nil
}
