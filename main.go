package main

import (
	"fmt"

	gogen "github.com/printchard/scapi/generators/go"
	"github.com/printchard/scapi/spec"
)

func main() {
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
					"age": {Ref: spec.TypeRef{Name: "integer"}, Optional: true},
				},
			},
		},
	}
	api := spec.NewAPISpec(endpoints, types)

	err := api.ValidateTypes()
	if err != nil {
		fmt.Printf("API Specification validation error: %v\n", err)
		return
	}
	err = api.ValidateEndpoints()
	if err != nil {
		fmt.Printf("API Specification validation error: %v\n", err)
		return
	}
	err = api.ValidatePaths()
	if err != nil {
		fmt.Printf("API Specification validation error: %v\n", err)
		return
	}

	fmt.Println(api)

	gen := gogen.NewGoGenerator(api)

	fmt.Println(gen.GenerateTypeDefs())
	fmt.Println(gen.GenerateEndpoints())
}
