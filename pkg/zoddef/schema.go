package zoddef

import "github.com/kalo-build/clone"

// Schema represents a Zod schema definition
type Schema struct {
	Name    string
	Imports []SchemaImport
	Fields  []SchemaField
}

func (s Schema) DeepClone() Schema {
	return Schema{
		Name:    s.Name,
		Imports: clone.DeepCloneSlice(s.Imports),
		Fields:  clone.DeepCloneSlice(s.Fields),
	}
}
