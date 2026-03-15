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

type CompileModelsTestSuite struct {
	suite.Suite
}

func TestCompileModelsTestSuite(t *testing.T) {
	suite.Run(t, new(CompileModelsTestSuite))
}

func (suite *CompileModelsTestSuite) TestMorpheModelToZodSchemas_BasicFields() {
	model := yaml.Model{
		Name: "User",
		Fields: map[string]yaml.ModelField{
			"ID": {
				Type: yaml.ModelFieldTypeAutoIncrement,
			},
			"Name": {
				Type: yaml.ModelFieldTypeString,
			},
			"Email": {
				Type:       yaml.ModelFieldTypeString,
				Attributes: []string{"optional"},
			},
		},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {
				Fields: []string{"ID"},
			},
		},
		Related: map[string]yaml.ModelRelation{},
	}

	r := registry.NewRegistry()

	schemas, err := compile.MorpheModelToZodSchemas(model, r, cfg.CasingCamel)

	suite.NoError(err)
	suite.NotNil(schemas)
	suite.Len(schemas, 2) // Main schema + primary identifier

	// Check main schema
	mainSchema := schemas[0]
	suite.Equal("User", mainSchema.Name)
	suite.Len(mainSchema.Fields, 3)

	// Fields without "optional" attribute should be required
	idField := mainSchema.Fields[1]
	suite.Equal("id", idField.Name)
	suite.False(idField.Optional)

	nameField := mainSchema.Fields[2]
	suite.Equal("name", nameField.Name)
	suite.False(nameField.Optional)

	// Email with "optional" attribute should be optional
	emailField := mainSchema.Fields[0]
	suite.Equal("email", emailField.Name)
	suite.True(emailField.Optional)
}

func (suite *CompileModelsTestSuite) TestMorpheModelToZodSchemas_AllFieldTypes() {
	model := yaml.Model{
		Name: "AllTypes",
		Fields: map[string]yaml.ModelField{
			"UUID":          {Type: yaml.ModelFieldTypeUUID},
			"AutoIncrement": {Type: yaml.ModelFieldTypeAutoIncrement},
			"String":        {Type: yaml.ModelFieldTypeString},
			"Integer":       {Type: yaml.ModelFieldTypeInteger},
			"Float":         {Type: yaml.ModelFieldTypeFloat},
			"Boolean":       {Type: yaml.ModelFieldTypeBoolean},
			"Time":          {Type: yaml.ModelFieldTypeTime},
			"Date":          {Type: yaml.ModelFieldTypeDate},
			"Protected":     {Type: yaml.ModelFieldTypeProtected},
			"Sealed":        {Type: yaml.ModelFieldTypeSealed},
		},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"AutoIncrement"}},
		},
		Related: map[string]yaml.ModelRelation{},
	}

	r := registry.NewRegistry()

	schemas, err := compile.MorpheModelToZodSchemas(model, r, cfg.CasingCamel)

	suite.NoError(err)
	suite.Len(schemas, 2)
	suite.Len(schemas[0].Fields, 10)
}

func (suite *CompileModelsTestSuite) TestMorpheModelToZodSchemas_EnumField() {
	model := yaml.Model{
		Name: "Person",
		Fields: map[string]yaml.ModelField{
			"ID": {
				Type: yaml.ModelFieldTypeAutoIncrement,
			},
			"Nationality": {
				Type: "Nationality",
			},
		},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.ModelRelation{},
	}

	r := registry.NewRegistry()
	r.SetEnum("Nationality", yaml.Enum{
		Name: "Nationality",
		Type: yaml.EnumTypeString,
		Entries: map[string]any{
			"US": "American",
			"DE": "German",
		},
	})

	schemas, err := compile.MorpheModelToZodSchemas(model, r, cfg.CasingCamel)

	suite.NoError(err)
	suite.Len(schemas, 2)

	// Check that Nationality field has enum type
	nationalityField := schemas[0].Fields[1]
	suite.Equal("nationality", nationalityField.Name)

	enumRef, ok := nationalityField.ZodType.(zoddef.ZodEnumRefType)
	suite.True(ok)
	suite.Equal("Nationality", enumRef.EnumName)
}

