package dsl

import (
	"github.com/printchard/scapi/spec"
)

type Translator struct{}

func getTypeName(te TypeExpression) string {
	switch t := te.(type) {
	case SimpleTypeExpression:
		return t.Name
	case ArrayTypeExpression:
		return getTypeName(t.ElementType)
	default:
		return "unknown"
	}
}

func stringToHTTPMethod(s string) spec.HTTPMethod {
	switch s {
	case "GET":
		return spec.Get
	case "POST":
		return spec.Post
	case "PUT":
		return spec.Put
	case "DELETE":
		return spec.Delete
	case "PATCH":
		return spec.Patch
	}
	return 0
}

func (t *Translator) Translate(s *Spec) (*spec.APISpec, error) {
	types := make(map[string]*spec.Type)
	endpoints := []spec.Endpoint{}
	for _, decl := range s.Declarations {
		switch d := decl.(type) {
		case TypeDeclaration:
			objType := &spec.ObjectType{
				Fields: make(map[string]spec.Field),
			}
			for _, fieldDecl := range d.FieldDeclarations {
				card := spec.Single
				if _, ok := fieldDecl.Type.(ArrayTypeExpression); ok {
					card = spec.Multiple
				}
				fieldTypeRef := spec.TypeRef{Name: getTypeName(fieldDecl.Type)}
				field := spec.Field{
					Ref:         fieldTypeRef,
					Optional:    fieldDecl.Optional,
					Nullable:    fieldDecl.Nullable,
					Cardinality: card,
				}
				objType.Fields[fieldDecl.Identifier] = field
			}
			types[d.Identifier] = &spec.Type{
				Kind:       spec.Object,
				ObjectType: objType,
			}
		case EndpointDeclaration:
			endpoint := spec.Endpoint{
				Name:   d.Name,
				Method: stringToHTTPMethod(d.Method),
				Path:   spec.NewPathTemplate(d.Path),
				Input: &spec.InputShape{
					Params: make(map[string]spec.Field),
					Query:  make(map[string]spec.Field),
				},
				Responses: []spec.Response{},
			}

			for _, fieldDecl := range d.Body {
				switch fd := fieldDecl.(type) {
				case ParamsDeclaration:
					for _, paramField := range fd.Fields {
						card := spec.Single
						if _, ok := paramField.Type.(ArrayTypeExpression); ok {
							card = spec.Multiple
						}
						fieldTypeRef := spec.TypeRef{Name: getTypeName(paramField.Type)}
						field := spec.Field{
							Ref:         fieldTypeRef,
							Optional:    paramField.Optional,
							Nullable:    paramField.Nullable,
							Cardinality: card,
						}
						endpoint.Input.Params[paramField.Identifier] = field
					}
				case QueryDeclaration:
					for _, queryField := range fd.Fields {
						card := spec.Single
						if _, ok := queryField.Type.(ArrayTypeExpression); ok {
							card = spec.Multiple
						}
						fieldTypeRef := spec.TypeRef{Name: getTypeName(queryField.Type)}
						field := spec.Field{
							Ref:         fieldTypeRef,
							Optional:    queryField.Optional,
							Nullable:    queryField.Nullable,
							Cardinality: card,
						}
						endpoint.Input.Query[queryField.Identifier] = field
					}
				case BodyDeclaration:
					fieldTypeRef := spec.TypeRef{Name: getTypeName(fd.Type)}
					endpoint.Input.Body = &fieldTypeRef
				case ResponseDeclaration:
					endpoint.Responses = append(endpoint.Responses, spec.Response{
						Code: fd.Code,
						Ref:  &spec.TypeRef{Name: getTypeName(fd.Type)},
					})
				}
			}
			endpoints = append(endpoints, endpoint)
		}
	}
	return spec.NewAPISpec(s.Name, "localhost:8080", endpoints, types)
}
