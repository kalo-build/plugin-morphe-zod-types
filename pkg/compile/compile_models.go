package compile

import (
	"fmt"
	"strings"

	"github.com/kalo-build/go-util/core"
	"github.com/kalo-build/go-util/strcase"
	"github.com/kalo-build/morphe-go/pkg/registry"
	"github.com/kalo-build/morphe-go/pkg/yaml"
	"github.com/kalo-build/plugin-morphe-zod-types/pkg/formatdef"
	"github.com/kalo-build/plugin-morphe-zod-types/pkg/typemap"
	"github.com/kalo-build/plugin-morphe-zod-types/pkg/zoddef"
)

// CompileAllModels compiles all models to Zod schemas
func CompileAllModels(config MorpheCompileConfig, r *registry.Registry, writer *MorpheWriter) error {
	modelContents := make(map[string][]byte)

	for modelName, model := range r.GetAllModels() {
		// Compile the model to Zod schema
		zodSchemas, err := MorpheModelToZodSchemas(model, r)
		if err != nil {
			return fmt.Errorf("failed to compile model %s: %w", modelName, err)
		}

		// Generate the TypeScript file content
		content := generateZodModelContent(model.Name, zodSchemas, r)
		modelContents[modelName] = content
	}

	// Write all model contents
	return writer.WriteAllModels(modelContents)
}

// MorpheModelToZodSchemas converts a Morphe model to Zod schema definitions
func MorpheModelToZodSchemas(model yaml.Model, r *registry.Registry) ([]*zoddef.Schema, error) {
	schemas := []*zoddef.Schema{}

	// Create the main model schema
	mainSchema := &zoddef.Schema{
		Name:    model.Name,
		Imports: []zoddef.SchemaImport{},
		Fields:  []zoddef.SchemaField{},
	}

	// Process regular fields
	fieldNames := core.MapKeysSorted(model.Fields)
	for _, fieldName := range fieldNames {
		field := model.Fields[fieldName]

		zodType, err := getZodTypeForField(field, r)
		if err != nil {
			return nil, err
		}

		// Check if field is mandatory
		isMandatory := false
		for _, attr := range field.Attributes {
			if attr == "mandatory" {
				isMandatory = true
				break
			}
		}

		mainSchema.Fields = append(mainSchema.Fields, zoddef.SchemaField{
			Name:     strcase.ToCamelCase(fieldName),
			ZodType:  zodType,
			Optional: !isMandatory,
		})
	}

	// Process relationships
	if len(model.Related) > 0 {
		relationFields, err := getRelationshipFields(model.Related, r)
		if err != nil {
			return nil, err
		}
		mainSchema.Fields = append(mainSchema.Fields, relationFields...)
	}

	schemas = append(schemas, mainSchema)

	// Create identifier schemas
	identifierSchemas, err := getIdentifierSchemas(model, mainSchema)
	if err != nil {
		return nil, err
	}
	schemas = append(schemas, identifierSchemas...)

	return schemas, nil
}

// getZodTypeForField returns the Zod type for a Morphe model field
func getZodTypeForField(field yaml.ModelField, r *registry.Registry) (zoddef.ZodType, error) {
	// Check if it's an enum field
	if _, exists := r.GetAllEnums()[string(field.Type)]; exists {
		return zoddef.ZodEnumRefType{EnumName: string(field.Type)}, nil
	}

	// Otherwise, map the primitive type
	return typemap.MorpheFieldTypeToZodType(field.Type)
}

