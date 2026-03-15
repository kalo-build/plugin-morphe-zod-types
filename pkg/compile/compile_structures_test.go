package compile_test

import (
	"testing"

	"github.com/kalo-build/morphe-go/pkg/registry"
	"github.com/kalo-build/morphe-go/pkg/yaml"
	"github.com/kalo-build/plugin-morphe-zod-types/pkg/compile"
	"github.com/kalo-build/plugin-morphe-zod-types/pkg/compile/cfg"
	"github.com/kalo-build/plugin-morphe-zod-types/pkg/zoddef"
	"github.com/stretchr/testify/suite"
)

type CompileStructuresTestSuite struct {
	suite.Suite
}

func TestCompileStructuresTestSuite(t *testing.T) {
	suite.Run(t, new(CompileStructuresTestSuite))
}

func (suite *CompileStructuresTestSuite) TestMorpheStructureToZodSchema_BasicFields() {
	structure := yaml.Structure{
		Name: "Address",
		Fields: map[string]yaml.StructureField{
			"Street":  {Type: yaml.StructureFieldTypeString},
			"City":    {Type: yaml.StructureFieldTypeString},
			"ZipCode": {Type: yaml.StructureFieldTypeString},
		},
	}

	r := registry.NewRegistry()

	schema, err := compile.MorpheStructureToZodSchema(structure, r, cfg.CasingCamel)

	suite.NoError(err)
	suite.NotNil(schema)
	suite.Equal("Address", schema.Name)
	suite.Len(schema.Fields, 3)

	// Structure fields without "optional" attribute should be required
	for _, field := range schema.Fields {
		suite.False(field.Optional, "Structure field %s should be required by default", field.Name)
	}
}

func (suite *CompileStructuresTestSuite) TestMorpheStructureToZodSchema_AllFieldTypes() {
	structure := yaml.Structure{
		Name: "AllTypes",
		Fields: map[string]yaml.StructureField{
			"UUID":          {Type: yaml.StructureFieldTypeUUID},
			"AutoIncrement": {Type: yaml.StructureFieldTypeAutoIncrement},
			"String":        {Type: yaml.StructureFieldTypeString},
			"Integer":       {Type: yaml.StructureFieldTypeInteger},
			"Float":         {Type: yaml.StructureFieldTypeFloat},
			"Boolean":       {Type: yaml.StructureFieldTypeBoolean},
			"Time":          {Type: yaml.StructureFieldTypeTime},
			"Date":          {Type: yaml.StructureFieldTypeDate},
		},
	}

	r := registry.NewRegistry()

	schema, err := compile.MorpheStructureToZodSchema(structure, r, cfg.CasingCamel)

	suite.NoError(err)
	suite.Len(schema.Fields, 8)

	// Verify field type mappings - all required by default (no optional attribute)
	for _, field := range schema.Fields {
		suite.NotNil(field.ZodType)
		suite.False(field.Optional, "Structure field %s should be required by default", field.Name)
	}
}

func (suite *CompileStructuresTestSuite) TestMorpheStructureToZodSchema_EnumField() {
	structure := yaml.Structure{
		Name: "Preferences",
		Fields: map[string]yaml.StructureField{
			"Theme": {Type: "ColorTheme"},
		},
	}

	r := registry.NewRegistry()
	r.SetEnum("ColorTheme", yaml.Enum{
		Name: "ColorTheme",
		Type: yaml.EnumTypeString,
		Entries: map[string]any{
			"Light": "light",
			"Dark":  "dark",
		},
	})

	schema, err := compile.MorpheStructureToZodSchema(structure, r, cfg.CasingCamel)

	suite.NoError(err)
	suite.Len(schema.Fields, 1)

	// Check enum reference
	themeField := schema.Fields[0]
	suite.Equal("theme", themeField.Name)

	enumRef, ok := themeField.ZodType.(zoddef.ZodEnumRefType)
	suite.True(ok)
	suite.Equal("ColorTheme", enumRef.EnumName)
}

func (suite *CompileStructuresTestSuite) TestMorpheStructureToZodSchema_OptionalAttribute() {
	structure := yaml.Structure{
		Name: "UserResponse",
		Fields: map[string]yaml.StructureField{
			"ID":             {Type: yaml.StructureFieldTypeUUID},
			"Email":          {Type: yaml.StructureFieldTypeString},
			"OrganizationID": {Type: yaml.StructureFieldTypeUUID, Attributes: []string{"optional"}},
		},
	}

	r := registry.NewRegistry()

	schema, err := compile.MorpheStructureToZodSchema(structure, r, cfg.CasingCamel)

	suite.NoError(err)
	suite.Len(schema.Fields, 3)

	// Find each field and check optionality
	for _, field := range schema.Fields {
		if field.Name == "organizationID" {
			suite.True(field.Optional, "OrganizationID should be optional")
		} else {
			suite.False(field.Optional, "Field %s should be required", field.Name)
		}
	}
}

func (suite *CompileStructuresTestSuite) TestMorpheStructureToZodSchema_EmptyStructure() {
	structure := yaml.Structure{
		Name:   "Empty",
		Fields: map[string]yaml.StructureField{},
	}

	r := registry.NewRegistry()

	schema, err := compile.MorpheStructureToZodSchema(structure, r, cfg.CasingCamel)

	suite.NoError(err)
	suite.NotNil(schema)
	suite.Len(schema.Fields, 0)
}

func (suite *CompileStructuresTestSuite) TestMorpheStructureToZodSchema_SingleField() {
	structure := yaml.Structure{
		Name: "Minimal",
		Fields: map[string]yaml.StructureField{
			"Value": {Type: yaml.StructureFieldTypeString},
		},
	}

	r := registry.NewRegistry()

	schema, err := compile.MorpheStructureToZodSchema(structure, r, cfg.CasingCamel)

	suite.NoError(err)
	suite.Len(schema.Fields, 1)
	suite.Equal("value", schema.Fields[0].Name)
	suite.Equal(zoddef.ZodTypeString, schema.Fields[0].ZodType)
}

func (suite *CompileStructuresTestSuite) TestMorpheStructureToZodSchema_StructureComposition() {
	lineItemStructure := yaml.Structure{
		Name: "InvoiceLineItem",
		Fields: map[string]yaml.StructureField{
			"Amount": {Type: yaml.StructureFieldTypeInteger},
		},
	}
	invoiceStructure := yaml.Structure{
		Name: "Invoice",
		Fields: map[string]yaml.StructureField{
			"ID":        {Type: yaml.StructureFieldTypeString},
			"LineItem": {Type: "InvoiceLineItem", Attributes: []string{"optional"}},
		},
	}

	r := registry.NewRegistry()
	r.SetStructure("InvoiceLineItem", lineItemStructure)
	r.SetStructure("Invoice", invoiceStructure)

	schema, err := compile.MorpheStructureToZodSchema(invoiceStructure, r, cfg.CasingCamel)

	suite.NoError(err)
	suite.NotNil(schema)
	suite.Equal("Invoice", schema.Name)
	suite.Len(schema.Fields, 2)

	suite.Equal("id", schema.Fields[0].Name)
	suite.Equal(zoddef.ZodTypeString, schema.Fields[0].ZodType)
	suite.False(schema.Fields[0].Optional)

	suite.Equal("lineItem", schema.Fields[1].Name)
	suite.True(schema.Fields[1].Optional)
	schemaRef, ok := schema.Fields[1].ZodType.(zoddef.ZodSchemaRefType)
	suite.True(ok, "LineItem should be ZodSchemaRefType")
	suite.Equal("InvoiceLineItem", schemaRef.SchemaName)
}
