package main

import (
	"fmt"

	"github.com/printchard/scapi/generators/golang"
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

	api, err := spec.NewAPISpec("MyAPI", "http://localhost:8080", endpoints, types)
	if err != nil {
		panic(err)
	}
	// fmt.Println(api)

	gen := golang.NewGoGenerator(api)

	fmt.Println(gen.GenerateTypeDefs())
	// fmt.Println(gen.GenerateEndpoints())
	fmt.Println(gen.GenerateClientMethods())
}
