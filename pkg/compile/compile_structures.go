package compile

import (
	"fmt"

	"github.com/kalo-build/go-util/core"
	"github.com/kalo-build/go-util/strcase"
	"github.com/kalo-build/morphe-go/pkg/registry"
	"github.com/kalo-build/morphe-go/pkg/yaml"
	"github.com/kalo-build/plugin-morphe-zod-types/pkg/formatdef"
	"github.com/kalo-build/plugin-morphe-zod-types/pkg/typemap"
	"github.com/kalo-build/plugin-morphe-zod-types/pkg/zoddef"
)

// CompileAllStructures compiles all structures to Zod schemas
func CompileAllStructures(config MorpheCompileConfig, r *registry.Registry, writer *MorpheWriter) error {
	structureContents := make(map[string][]byte)

	for structureName, structure := range r.GetAllStructures() {
		// Compile the structure to Zod schema
		zodSchema, err := MorpheStructureToZodSchema(structure, r)
		if err != nil {
			return fmt.Errorf("failed to compile structure %s: %w", structureName, err)
		}

		// Generate the TypeScript file content
		content := generateZodStructureContent(structure.Name, zodSchema, r)
		structureContents[structureName] = content
	}

	// Write all structure contents
	return writer.WriteAllStructures(structureContents)
}

// MorpheStructureToZodSchema converts a Morphe structure to a Zod schema
func MorpheStructureToZodSchema(structure yaml.Structure, r *registry.Registry) (*zoddef.Schema, error) {
	schema := &zoddef.Schema{
		Name:    structure.Name,
		Imports: []zoddef.SchemaImport{},
		Fields:  []zoddef.SchemaField{},
	}

	// Process fields
	fieldNames := core.MapKeysSorted(structure.Fields)
	for _, fieldName := range fieldNames {
		field := structure.Fields[fieldName]

		zodType, err := getZodTypeForStructureField(field, r)
		if err != nil {
			return nil, err
		}

		// Structures don't have mandatory attributes typically, all fields are optional by default
		schema.Fields = append(schema.Fields, zoddef.SchemaField{
			Name:     strcase.ToCamelCase(fieldName),
			ZodType:  zodType,
			Optional: true,
		})
	}

	return schema, nil
}

// generateZodStructureContent generates the TypeScript file content for a structure Zod schema
func generateZodStructureContent(structureName string, schema *zoddef.Schema, r *registry.Registry) []byte {
	cb := formatdef.NewContentBuilder("  ")

	// Collect imports
	imports := collectImportsForSchema(schema, structureName, r)

	// Write imports
	cb.Line("import { z } from 'zod'")
	for _, imp := range imports {
		if len(imp.Names) > 0 {
			cb.Line("import { %s } from '%s'", joinStrings(imp.Names, ", "), imp.ImportPath)
		}
	}

	cb.EmptyLine()

	// Generate schema definition
	generateSchemaDefinition(cb, schema)
	cb.EmptyLine()

	return cb.Build()
}

// collectImportsForSchema gathers imports for a single schema
func collectImportsForSchema(schema *zoddef.Schema, currentName string, r *registry.Registry) []zoddef.SchemaImport {
	importMap := make(map[string]map[string]importKind)

	for _, field := range schema.Fields {
		addImportsForType(field.ZodType, currentName, importMap, r)
	}

	imports := []zoddef.SchemaImport{}
	for path, names := range importMap {
		nameSlice := []string{}
		for name, kind := range names {
			if kind == importType {
				nameSlice = append(nameSlice, "type "+name)
			} else {
				nameSlice = append(nameSlice, name)
			}
		}
		if len(nameSlice) > 0 {
			imports = append(imports, zoddef.SchemaImport{
				Names:      nameSlice,
				ImportPath: path,
			})
		}
	}

	return imports
}

// getZodTypeForStructureField returns the Zod type for a Morphe structure field
func getZodTypeForStructureField(field yaml.StructureField, r *registry.Registry) (zoddef.ZodType, error) {
	// Check if it's an enum field
	if _, exists := r.GetAllEnums()[string(field.Type)]; exists {
		return zoddef.ZodEnumRefType{EnumName: string(field.Type)}, nil
	}

	// Otherwise, map the primitive type
	return typemap.MorpheStructureFieldTypeToZodType(field.Type)
}