func (suite *CompileModelsTestSuite) TestMorpheModelToZodSchemas_HasOne() {
	userModel := yaml.Model{
		Name: "User",
		Fields: map[string]yaml.ModelField{
			"ID": {Type: yaml.ModelFieldTypeAutoIncrement},
		},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.ModelRelation{
			"Profile": {Type: "HasOne"},
		},
	}

	r := registry.NewRegistry()
	r.SetModel("Profile", yaml.Model{
		Name:   "Profile",
		Fields: map[string]yaml.ModelField{"ID": {Type: yaml.ModelFieldTypeAutoIncrement}},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
	})

	schemas, err := compile.MorpheModelToZodSchemas(userModel, r, cfg.CasingCamel)

	suite.NoError(err)
	suite.Len(schemas[0].Fields, 3) // ID + ProfileID + Profile

	// Check ProfileID field
	profileIDField := schemas[0].Fields[1]
	suite.Equal("profileID", profileIDField.Name)
	suite.False(profileIDField.Optional)
	suite.Equal(zoddef.ZodTypeNumber, profileIDField.ZodType)

	// Check Profile field
	profileField := schemas[0].Fields[2]
	suite.Equal("profile", profileField.Name)
	suite.True(profileField.Optional)

	lazyType, ok := profileField.ZodType.(zoddef.ZodLazyType)
	suite.True(ok)
	suite.Equal("Profile", lazyType.TypeName)
}

func (suite *CompileModelsTestSuite) TestMorpheModelToZodSchemas_HasMany() {
	companyModel := yaml.Model{
		Name: "Company",
		Fields: map[string]yaml.ModelField{
			"ID": {Type: yaml.ModelFieldTypeAutoIncrement},
		},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.ModelRelation{
			"Employee": {Type: "HasMany"},
		},
	}

	r := registry.NewRegistry()
	r.SetModel("Employee", yaml.Model{
		Name:   "Employee",
		Fields: map[string]yaml.ModelField{"ID": {Type: yaml.ModelFieldTypeAutoIncrement}},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
	})

	schemas, err := compile.MorpheModelToZodSchemas(companyModel, r, cfg.CasingCamel)

	suite.NoError(err)
	suite.Len(schemas[0].Fields, 3) // ID + EmployeeIDs + Employees

	// Check EmployeeIDs field
	employeeIDsField := schemas[0].Fields[1]
	suite.Equal("employeeIDs", employeeIDsField.Name)
	suite.False(employeeIDsField.Optional)

	arrayType, ok := employeeIDsField.ZodType.(zoddef.ZodArrayType)
	suite.True(ok)
	suite.Equal(zoddef.ZodTypeNumber, arrayType.ElementType)

	// Check Employees field (pluralized)
	employeesField := schemas[0].Fields[2]
	suite.Equal("employees", employeesField.Name)
	suite.True(employeesField.Optional)

	arrayType2, ok := employeesField.ZodType.(zoddef.ZodArrayType)
	suite.True(ok)

	lazyType, ok := arrayType2.ElementType.(zoddef.ZodLazyType)
	suite.True(ok)
	suite.Equal("Employee", lazyType.TypeName)
}

func (suite *CompileModelsTestSuite) TestMorpheModelToZodSchemas_AliasedRelationship() {
	personModel := yaml.Model{
		Name: "Person",
		Fields: map[string]yaml.ModelField{
			"ID": {Type: yaml.ModelFieldTypeAutoIncrement},
		},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.ModelRelation{
			"WorkContact": {
				Type:    "ForOne",
				Aliased: "Contact",
			},
			"PersonalContact": {
				Type:    "ForOne",
				Aliased: "Contact",
			},
		},
	}

	r := registry.NewRegistry()
	r.SetModel("Contact", yaml.Model{
		Name:   "Contact",
		Fields: map[string]yaml.ModelField{"ID": {Type: yaml.ModelFieldTypeAutoIncrement}},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
	})

	schemas, err := compile.MorpheModelToZodSchemas(personModel, r, cfg.CasingCamel)

	suite.NoError(err)
	suite.Len(schemas[0].Fields, 5) // ID + WorkContactID + WorkContact + PersonalContactID + PersonalContact

	// Check WorkContact uses Contact as target
	workContactField := schemas[0].Fields[4]
	suite.Equal("workContact", workContactField.Name)

	lazyType, ok := workContactField.ZodType.(zoddef.ZodLazyType)
	suite.True(ok)
	suite.Equal("Contact", lazyType.TypeName) // Should reference Contact, not WorkContact
}

