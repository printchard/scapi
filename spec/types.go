package spec

type HTTPMethod int

const (
	Get HTTPMethod = iota
	Post
	Put
	Delete
	Patch
)

func (m HTTPMethod) String() string {
	switch m {
	case Get:
		return "GET"
	case Post:
		return "POST"
	case Put:
		return "PUT"
	case Delete:
		return "DELETE"
	case Patch:
		return "PATCH"
	default:
		return "UNKNOWN"
	}
}

type TypeKind string

const (
	Object    TypeKind = "object"
	Primitive TypeKind = "primitive"
)

type Type struct {
	Kind          TypeKind
	ObjectType    *ObjectType
	PrimitiveType PrimitiveType
}

type ObjectType struct {
	Fields map[string]Field
}

type Cardinality int

const (
	Single Cardinality = iota
	Multiple
)

type Field struct {
	Ref         TypeRef
	Cardinality Cardinality
	Optional    bool
	Nullable    bool
}

type PrimitiveType int

const (
	String PrimitiveType = iota
	Integer
	Float
	Boolean
)

func (p PrimitiveType) String() string {
	switch p {
	case String:
		return "string"
	case Integer:
		return "integer"
	case Float:
		return "float"
	case Boolean:
		return "boolean"
	default:
		return "unknown"
	}
}

type TypeRef struct {
	Name string
}

type Response struct {
	Code int
	Ref  *TypeRef
}
