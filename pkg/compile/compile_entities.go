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

// CompileAllEntities compiles all entities to Zod schemas
func CompileAllEntities(config MorpheCompileConfig, r *registry.Registry, writer *MorpheWriter) error {
	entityContents := make(map[string][]byte)

	for entityName, entity := range r.GetAllEntities() {
		// Compile the entity to Zod schemas
		zodSchemas, err := MorpheEntityToZodSchemas(entity, r)
		if err != nil {
			return fmt.Errorf("failed to compile entity %s: %w", entityName, err)
		}

		// Generate the TypeScript file content
		content := generateZodEntityContent(entity.Name, zodSchemas, r)
		entityContents[entityName] = content
	}

	// Write all entity contents
	return writer.WriteAllEntities(entityContents)
}

// MorpheEntityToZodSchemas converts a Morphe entity to Zod schema definitions
func MorpheEntityToZodSchemas(entity yaml.Entity, r *registry.Registry) ([]*zoddef.Schema, error) {
	schemas := []*zoddef.Schema{}

	// Create the main entity schema
	mainSchema := &zoddef.Schema{
		Name:    entity.Name,
		Imports: []zoddef.SchemaImport{},
		Fields:  []zoddef.SchemaField{},
	}

	// Process entity fields (indirected from models)
	fieldNames := core.MapKeysSorted(entity.Fields)
	for _, fieldName := range fieldNames {
		field := entity.Fields[fieldName]

		zodType, err := getZodTypeForEntityField(field, r)
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

	// Process entity relationships
	if len(entity.Related) > 0 {
		relationFields, err := getEntityRelationshipFields(entity.Related, r)
		if err != nil {
			return nil, err
		}
		mainSchema.Fields = append(mainSchema.Fields, relationFields...)
	}

	schemas = append(schemas, mainSchema)

	// Create identifier schemas for entities
	identifierSchemas, err := getEntityIdentifierSchemas(entity, mainSchema)
	if err != nil {
		return nil, err
	}
	schemas = append(schemas, identifierSchemas...)

	return schemas, nil
}

// getZodTypeForEntityField returns the Zod type for an entity field (which is indirected)
func getZodTypeForEntityField(field yaml.EntityField, r *registry.Registry) (zoddef.ZodType, error) {
	// Parse the indirected type path (e.g., "Person.FirstName" or "Person.ContactInfo.Email")
	parts := strings.Split(string(field.Type), ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid entity field type: %s", field.Type)
	}

	// Get the root model
	rootModelName := parts[0]
	rootModel, exists := r.GetAllModels()[rootModelName]
	if !exists {
		return nil, fmt.Errorf("model not found: %s", rootModelName)
	}

	// Traverse the path to find the final field
	currentModel := rootModel
	for i := 1; i < len(parts); i++ {
		fieldName := parts[i]

		// Check if this is the last part (the actual field)
		if i == len(parts)-1 {
			// This is the final field
			if modelField, exists := currentModel.Fields[fieldName]; exists {
				// Check if it's an enum
				if _, isEnum := r.GetAllEnums()[string(modelField.Type)]; isEnum {
					return zoddef.ZodEnumRefType{EnumName: string(modelField.Type)}, nil
				}
				// Map the primitive type
				return typemap.MorpheFieldTypeToZodType(modelField.Type)
			}
			return nil, fmt.Errorf("field not found: %s in model %s", fieldName, currentModel.Name)
		}

		// This is a relationship, traverse to the related model
		if relation, exists := currentModel.Related[fieldName]; exists {
			nextModelName := fieldName
			if relation.Aliased != "" {
				nextModelName = relation.Aliased
			}

			nextModel, exists := r.GetAllModels()[nextModelName]
			if !exists {
				return nil, fmt.Errorf("related model not found: %s", nextModelName)
			}
			currentModel = nextModel
		} else {
			return nil, fmt.Errorf("relation not found: %s in model %s", fieldName, currentModel.Name)
		}
	}

	return nil, fmt.Errorf("could not resolve entity field type: %s", field.Type)
}

// getEntityRelationshipFields generates fields for entity relationships
func getEntityRelationshipFields(related map[string]yaml.EntityRelation, r *registry.Registry) ([]zoddef.SchemaField, error) {
	fields := []zoddef.SchemaField{}

	relationNames := core.MapKeysSorted(related)
	for _, relationName := range relationNames {
		relation := related[relationName]

		// Determine the target entity name
		targetEntityName := relationName
		if relation.Aliased != "" {
			targetEntityName = relation.Aliased
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
				ZodType:  zoddef.ZodLazyType{TypeName: targetEntityName},
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
				ZodType:  zoddef.ZodArrayType{ElementType: zoddef.ZodLazyType{TypeName: targetEntityName}},
				Optional: true,
			})

		case "HasOnePoly":
			fields = append(fields, zoddef.SchemaField{
				Name:     strcase.ToCamelCase(relationName) + "ID",
				ZodType:  zoddef.ZodTypeNumber,
				Optional: true,
			})
			fields = append(fields, zoddef.SchemaField{
				Name:     strcase.ToCamelCase(relationName),
				ZodType:  zoddef.ZodLazyType{TypeName: targetEntityName},
				Optional: true,
			})

		case "HasManyPoly":
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
				ZodType:  zoddef.ZodArrayType{ElementType: zoddef.ZodLazyType{TypeName: targetEntityName}},
				Optional: true,
			})

		case "ForOnePoly", "ForManyPoly":
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

// getEntityIdentifierSchemas creates schemas for entity identifiers
func getEntityIdentifierSchemas(entity yaml.Entity, mainSchema *zoddef.Schema) ([]*zoddef.Schema, error) {
	schemas := []*zoddef.Schema{}

	identifierNames := core.MapKeysSorted(entity.Identifiers)
	for _, identifierName := range identifierNames {
		identifier := entity.Identifiers[identifierName]

		idSchema := &zoddef.Schema{
			Name:    fmt.Sprintf("%sID%s", entity.Name, strcase.ToPascalCase(identifierName)),
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

// generateZodEntityContent generates the TypeScript file content for entity Zod schemas
func generateZodEntityContent(entityName string, schemas []*zoddef.Schema, r *registry.Registry) []byte {
	cb := formatdef.NewContentBuilder("  ")

	// Collect all imports
	imports := collectImports(schemas, entityName, r)

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