// getRelationshipFields generates fields for model relationships
func getRelationshipFields(related map[string]yaml.ModelRelation, r *registry.Registry) ([]zoddef.SchemaField, error) {
	fields := []zoddef.SchemaField{}

	relationNames := core.MapKeysSorted(related)
	for _, relationName := range relationNames {
		relation := related[relationName]

		// Determine the target model name
		targetModelName := relationName
		if relation.Aliased != "" {
			targetModelName = relation.Aliased
		}

		// Generate fields based on relationship type
		switch relation.Type {
		case "HasOne", "ForOne":
			// Add ID field
			fields = append(fields, zoddef.SchemaField{
				Name:     strcase.ToCamelCase(relationName) + "ID",
				ZodType:  zoddef.ZodTypeNumber,
				Optional: true,
			})
			// Add reference field
			fields = append(fields, zoddef.SchemaField{
				Name:     strcase.ToCamelCase(relationName),
				ZodType:  zoddef.ZodLazyType{TypeName: targetModelName},
				Optional: true,
			})

		case "HasMany", "ForMany":
			// Add IDs array field
			fields = append(fields, zoddef.SchemaField{
				Name:     strcase.ToCamelCase(relationName) + "IDs",
				ZodType:  zoddef.ZodArrayType{ElementType: zoddef.ZodTypeNumber},
				Optional: true,
			})
			// Add references array field (pluralize the field name)
			pluralName := strcase.ToCamelCase(relationName)
			if !strings.HasSuffix(pluralName, "s") {
				pluralName += "s"
			}
			fields = append(fields, zoddef.SchemaField{
				Name:     pluralName,
				ZodType:  zoddef.ZodArrayType{ElementType: zoddef.ZodLazyType{TypeName: targetModelName}},
				Optional: true,
			})

		case "HasOnePoly":
			// Polymorphic HasOne - just like HasOne but referencing the aliased type
			fields = append(fields, zoddef.SchemaField{
				Name:     strcase.ToCamelCase(relationName) + "ID",
				ZodType:  zoddef.ZodTypeNumber,
				Optional: true,
			})
			fields = append(fields, zoddef.SchemaField{
				Name:     strcase.ToCamelCase(relationName),
				ZodType:  zoddef.ZodLazyType{TypeName: targetModelName},
				Optional: true,
			})

		case "HasManyPoly":
			// Polymorphic HasMany
			fields = append(fields, zoddef.SchemaField{
				Name:     strcase.ToCamelCase(relationName) + "IDs",
				ZodType:  zoddef.ZodArrayType{ElementType: zoddef.ZodTypeNumber},
				Optional: true,
			})
			// Add references array field (pluralize the field name)
			pluralName := strcase.ToCamelCase(relationName)
			if !strings.HasSuffix(pluralName, "s") {
				pluralName += "s"
			}
			fields = append(fields, zoddef.SchemaField{
				Name:     pluralName,
				ZodType:  zoddef.ZodArrayType{ElementType: zoddef.ZodLazyType{TypeName: targetModelName}},
				Optional: true,
			})

		case "ForOnePoly", "ForManyPoly":
			// For polymorphic relationships, we include a type discriminator
			fields = append(fields, zoddef.SchemaField{
				Name:     strcase.ToCamelCase(relationName) + "Type",
				ZodType:  zoddef.ZodTypeString,
				Optional: true,
			})
			fields = append(fields, zoddef.SchemaField{
				Name:     strcase.ToCamelCase(relationName) + "ID",
				ZodType:  zoddef.ZodTypeNumber,
				Optional: true,
			})
		}
	}

	return fields, nil
}

// getIdentifierSchemas creates schemas for model identifiers
func getIdentifierSchemas(model yaml.Model, mainSchema *zoddef.Schema) ([]*zoddef.Schema, error) {
	schemas := []*zoddef.Schema{}

	identifierNames := core.MapKeysSorted(model.Identifiers)
	for _, identifierName := range identifierNames {
		identifier := model.Identifiers[identifierName]

		idSchema := &zoddef.Schema{
			Name:    fmt.Sprintf("%sID%s", model.Name, strcase.ToPascalCase(identifierName)),
			Imports: []zoddef.SchemaImport{},
			Fields:  []zoddef.SchemaField{},
		}

		// Add fields from the identifier
		for _, fieldName := range identifier.Fields {
			camelFieldName := strcase.ToCamelCase(fieldName)

			// Find the field in the main schema
			for _, field := range mainSchema.Fields {
				if field.Name == camelFieldName {
					idSchema.Fields = append(idSchema.Fields, field)
					break
				}
			}
		}

		if len(idSchema.Fields) > 0 {
			schemas = append(schemas, idSchema)
		}
	}

	return schemas, nil
}

// generateZodModelContent generates the TypeScript file content for model Zod schemas
func generateZodModelContent(modelName string, schemas []*zoddef.Schema, r *registry.Registry) []byte {
	cb := formatdef.NewContentBuilder("  ")

	// Collect all imports
	imports := collectImports(schemas, modelName, r)

	// Write imports
	cb.Line("import { z } from 'zod'")
	for _, imp := range imports {
		if len(imp.Names) > 0 {
			cb.Line("import { %s } from '%s'", joinStrings(imp.Names, ", "), imp.ImportPath)
		}
	}

	cb.EmptyLine()

	// Generate each schema
	for i, schema := range schemas {
		if i > 0 {
			cb.EmptyLine()
		}
		generateSchemaDefinition(cb, schema)
	}
	cb.EmptyLine()

	return cb.Build()
}

