package spec_test

import (
	"testing"

	"github.com/printchard/scapi/spec"
)

func DefaultApiSpec() *spec.APISpec {
	endpoints := []spec.Endpoint{
		{
			Name:   "GetUser",
			Method: spec.Get,
			Path:   spec.NewPathTemplate("/users/{id}"),
			Input: &spec.InputShape{
				Params: map[string]spec.Field{
					"id": {Ref: spec.TypeRef{Name: "string"}},
				},
				Query: map[string]spec.Field{
					"verbose": {Ref: spec.TypeRef{Name: "boolean"}},
				},
				Body: &spec.TypeRef{Name: "UserRequest"},
			},
			Responses: []spec.Response{
				{Code: 200, Ref: &spec.TypeRef{Name: "UserResponse"}},
				{Code: 404, Ref: &spec.TypeRef{Name: "ErrorResponse"}},
			},
		},
	}

	types := map[string]*spec.Type{
		"UserResponse": {
			Kind: spec.Object,
			ObjectType: &spec.ObjectType{
				Fields: map[string]spec.Field{
					"id":   {Ref: spec.TypeRef{Name: "string"}},
					"name": {Ref: spec.TypeRef{Name: "string"}},
				},
			},
		},
		"ErrorResponse": {
			Kind: spec.Object,
			ObjectType: &spec.ObjectType{
				Fields: map[string]spec.Field{
					"error": {Ref: spec.TypeRef{Name: "string"}},
				},
			},
		},
		"UserRequest": {
			Kind: spec.Object,
			ObjectType: &spec.ObjectType{
				Fields: map[string]spec.Field{
					"id":  {Ref: spec.TypeRef{Name: "string"}},
					"age": {Ref: spec.TypeRef{Name: "number"}, Optional: true},
				},
			},
		},
	}
	return spec.NewAPISpec(endpoints, types)
}

func TestParseApiSpec(t *testing.T) {
	api := DefaultApiSpec()
	if err := api.Validate(); err != nil {
		t.Fatalf("expected valid API spec, got error: %v", err)
	}
}
