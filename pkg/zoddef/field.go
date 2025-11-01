package zoddef

// SchemaField represents a field in a Zod schema
type SchemaField struct {
	Name     string
	ZodType  ZodType
	Optional bool
}

func (f SchemaField) DeepClone() SchemaField {
	return SchemaField{
		Name:     f.Name,
		ZodType:  DeepCloneZodType(f.ZodType),
		Optional: f.Optional,
	}
}