// collectImports gathers all necessary imports for the schemas
func collectImports(schemas []*zoddef.Schema, modelName string, r *registry.Registry) []zoddef.SchemaImport {
	importMap := make(map[string]map[string]bool) // path -> names

	for _, schema := range schemas {
		for _, field := range schema.Fields {
			addImportsForType(field.ZodType, modelName, importMap, r)
		}
	}

	// Convert map to slice and sort
	imports := []zoddef.SchemaImport{}
	paths := make([]string, 0, len(importMap))
	for path := range importMap {
		paths = append(paths, path)
	}

	// Sort paths (enums should come before models/schemas)
	sortImportPaths(paths)

	for _, path := range paths {
		names := importMap[path]
		nameSlice := make([]string, 0, len(names))
		for name := range names {
			nameSlice = append(nameSlice, name)
		}
		// Sort names alphabetically
		sortStrings(nameSlice)

		if len(nameSlice) > 0 {
			imports = append(imports, zoddef.SchemaImport{
				Names:      nameSlice,
				ImportPath: path,
			})
		}
	}

	return imports
}

// sortImportPaths sorts import paths with enums first, then models
func sortImportPaths(paths []string) {
	// Simple bubble sort - enums (../) come before models (./)
	for i := 0; i < len(paths); i++ {
		for j := i + 1; j < len(paths); j++ {
			if paths[j] < paths[i] {
				paths[i], paths[j] = paths[j], paths[i]
			}
		}
	}
}

// sortStrings sorts a slice of strings alphabetically
func sortStrings(strs []string) {
	for i := 0; i < len(strs); i++ {
		for j := i + 1; j < len(strs); j++ {
			if strs[j] < strs[i] {
				strs[i], strs[j] = strs[j], strs[i]
			}
		}
	}
}

// addImportsForType recursively adds imports for a Zod type
func addImportsForType(zodType zoddef.ZodType, currentModel string, importMap map[string]map[string]bool, r *registry.Registry) {
	switch t := zodType.(type) {
	case zoddef.ZodEnumRefType:
		// Import enum
		path := "../enums/" + toFileName(t.EnumName)
		if importMap[path] == nil {
			importMap[path] = make(map[string]bool)
		}
		// Add enum name first, then schema (alphabetical order)
		importMap[path][t.EnumName] = true
		importMap[path][t.EnumName+"Schema"] = true

	case zoddef.ZodSchemaRefType:
		// Import another schema if it's not the current one
		if t.SchemaName != currentModel {
			path := "./" + toFileName(t.SchemaName)
			if importMap[path] == nil {
				importMap[path] = make(map[string]bool)
			}
			importMap[path][t.SchemaName+"Schema"] = true
		}

	case zoddef.ZodLazyType:
		// Import the referenced type if it's not the current one
		if t.TypeName != currentModel {
			path := "./" + toFileName(t.TypeName)
			if importMap[path] == nil {
				importMap[path] = make(map[string]bool)
			}
			importMap[path][t.TypeName+"Schema"] = true
		}

	case zoddef.ZodArrayType:
		addImportsForType(t.ElementType, currentModel, importMap, r)
	}
}

// generateSchemaDefinition generates a single Zod schema definition
func generateSchemaDefinition(cb *formatdef.ContentBuilder, schema *zoddef.Schema) {
	cb.Line("export const %sSchema = z.object({", schema.Name)
	cb.Indent()

	for i, field := range schema.Fields {
		fieldDef := field.ZodType.GetZodTypeString()
		if field.Optional {
			fieldDef += ".optional()"
		}

		line := fmt.Sprintf("%s: %s", field.Name, fieldDef)
		if i < len(schema.Fields)-1 {
			line += ","
		}
		cb.Line(line)
	}

	cb.Dedent()
	cb.Line("})")
	cb.EmptyLine()
	cb.Line("export type %s = z.infer<typeof %sSchema>", schema.Name, schema.Name)
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
