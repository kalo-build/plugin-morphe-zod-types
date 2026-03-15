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

// CompileAllEntities compiles all entities to Zod schemas
func CompileAllEntities(config MorpheCompileConfig, r *registry.Registry, writer *MorpheWriter) error {
	entityContents := make(map[string][]byte)
	fieldCasing := config.FormatConfig.FieldCasing

	for entityName, entity := range r.GetAllEntities() {
		// Compile the entity to Zod schemas
		zodSchemas, err := MorpheEntityToZodSchemas(entity, r, fieldCasing)
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
func MorpheEntityToZodSchemas(entity yaml.Entity, r *registry.Registry, fieldCasing cfg.Casing) ([]*zoddef.Schema, error) {
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

	// Process entity relationships
	if len(entity.Related) > 0 {
		relationFields, err := getEntityRelationshipFields(entity.Related, r, fieldCasing)
		if err != nil {
			return nil, err
		}
		mainSchema.Fields = append(mainSchema.Fields, relationFields...)
	}

	schemas = append(schemas, mainSchema)

	// Create identifier schemas for entities
	identifierSchemas, err := getEntityIdentifierSchemas(entity, mainSchema, fieldCasing)
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

func hasAttribute(attributes []string, attr string) bool {
	for _, a := range attributes {
		if a == attr {
			return true
		}
	}
	return false
}

// lookupEntityFKZodType resolves the Zod type for an entity FK field by
// resolving the target entity's primary identifier through its indirected model path.
func lookupEntityFKZodType(r *registry.Registry, targetEntityName string) (zoddef.ZodType, error) {
	targetEntity, err := r.GetEntity(targetEntityName)
	if err != nil {
		return nil, fmt.Errorf("entity with name '%s' not found in registry", targetEntityName)
	}
	primaryIDFieldName, err := yamlops.GetEntityPrimaryIdentifierFieldName(targetEntity)
	if err != nil {
		return nil, err
	}
	primaryIDField, exists := targetEntity.Fields[primaryIDFieldName]
	if !exists {
		return nil, fmt.Errorf("primary identifier field '%s' not found on entity '%s'", primaryIDFieldName, targetEntityName)
	}
	return getZodTypeForEntityField(primaryIDField, r)
}

// getEntityRelationshipFields generates fields for entity relationships
func getEntityRelationshipFields(related map[string]yaml.EntityRelation, r *registry.Registry, fieldCasing cfg.Casing) ([]zoddef.SchemaField, error) {
	fields := []zoddef.SchemaField{}

	relationNames := core.MapKeysSorted(related)
	for _, relationName := range relationNames {
		relation := related[relationName]

		targetEntityName := relationName
		if relation.Aliased != "" {
			targetEntityName = relation.Aliased
		}

		baseName := fieldCasing.Apply(relationName)
		isOptional := hasAttribute(relation.Attributes, "optional")

		switch relation.Type {
		case "HasOne", "ForOne":
			fkZodType, err := lookupEntityFKZodType(r, targetEntityName)
			if err != nil {
				return nil, fmt.Errorf("entity relationship %q: %w", relationName, err)
			}
			fields = append(fields, zoddef.SchemaField{
				Name:     baseName + idSuffix(fieldCasing),
				ZodType:  fkZodType,
				Optional: isOptional,
			})
			fields = append(fields, zoddef.SchemaField{
				Name:     baseName,
				ZodType:  zoddef.ZodLazyType{TypeName: targetEntityName},
				Optional: true,
			})

		case "HasMany", "ForMany":
			fkZodType, err := lookupEntityFKZodType(r, targetEntityName)
			if err != nil {
				return nil, fmt.Errorf("entity relationship %q: %w", relationName, err)
			}
			fields = append(fields, zoddef.SchemaField{
				Name:     baseName + idsSuffix(fieldCasing),
				ZodType:  zoddef.ZodArrayType{ElementType: fkZodType},
				Optional: isOptional,
			})
			pluralName := baseName
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
				Name:     baseName + idSuffix(fieldCasing),
				ZodType:  zoddef.ZodTypeString,
				Optional: isOptional,
			})
			fields = append(fields, zoddef.SchemaField{
				Name:     baseName,
				ZodType:  zoddef.ZodLazyType{TypeName: targetEntityName},
				Optional: true,
			})

		case "HasManyPoly":
			fields = append(fields, zoddef.SchemaField{
				Name:     baseName + idsSuffix(fieldCasing),
				ZodType:  zoddef.ZodArrayType{ElementType: zoddef.ZodTypeString},
				Optional: isOptional,
			})
			pluralName := baseName
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
				Name:     baseName + typeSuffix(fieldCasing),
				ZodType:  zoddef.ZodTypeString,
				Optional: isOptional,
			})
			fields = append(fields, zoddef.SchemaField{
				Name:     baseName + idSuffix(fieldCasing),
				ZodType:  zoddef.ZodTypeString,
				Optional: isOptional,
			})
		}
	}

	return fields, nil
}

// getEntityIdentifierSchemas creates schemas for entity identifiers
func getEntityIdentifierSchemas(entity yaml.Entity, mainSchema *zoddef.Schema, fieldCasing cfg.Casing) ([]*zoddef.Schema, error) {
	schemas := []*zoddef.Schema{}

	identifierNames := core.MapKeysSorted(entity.Identifiers)
	for _, identifierName := range identifierNames {
		identifier := entity.Identifiers[identifierName]

		idSchema := &zoddef.Schema{
			Name:    fmt.Sprintf("%sID%s", entity.Name, strcase.ToPascalCase(identifierName)),
			Imports: []zoddef.SchemaImport{},
			Fields:  []zoddef.SchemaField{},
		}

		for _, fieldName := range identifier.Fields {
			targetFieldNames := resolveEntityIdentifierFieldNames(fieldName, entity.Related, fieldCasing)

			for _, targetName := range targetFieldNames {
				for _, field := range mainSchema.Fields {
					if field.Name == targetName {
						idSchema.Fields = append(idSchema.Fields, field)
						break
					}
				}
			}
		}

		if len(idSchema.Fields) > 0 {
			schemas = append(schemas, idSchema)
		}
	}

	return schemas, nil
}

func resolveEntityIdentifierFieldNames(fieldName string, related map[string]yaml.EntityRelation, fieldCasing cfg.Casing) []string {
	if !strings.HasPrefix(fieldName, "rel:") {
		return []string{fieldCasing.Apply(fieldName)}
	}

	relationName := strings.TrimPrefix(fieldName, "rel:")
	baseName := fieldCasing.Apply(relationName)

	relation, exists := related[relationName]
	if !exists {
		return nil
	}

	relationType := string(relation.Type)
	if yamlops.IsRelationPolyFor(relationType) {
		return []string{
			baseName + typeSuffix(fieldCasing),
			baseName + idSuffix(fieldCasing),
		}
	}

	return []string{baseName + idSuffix(fieldCasing)}
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
