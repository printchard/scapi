package spec

type TypeResolver interface {
	IsPrimitive(TypeRef) bool
	IsObject(TypeRef) bool
	PrimitiveOf(TypeRef) (PrimitiveType, bool)
	ObjectOf(TypeRef) (*ObjectType, bool)
	MustResolve(TypeRef) *Type
	IsOptional(Field) bool
	ResolveSuccessResponse(Endpoint) *Response
}

type defaultTypeResolver struct {
	api *APISpec
}

func NewTypeResolver(api *APISpec) TypeResolver {
	return &defaultTypeResolver{api: api}
}

func (r *defaultTypeResolver) IsPrimitive(ref TypeRef) bool {
	typ := r.api.ResolveTypeRefOrPanic(ref)
	return typ.Kind == Primitive
}

func (r *defaultTypeResolver) IsObject(ref TypeRef) bool {
	typ := r.api.ResolveTypeRefOrPanic(ref)
	return typ.Kind == Object
}

func (r *defaultTypeResolver) PrimitiveOf(ref TypeRef) (PrimitiveType, bool) {
	typ := r.api.ResolveTypeRefOrPanic(ref)
	if typ.Kind != Primitive {
		return 0, false
	}
	return typ.PrimitiveType, true
}

func (r *defaultTypeResolver) ObjectOf(ref TypeRef) (*ObjectType, bool) {
	typ := r.api.ResolveTypeRefOrPanic(ref)
	if typ.Kind != Object {
		return nil, false
	}
	return typ.ObjectType, true
}

func (r *defaultTypeResolver) MustResolve(ref TypeRef) *Type {
	return r.api.ResolveTypeRefOrPanic(ref)
}

func (r *defaultTypeResolver) IsOptional(field Field) bool {
	return field.Optional
}

func (r *defaultTypeResolver) ResolveSuccessResponse(endpoint Endpoint) *Response {
	for _, resp := range endpoint.Responses {
		if resp.Code >= 200 && resp.Code < 300 {
			return &resp
		}
	}
	panic("no success response found")
}
