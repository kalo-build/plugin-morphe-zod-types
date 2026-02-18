package compile

import (
	"fmt"
	"strings"

	"github.com/kalo-build/go-util/core"
	"github.com/kalo-build/go-util/strcase"
	"github.com/kalo-build/morphe-go/pkg/registry"
	"github.com/kalo-build/morphe-go/pkg/yaml"
	"github.com/kalo-build/morphe-go/pkg/yamlops"
	"github.com/kalo-build/plugin-morphe-zod-types/pkg/compile/cfg"
	"github.com/kalo-build/plugin-morphe-zod-types/pkg/formatdef"
	"github.com/kalo-build/plugin-morphe-zod-types/pkg/typemap"
	"github.com/kalo-build/plugin-morphe-zod-types/pkg/zoddef"
)

// CompileAllModels compiles all models to Zod schemas
func CompileAllModels(config MorpheCompileConfig, r *registry.Registry, writer *MorpheWriter) error {
	modelContents := make(map[string][]byte)
	fieldCasing := config.FormatConfig.FieldCasing

	for modelName, model := range r.GetAllModels() {
		// Compile the model to Zod schema
		zodSchemas, err := MorpheModelToZodSchemas(model, r, fieldCasing)
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
func MorpheModelToZodSchemas(model yaml.Model, r *registry.Registry, fieldCasing cfg.Casing) ([]*zoddef.Schema, error) {
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

		// Fields are required by default; only fields with the "optional" attribute are optional
		isOptional := false
		for _, attr := range field.Attributes {
			if attr == "optional" {
				isOptional = true
				break
			}
		}

		mainSchema.Fields = append(mainSchema.Fields, zoddef.SchemaField{
			Name:     fieldCasing.Apply(fieldName),
			ZodType:  zodType,
			Optional: isOptional,
		})
	}

	// Process relationships
	if len(model.Related) > 0 {
		relationFields, err := getRelationshipFields(model.Related, r, fieldCasing)
		if err != nil {
			return nil, err
		}
		mainSchema.Fields = append(mainSchema.Fields, relationFields...)
	}

	schemas = append(schemas, mainSchema)

	// Create identifier schemas
	identifierSchemas, err := getIdentifierSchemas(model, mainSchema, fieldCasing)
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
func getRelationshipFields(related map[string]yaml.ModelRelation, r *registry.Registry, fieldCasing cfg.Casing) ([]zoddef.SchemaField, error) {
	fields := []zoddef.SchemaField{}

	relationNames := core.MapKeysSorted(related)
	for _, relationName := range relationNames {
		relation := related[relationName]

		// Determine the target model name
		targetModelName := yamlops.GetRelationTargetName(relationName, relation.Aliased)

		// Get the base field name with casing applied
		baseName := fieldCasing.Apply(relationName)

		// Generate fields based on relationship type
		switch relation.Type {
		case "HasOne", "ForOne":
			fkZodType, err := lookupFKZodType(r, targetModelName)
			if err != nil {
				return nil, fmt.Errorf("relationship %q: %w", relationName, err)
			}
			fields = append(fields, zoddef.SchemaField{
				Name:     baseName + idSuffix(fieldCasing),
				ZodType:  fkZodType,
				Optional: true,
			})
			fields = append(fields, zoddef.SchemaField{
				Name:     baseName,
				ZodType:  zoddef.ZodLazyType{TypeName: targetModelName},
				Optional: true,
			})

		case "HasMany", "ForMany":
			fkZodType, err := lookupFKZodType(r, targetModelName)
			if err != nil {
				return nil, fmt.Errorf("relationship %q: %w", relationName, err)
			}
			fields = append(fields, zoddef.SchemaField{
				Name:     baseName + idsSuffix(fieldCasing),
				ZodType:  zoddef.ZodArrayType{ElementType: fkZodType},
				Optional: true,
			})
			pluralName := baseName
			if !strings.HasSuffix(pluralName, "s") {
				pluralName += "s"
			}
			fields = append(fields, zoddef.SchemaField{
				Name:     pluralName,
				ZodType:  zoddef.ZodArrayType{ElementType: zoddef.ZodLazyType{TypeName: targetModelName}},
				Optional: true,
			})

		case "HasOnePoly":
			fkZodType, err := lookupPolyFKZodType(r, relation.For)
			if err != nil {
				return nil, fmt.Errorf("relationship %q: %w", relationName, err)
			}
			fields = append(fields, zoddef.SchemaField{
				Name:     baseName + idSuffix(fieldCasing),
				ZodType:  fkZodType,
				Optional: true,
			})
			fields = append(fields, zoddef.SchemaField{
				Name:     baseName,
				ZodType:  zoddef.ZodLazyType{TypeName: targetModelName},
				Optional: true,
			})

		case "HasManyPoly":
			fkZodType, err := lookupPolyFKZodType(r, relation.For)
			if err != nil {
				return nil, fmt.Errorf("relationship %q: %w", relationName, err)
			}
			fields = append(fields, zoddef.SchemaField{
				Name:     baseName + idsSuffix(fieldCasing),
				ZodType:  zoddef.ZodArrayType{ElementType: fkZodType},
				Optional: true,
			})
			pluralName := baseName
			if !strings.HasSuffix(pluralName, "s") {
				pluralName += "s"
			}
			fields = append(fields, zoddef.SchemaField{
				Name:     pluralName,
				ZodType:  zoddef.ZodArrayType{ElementType: zoddef.ZodLazyType{TypeName: targetModelName}},
				Optional: true,
			})

		case "ForOnePoly", "ForManyPoly":
			fkZodType, err := lookupPolyFKZodType(r, relation.For)
			if err != nil {
				return nil, fmt.Errorf("relationship %q: %w", relationName, err)
			}
			fields = append(fields, zoddef.SchemaField{
				Name:     baseName + typeSuffix(fieldCasing),
				ZodType:  zoddef.ZodTypeString,
				Optional: true,
			})
			fields = append(fields, zoddef.SchemaField{
				Name:     baseName + idSuffix(fieldCasing),
				ZodType:  fkZodType,
				Optional: true,
			})
		}
	}

	return fields, nil
}

// lookupFKZodType resolves the Zod type for a FK ID field by looking up the
// target model's primary key field type in the registry.
func lookupFKZodType(r *registry.Registry, targetModelName string) (zoddef.ZodType, error) {
	targetModel, err := r.GetModel(targetModelName)
	if err != nil {
		return nil, fmt.Errorf("model with name '%s' not found in registry", targetModelName)
	}
	primaryIDFieldName, err := yamlops.GetModelPrimaryIdentifierFieldName(targetModel)
	if err != nil {
		return nil, err
	}
	primaryIDField, err := yamlops.GetModelFieldDefinitionByName(targetModel, primaryIDFieldName)
	if err != nil {
		return nil, err
	}
	return typemap.MorpheFieldTypeToZodType(primaryIDField.Type)
}

// lookupPolyFKZodType resolves the FK type for polymorphic relationships by
// using the first target in the For list (all targets should share the same
// primary key type convention).
func lookupPolyFKZodType(r *registry.Registry, forModels []string) (zoddef.ZodType, error) {
	return zoddef.ZodTypeString, nil
}

// idSuffix returns the appropriate suffix for ID fields based on casing
func idSuffix(casing cfg.Casing) string {
	switch casing {
	case cfg.CasingSnake:
		return "_id"
	default:
		return "ID"
	}
}

// idsSuffix returns the appropriate suffix for IDs array fields based on casing
func idsSuffix(casing cfg.Casing) string {
	switch casing {
	case cfg.CasingSnake:
		return "_ids"
	default:
		return "IDs"
	}
}

// typeSuffix returns the appropriate suffix for Type fields based on casing
func typeSuffix(casing cfg.Casing) string {
	switch casing {
	case cfg.CasingSnake:
		return "_type"
	default:
		return "Type"
	}
}

// getIdentifierSchemas creates schemas for model identifiers
func getIdentifierSchemas(model yaml.Model, mainSchema *zoddef.Schema, fieldCasing cfg.Casing) ([]*zoddef.Schema, error) {
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
			casedFieldName := fieldCasing.Apply(fieldName)

			// Find the field in the main schema
			for _, field := range mainSchema.Fields {
				if field.Name == casedFieldName {
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

// importKind tracks whether an import is a type-only or value import
type importKind int

const (
	importValue importKind = iota
	importType
)

// collectImports gathers all necessary imports for the schemas
// Returns a map of path -> (name -> isTypeOnly)
func collectImports(schemas []*zoddef.Schema, modelName string, r *registry.Registry) []zoddef.SchemaImport {
	importMap := make(map[string]map[string]importKind) // path -> name -> kind

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
		for name, kind := range names {
			if kind == importType {
				nameSlice = append(nameSlice, "type "+name)
			} else {
				nameSlice = append(nameSlice, name)
			}
		}
		// Sort names alphabetically (type imports will sort after due to "type " prefix)
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
func addImportsForType(zodType zoddef.ZodType, currentModel string, importMap map[string]map[string]importKind, r *registry.Registry) {
	switch t := zodType.(type) {
	case zoddef.ZodEnumRefType:
		// Import enum (type for interface, value for schema)
		path := "../enums/" + toFileName(t.EnumName)
		if importMap[path] == nil {
			importMap[path] = make(map[string]importKind)
		}
		importMap[path][t.EnumName] = importType
		importMap[path][t.EnumName+"Schema"] = importValue

	case zoddef.ZodSchemaRefType:
		// Import another schema if it's not the current one
		if t.SchemaName != currentModel {
			path := "./" + toFileName(t.SchemaName)
			if importMap[path] == nil {
				importMap[path] = make(map[string]importKind)
			}
			importMap[path][t.SchemaName+"Schema"] = importValue
		}

	case zoddef.ZodLazyType:
		// Import the referenced type if it's not the current one
		if t.TypeName != currentModel {
			path := "./" + toFileName(t.TypeName)
			if importMap[path] == nil {
				importMap[path] = make(map[string]importKind)
			}
			// Import type (for interface) and schema (for z.lazy)
			importMap[path][t.TypeName] = importType
			importMap[path][t.TypeName+"Schema"] = importValue
		}

	case zoddef.ZodArrayType:
		addImportsForType(t.ElementType, currentModel, importMap, r)
	}
}

// generateSchemaDefinition generates a single Zod schema definition
func generateSchemaDefinition(cb *formatdef.ContentBuilder, schema *zoddef.Schema) {
	// Check if schema has any lazy references (circular dependencies)
	hasLazyRefs := schemaHasLazyRefs(schema)

	if hasLazyRefs {
		// For schemas with circular refs, we need to define the interface first
		// to help TypeScript resolve the types
		cb.Line("export interface %s {", schema.Name)
		cb.Indent()
		for i, field := range schema.Fields {
			tsType := field.ZodType.GetTypeScriptType()
			if field.Optional {
				tsType += " | undefined"
			}
			line := fmt.Sprintf("%s%s: %s", field.Name, optionalMarker(field.Optional), tsType)
			if i < len(schema.Fields)-1 {
				line += ";"
			}
			cb.Line(line)
		}
		cb.Dedent()
		cb.Line("}")
		cb.EmptyLine()

		// Use explicit type annotation to break circular inference
		cb.Line("export const %sSchema: z.ZodType<%s> = z.object({", schema.Name, schema.Name)
	} else {
		cb.Line("export const %sSchema = z.object({", schema.Name)
	}

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

	if !hasLazyRefs {
		// Only generate inferred type if we didn't define interface above
		cb.EmptyLine()
		cb.Line("export type %s = z.infer<typeof %sSchema>", schema.Name, schema.Name)
	}
}

// schemaHasLazyRefs checks if a schema has any z.lazy() references
func schemaHasLazyRefs(schema *zoddef.Schema) bool {
	for _, field := range schema.Fields {
		if hasLazyType(field.ZodType) {
			return true
		}
	}
	return false
}

// hasLazyType recursively checks if a type contains z.lazy()
func hasLazyType(t zoddef.ZodType) bool {
	switch v := t.(type) {
	case zoddef.ZodLazyType:
		return true
	case zoddef.ZodArrayType:
		return hasLazyType(v.ElementType)
	default:
		return false
	}
}

// optionalMarker returns "?" if the field is optional
func optionalMarker(optional bool) string {
	if optional {
		return "?"
	}
	return ""
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
