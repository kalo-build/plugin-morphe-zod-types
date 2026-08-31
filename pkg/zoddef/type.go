package zoddef

import "github.com/kalo-build/clone"

// ZodType represents a Zod type
type ZodType interface {
	GetZodTypeString() string
	GetTypeScriptType() string
	DeepClone() ZodType
}

// ZodPrimitiveType represents a primitive Zod type
type ZodPrimitiveType struct {
	TypeName string // e.g., "string", "number", "boolean", "date"
}

func (t ZodPrimitiveType) GetZodTypeString() string {
	return "z." + t.TypeName + "()"
}

func (t ZodPrimitiveType) GetTypeScriptType() string {
	switch t.TypeName {
	case "string":
		return "string"
	case "number":
		return "number"
	case "boolean":
		return "boolean"
	case "date":
		return "Date"
	case "any":
		return "unknown"
	default:
		return "unknown"
	}
}

func (t ZodPrimitiveType) DeepClone() ZodType {
	return ZodPrimitiveType{TypeName: t.TypeName}
}

// ZodArrayType represents an array type in Zod
type ZodArrayType struct {
	ElementType ZodType
}

func (t ZodArrayType) GetZodTypeString() string {
	return t.ElementType.GetZodTypeString() + ".array()"
}

func (t ZodArrayType) GetTypeScriptType() string {
	return t.ElementType.GetTypeScriptType() + "[]"
}

func (t ZodArrayType) DeepClone() ZodType {
	return ZodArrayType{ElementType: DeepCloneZodType(t.ElementType)}
}

// ZodEnumRefType represents a reference to an enum
type ZodEnumRefType struct {
	EnumName string
}

func (t ZodEnumRefType) GetZodTypeString() string {
	return t.EnumName + "Schema"
}

func (t ZodEnumRefType) GetTypeScriptType() string {
	return t.EnumName
}

func (t ZodEnumRefType) DeepClone() ZodType {
	return ZodEnumRefType{EnumName: t.EnumName}
}

// ZodSchemaRefType represents a reference to another schema
type ZodSchemaRefType struct {
	SchemaName string
}

func (t ZodSchemaRefType) GetZodTypeString() string {
	return t.SchemaName + "Schema"
}

func (t ZodSchemaRefType) GetTypeScriptType() string {
	return t.SchemaName
}

func (t ZodSchemaRefType) DeepClone() ZodType {
	return ZodSchemaRefType{SchemaName: t.SchemaName}
}

// ZodLazyType represents a lazy reference (for circular dependencies)
type ZodLazyType struct {
	TypeName string
}

func (t ZodLazyType) GetZodTypeString() string {
	return "z.lazy(() => " + t.TypeName + "Schema)"
}

func (t ZodLazyType) GetTypeScriptType() string {
	return t.TypeName
}

func (t ZodLazyType) DeepClone() ZodType {
	return ZodLazyType{TypeName: t.TypeName}
}

// ZodCoerceType wraps a primitive type with z.coerce for accepting
// string inputs (e.g., HTML form fields) and coercing them to the target type.
type ZodCoerceType struct {
	Inner ZodPrimitiveType
}

func (t ZodCoerceType) GetZodTypeString() string {
	return "z.coerce." + t.Inner.TypeName + "()"
}

func (t ZodCoerceType) GetTypeScriptType() string {
	return t.Inner.GetTypeScriptType()
}

func (t ZodCoerceType) DeepClone() ZodType {
	return ZodCoerceType{Inner: ZodPrimitiveType{TypeName: t.Inner.TypeName}}
}

// Common Zod types
var (
	ZodTypeString  = ZodPrimitiveType{TypeName: "string"}
	ZodTypeNumber  = ZodPrimitiveType{TypeName: "number"}
	ZodTypeBoolean = ZodPrimitiveType{TypeName: "boolean"}
	ZodTypeDate    = ZodCoerceType{Inner: ZodPrimitiveType{TypeName: "date"}}
	ZodTypeAny     = ZodPrimitiveType{TypeName: "any"}
)

// DeepCloneZodType creates a deep clone of a ZodType
func DeepCloneZodType(t ZodType) ZodType {
	if t == nil {
		return nil
	}
	return t.DeepClone()
}

// DeepCloneZodTypeSlice creates a deep clone of a slice of ZodTypes
func DeepCloneZodTypeSlice(types []ZodType) []ZodType {
	return clone.DeepCloneSlice(types)
}
