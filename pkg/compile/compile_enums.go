package compile

import (
	"fmt"

	"github.com/kalo-build/go-util/core"
	"github.com/kalo-build/morphe-go/pkg/registry"
	"github.com/kalo-build/morphe-go/pkg/yaml"
	"github.com/kalo-build/plugin-morphe-zod-types/pkg/formatdef"
	"github.com/kalo-build/plugin-morphe-zod-types/pkg/zoddef"
)

// CompileAllEnums compiles all enums to Zod schemas
func CompileAllEnums(config MorpheCompileConfig, r *registry.Registry, writer *MorpheWriter) error {
	enumContents := make(map[string][]byte)

	for enumName, enum := range r.GetAllEnums() {
		// Compile the enum to Zod schema
		zodEnum, err := MorpheEnumToZodEnum(enum)
		if err != nil {
			return fmt.Errorf("failed to compile enum %s: %w", enumName, err)
		}

		// Generate the TypeScript file content
		content := generateZodEnumContent(zodEnum)
		enumContents[enumName] = content
	}

	// Write all enum contents
	return writer.WriteAllEnums(enumContents)
}

// MorpheEnumToZodEnum converts a Morphe enum to a Zod enum definition
func MorpheEnumToZodEnum(enum yaml.Enum) (*zoddef.Enum, error) {
	zodEnum := &zoddef.Enum{
		Name:    enum.Name,
		Type:    mapEnumTypeToZod(enum.Type),
		Entries: make([]zoddef.EnumEntry, 0, len(enum.Entries)),
	}

	// Sort entries for consistent output
	entryNames := core.MapKeysSorted(enum.Entries)

	// Convert each enum entry
	for _, entryName := range entryNames {
		entry := zoddef.EnumEntry{
			Name:  entryName,
			Value: enum.Entries[entryName],
		}
		zodEnum.Entries = append(zodEnum.Entries, entry)
	}

	return zodEnum, nil
}

// mapEnumTypeToZod maps Morphe enum types to Zod types
func mapEnumTypeToZod(morpheType yaml.EnumType) zoddef.ZodType {
	switch morpheType {
	case yaml.EnumTypeInteger, yaml.EnumTypeFloat:
		return zoddef.ZodTypeNumber
	case yaml.EnumTypeString:
		return zoddef.ZodTypeString
	default:
		return zoddef.ZodTypeString
	}
}

// generateZodEnumContent generates the TypeScript file content for a Zod enum schema
func generateZodEnumContent(zodEnum *zoddef.Enum) []byte {
	cb := formatdef.NewContentBuilder("  ")

	// Import statement
	cb.Line("import { z } from 'zod'")

	// Generate the native TypeScript enum
	cb.Line("export enum %s {", zodEnum.Name)
	cb.Indent()

	for i, entry := range zodEnum.Entries {
		// Format the entry value based on type
		var valueStr string
		switch v := entry.Value.(type) {
		case string:
			valueStr = fmt.Sprintf("'%s'", v)
		default:
			valueStr = fmt.Sprintf("%v", v)
		}

		line := fmt.Sprintf("%s = %s", entry.Name, valueStr)
		if i < len(zodEnum.Entries)-1 {
			line += ","
		}
		cb.Line(line)
	}

	cb.Dedent()
	cb.Line("}")

	cb.EmptyLine()
	// Generate the Zod schema
	cb.Line("export const %sSchema = z.nativeEnum(%s)", zodEnum.Name, zodEnum.Name)
	cb.EmptyLine()
	// Generate the TypeScript type inference
	cb.Line("export type %sType = z.infer<typeof %sSchema>", zodEnum.Name, zodEnum.Name)
	cb.EmptyLine()

	return cb.Build()
}