func (suite *CompileModelsTestSuite) TestMorpheModelToZodSchemas_PolymorphicHasOnePoly() {
	personModel := yaml.Model{
		Name: "Person",
		Fields: map[string]yaml.ModelField{
			"ID": {Type: yaml.ModelFieldTypeAutoIncrement},
		},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.ModelRelation{
			"Note": {
				Type:    "HasOnePoly",
				Through: "Commentable",
				Aliased: "Comment",
			},
		},
	}

	r := registry.NewRegistry()
	r.SetModel("Comment", yaml.Model{
		Name:   "Comment",
		Fields: map[string]yaml.ModelField{"ID": {Type: yaml.ModelFieldTypeAutoIncrement}},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
	})

	schemas, err := compile.MorpheModelToZodSchemas(personModel, r, cfg.CasingCamel)

	suite.NoError(err)
	suite.Len(schemas[0].Fields, 3) // ID + NoteID + Note

	// Check Note field references Comment
	noteField := schemas[0].Fields[2]
	suite.Equal("note", noteField.Name)

	lazyType, ok := noteField.ZodType.(zoddef.ZodLazyType)
	suite.True(ok)
	suite.Equal("Comment", lazyType.TypeName)
}

func (suite *CompileModelsTestSuite) TestMorpheModelToZodSchemas_PolymorphicForOnePoly() {
	commentModel := yaml.Model{
		Name: "Comment",
		Fields: map[string]yaml.ModelField{
			"ID": {Type: yaml.ModelFieldTypeAutoIncrement},
		},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.ModelRelation{
			"Commentable": {
				Type: "ForOnePoly",
				For:  []string{"Person", "Company"},
			},
		},
	}

	r := registry.NewRegistry()
	r.SetModel("Person", yaml.Model{
		Name:   "Person",
		Fields: map[string]yaml.ModelField{"ID": {Type: yaml.ModelFieldTypeAutoIncrement}},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
	})
	r.SetModel("Company", yaml.Model{
		Name:   "Company",
		Fields: map[string]yaml.ModelField{"ID": {Type: yaml.ModelFieldTypeAutoIncrement}},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
	})

	schemas, err := compile.MorpheModelToZodSchemas(commentModel, r, cfg.CasingCamel)

	suite.NoError(err)
	suite.Len(schemas[0].Fields, 3) // ID + CommentableType + CommentableID

	// Check CommentableType field
	typeField := schemas[0].Fields[1]
	suite.Equal("commentableType", typeField.Name)
	suite.Equal(zoddef.ZodTypeString, typeField.ZodType)
	suite.False(typeField.Optional)

	// Polymorphic FK is always z.string() regardless of target PK types
	idField := schemas[0].Fields[2]
	suite.Equal("commentableID", idField.Name)
	suite.Equal(zoddef.ZodTypeString, idField.ZodType)
	suite.False(idField.Optional)
}

func (suite *CompileModelsTestSuite) TestMorpheModelToZodSchemas_MultipleIdentifiers() {
	model := yaml.Model{
		Name: "Person",
		Fields: map[string]yaml.ModelField{
			"ID":        {Type: yaml.ModelFieldTypeAutoIncrement},
			"FirstName": {Type: yaml.ModelFieldTypeString},
			"LastName":  {Type: yaml.ModelFieldTypeString},
			"Email":     {Type: yaml.ModelFieldTypeString},
		},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
			"name":    {Fields: []string{"FirstName", "LastName"}},
			"email":   {Fields: []string{"Email"}},
		},
		Related: map[string]yaml.ModelRelation{},
	}

	r := registry.NewRegistry()

	schemas, err := compile.MorpheModelToZodSchemas(model, r, cfg.CasingCamel)

	suite.NoError(err)
	suite.Len(schemas, 4) // Main + 3 identifiers

	suite.Equal("Person", schemas[0].Name)
	suite.Equal("PersonIDEmail", schemas[1].Name)
	suite.Equal("PersonIDName", schemas[2].Name)
	suite.Equal("PersonIDPrimary", schemas[3].Name)

	// Check name identifier has both fields
	nameSchema := schemas[2]
	suite.Len(nameSchema.Fields, 2)
}

