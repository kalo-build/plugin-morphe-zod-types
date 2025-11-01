package zoddef

import "github.com/kalo-build/clone"

// Enum represents a Zod enum definition
type Enum struct {
	Name    string
	Type    ZodType
	Entries []EnumEntry
}

func (e Enum) DeepClone() Enum {
	return Enum{
		Name:    e.Name,
		Type:    DeepCloneZodType(e.Type),
		Entries: clone.DeepCloneSlice(e.Entries),
	}
}
