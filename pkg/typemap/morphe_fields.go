package typemap

import (
	"fmt"

	"github.com/kalo-build/morphe-go/pkg/yaml"
	"github.com/kalo-build/plugin-morphe-zod-types/pkg/zoddef"
)

// MorpheFieldTypeToZodType converts a Morphe field type to a Zod type
func MorpheFieldTypeToZodType(fieldType yaml.ModelFieldType) (zoddef.ZodType, error) {
	switch fieldType {
	case yaml.ModelFieldTypeString:
		return zoddef.ZodTypeString, nil
	case yaml.ModelFieldTypeInteger:
		return zoddef.ZodTypeNumber, nil
	case yaml.ModelFieldTypeFloat:
		return zoddef.ZodTypeNumber, nil
	case yaml.ModelFieldTypeBoolean:
		return zoddef.ZodTypeBoolean, nil
	case yaml.ModelFieldTypeUUID:
		return zoddef.ZodTypeString, nil
	case yaml.ModelFieldTypeAutoIncrement:
		return zoddef.ZodTypeNumber, nil
	case yaml.ModelFieldTypeTime:
		return zoddef.ZodTypeDate, nil
	case yaml.ModelFieldTypeDate:
		return zoddef.ZodTypeDate, nil
	case yaml.ModelFieldTypeProtected:
		return zoddef.ZodTypeString, nil
	case yaml.ModelFieldTypeSealed:
		return zoddef.ZodTypeString, nil
	default:
		return nil, fmt.Errorf("unsupported field type: %s", fieldType)
	}
}

// MorpheStructureFieldTypeToZodType converts a Morphe structure field type to a Zod type
func MorpheStructureFieldTypeToZodType(fieldType yaml.StructureFieldType) (zoddef.ZodType, error) {
	switch fieldType {
	case yaml.StructureFieldTypeString:
		return zoddef.ZodTypeString, nil
	case yaml.StructureFieldTypeInteger:
		return zoddef.ZodTypeNumber, nil
	case yaml.StructureFieldTypeFloat:
		return zoddef.ZodTypeNumber, nil
	case yaml.StructureFieldTypeBoolean:
		return zoddef.ZodTypeBoolean, nil
	case yaml.StructureFieldTypeUUID:
		return zoddef.ZodTypeString, nil
	case yaml.StructureFieldTypeAutoIncrement:
		return zoddef.ZodTypeNumber, nil
	case yaml.StructureFieldTypeTime:
		return zoddef.ZodTypeDate, nil
	case yaml.StructureFieldTypeDate:
		return zoddef.ZodTypeDate, nil
	case yaml.StructureFieldTypeProtected:
		return zoddef.ZodTypeString, nil
	case yaml.StructureFieldTypeSealed:
		return zoddef.ZodTypeString, nil
	default:
		return nil, fmt.Errorf("unsupported structure field type: %s", fieldType)
	}
}