func (suite *CompileModelsTestSuite) TestMorpheModelToZodSchemas_NoRelationships() {
	model := yaml.Model{
		Name: "Simple",
		Fields: map[string]yaml.ModelField{
			"ID":   {Type: yaml.ModelFieldTypeAutoIncrement},
			"Name": {Type: yaml.ModelFieldTypeString},
		},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.ModelRelation{},
	}

	r := registry.NewRegistry()

	schemas, err := compile.MorpheModelToZodSchemas(model, r, cfg.CasingCamel)

	suite.NoError(err)
	suite.Len(schemas, 2)
	suite.Len(schemas[0].Fields, 2) // Only ID and Name, no relationship fields
}

func (suite *CompileModelsTestSuite) TestMorpheModelToZodSchemas_AllFieldsRequiredByDefault() {
	model := yaml.Model{
		Name: "Strict",
		Fields: map[string]yaml.ModelField{
			"ID":    {Type: yaml.ModelFieldTypeAutoIncrement},
			"Name":  {Type: yaml.ModelFieldTypeString},
			"Email": {Type: yaml.ModelFieldTypeString},
		},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.ModelRelation{},
	}

	r := registry.NewRegistry()

	schemas, err := compile.MorpheModelToZodSchemas(model, r, cfg.CasingCamel)

	suite.NoError(err)
	suite.Len(schemas[0].Fields, 3)

	// All fields without "optional" attribute should be required
	for _, field := range schemas[0].Fields {
		suite.False(field.Optional, "Field %s should be required by default", field.Name)
	}
}

func (suite *CompileModelsTestSuite) TestMorpheModelToZodSchemas_PolymorphicForOnePoly_MixedPKTypes() {
	// Person uses AutoIncrement (z.number()), Organization uses UUID (z.string()).
	// The poly FK should always be z.string() to match psql (TEXT), go-struct (string),
	// and ts-types (string). Inferring from forModels[0] is order-dependent and wrong.
	commentModel := yaml.Model{
		Name: "Comment",
		Fields: map[string]yaml.ModelField{
			"ID": {Type: yaml.ModelFieldTypeUUID},
		},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.ModelRelation{
			"Commentable": {
				Type: "ForOnePoly",
				For:  []string{"Person", "Organization"},
			},
		},
	}

	r := registry.NewRegistry()
	r.SetModel("Person", yaml.Model{
		Name:   "Person",
		Fields: map[string]yaml.ModelField{"ID": {Type: yaml.ModelFieldTypeAutoIncrement}},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
	})
	r.SetModel("Organization", yaml.Model{
		Name:   "Organization",
		Fields: map[string]yaml.ModelField{"ID": {Type: yaml.ModelFieldTypeUUID}},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
	})

	schemas, err := compile.MorpheModelToZodSchemas(commentModel, r, cfg.CasingCamel)
	suite.NoError(err)

	// Find the commentableID field
	var idField zoddef.SchemaField
	for _, f := range schemas[0].Fields {
		if f.Name == "commentableID" {
			idField = f
			break
		}
	}

	// Poly FK must always be z.string() regardless of target PK types,
	// matching the convention in plugin-morphe-psql-types (TEXT),
	// plugin-morphe-go-struct (GoTypeString), and plugin-morphe-ts-types (TsTypeString).
	suite.Equal(zoddef.ZodTypeString, idField.ZodType,
		"polymorphic FK should always be z.string(), not inferred from forModels[0]")
}
