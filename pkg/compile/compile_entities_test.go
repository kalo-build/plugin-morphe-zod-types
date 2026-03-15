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

type CompileEntitiesTestSuite struct {
	suite.Suite
}

func TestCompileEntitiesTestSuite(t *testing.T) {
	suite.Run(t, new(CompileEntitiesTestSuite))
}

func (suite *CompileEntitiesTestSuite) TestMorpheEntityToZodSchemas_BasicFields() {
	r := registry.NewRegistry()

	// Create underlying model
	r.SetModel("User", yaml.Model{
		Name: "User",
		Fields: map[string]yaml.ModelField{
			"ID":   {Type: yaml.ModelFieldTypeAutoIncrement},
			"Name": {Type: yaml.ModelFieldTypeString},
		},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.ModelRelation{},
	})

	entity := yaml.Entity{
		Name: "UserProfile",
		Fields: map[string]yaml.EntityField{
			"ID": {
				Type: "User.ID",
			},
			"Name": {
				Type:       "User.Name",
				Attributes: []string{"optional"},
			},
		},
		Identifiers: map[string]yaml.EntityIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.EntityRelation{},
	}

	schemas, err := compile.MorpheEntityToZodSchemas(entity, r, cfg.CasingCamel)

	suite.NoError(err)
	suite.Len(schemas, 2) // Main + primary identifier

	// Check main schema
	mainSchema := schemas[0]
	suite.Equal("UserProfile", mainSchema.Name)
	suite.Len(mainSchema.Fields, 2)

	// ID should be required (no optional attribute)
	idField := mainSchema.Fields[0]
	suite.Equal("id", idField.Name)
	suite.False(idField.Optional)

	// Name with "optional" attribute should be optional
	nameField := mainSchema.Fields[1]
	suite.Equal("name", nameField.Name)
	suite.True(nameField.Optional)
}

func (suite *CompileEntitiesTestSuite) TestMorpheEntityToZodSchemas_IndirectedThroughRelationship() {
	r := registry.NewRegistry()

	// Create Person model with ContactInfo relationship
	r.SetModel("Person", yaml.Model{
		Name: "Person",
		Fields: map[string]yaml.ModelField{
			"ID": {Type: yaml.ModelFieldTypeAutoIncrement},
		},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.ModelRelation{
			"ContactInfo": {Type: "HasOne"},
		},
	})

	// Create ContactInfo model
	r.SetModel("ContactInfo", yaml.Model{
		Name: "ContactInfo",
		Fields: map[string]yaml.ModelField{
			"ID":    {Type: yaml.ModelFieldTypeAutoIncrement},
			"Email": {Type: yaml.ModelFieldTypeString},
		},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.ModelRelation{},
	})

	entity := yaml.Entity{
		Name: "PersonView",
		Fields: map[string]yaml.EntityField{
			"ID": {
				Type: "Person.ID",
			},
			"Email": {
				Type: "Person.ContactInfo.Email", // Indirected through relationship
			},
		},
		Identifiers: map[string]yaml.EntityIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.EntityRelation{},
	}

	schemas, err := compile.MorpheEntityToZodSchemas(entity, r, cfg.CasingCamel)

	suite.NoError(err)
	suite.Len(schemas, 2)

	// Check that Email field is resolved
	emailField := schemas[0].Fields[0]
	suite.Equal("email", emailField.Name)
	suite.Equal(zoddef.ZodTypeString, emailField.ZodType)
}

func (suite *CompileEntitiesTestSuite) TestMorpheEntityToZodSchemas_EnumField() {
	r := registry.NewRegistry()

	r.SetEnum("Nationality", yaml.Enum{
		Name: "Nationality",
		Type: yaml.EnumTypeString,
		Entries: map[string]any{
			"US": "American",
		},
	})

	r.SetModel("Person", yaml.Model{
		Name: "Person",
		Fields: map[string]yaml.ModelField{
			"ID":          {Type: yaml.ModelFieldTypeAutoIncrement},
			"Nationality": {Type: "Nationality"},
		},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.ModelRelation{},
	})

	entity := yaml.Entity{
		Name: "PersonView",
		Fields: map[string]yaml.EntityField{
			"ID": {
				Type: "Person.ID",
			},
			"Nationality": {
				Type: "Person.Nationality",
			},
		},
		Identifiers: map[string]yaml.EntityIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.EntityRelation{},
	}

	schemas, err := compile.MorpheEntityToZodSchemas(entity, r, cfg.CasingCamel)

	suite.NoError(err)

	// Check enum field
	nationalityField := schemas[0].Fields[1]
	suite.Equal("nationality", nationalityField.Name)

	enumRef, ok := nationalityField.ZodType.(zoddef.ZodEnumRefType)
	suite.True(ok)
	suite.Equal("Nationality", enumRef.EnumName)
}

func (suite *CompileEntitiesTestSuite) TestMorpheEntityToZodSchemas_InvalidRootModel() {
	r := registry.NewRegistry()

	entity := yaml.Entity{
		Name: "Invalid",
		Fields: map[string]yaml.EntityField{
			"ID": {
				Type: "NonExistent.ID",
			},
		},
		Identifiers: map[string]yaml.EntityIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.EntityRelation{},
	}

	schemas, err := compile.MorpheEntityToZodSchemas(entity, r, cfg.CasingCamel)

	suite.Error(err)
	suite.Nil(schemas)
	suite.Contains(err.Error(), "model not found: NonExistent")
}

func (suite *CompileEntitiesTestSuite) TestMorpheEntityToZodSchemas_AliasedRelationship() {
	r := registry.NewRegistry()

	r.SetModel("Contact", yaml.Model{
		Name: "Contact",
		Fields: map[string]yaml.ModelField{
			"ID": {Type: yaml.ModelFieldTypeAutoIncrement},
		},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.ModelRelation{},
	})

	r.SetEntity("Contact", yaml.Entity{
		Name: "Contact",
		Fields: map[string]yaml.EntityField{
			"ID": {Type: "Contact.ID"},
		},
		Identifiers: map[string]yaml.EntityIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.EntityRelation{},
	})

	r.SetModel("Person", yaml.Model{
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
		},
	})

	entity := yaml.Entity{
		Name: "PersonView",
		Fields: map[string]yaml.EntityField{
			"ID": {
				Type: "Person.ID",
			},
		},
		Identifiers: map[string]yaml.EntityIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.EntityRelation{
			"WorkContact": {
				Type:    "ForOne",
				Aliased: "Contact",
			},
		},
	}

	schemas, err := compile.MorpheEntityToZodSchemas(entity, r, cfg.CasingCamel)

	suite.NoError(err)
	suite.Len(schemas[0].Fields, 3) // ID + WorkContactID + WorkContact

	// Check that WorkContact references Contact
	workContactField := schemas[0].Fields[2]
	suite.Equal("workContact", workContactField.Name)

	lazyType, ok := workContactField.ZodType.(zoddef.ZodLazyType)
	suite.True(ok)
	suite.Equal("Contact", lazyType.TypeName)
}

func (suite *CompileEntitiesTestSuite) TestMorpheEntityToZodSchemas_RelationshipHasMany() {
	r := registry.NewRegistry()

	r.SetModel("Post", yaml.Model{
		Name: "Post",
		Fields: map[string]yaml.ModelField{
			"ID": {Type: yaml.ModelFieldTypeAutoIncrement},
		},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.ModelRelation{},
	})

	r.SetEntity("Post", yaml.Entity{
		Name: "Post",
		Fields: map[string]yaml.EntityField{
			"ID": {Type: "Post.ID"},
		},
		Identifiers: map[string]yaml.EntityIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.EntityRelation{},
	})

	entity := yaml.Entity{
		Name: "Blog",
		Fields: map[string]yaml.EntityField{
			"ID": {
				Type: "Blog.ID",
			},
		},
		Identifiers: map[string]yaml.EntityIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.EntityRelation{
			"Post": {Type: "HasMany"},
		},
	}

	r.SetModel("Blog", yaml.Model{
		Name: "Blog",
		Fields: map[string]yaml.ModelField{
			"ID": {Type: yaml.ModelFieldTypeAutoIncrement},
		},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.ModelRelation{},
	})

	schemas, err := compile.MorpheEntityToZodSchemas(entity, r, cfg.CasingCamel)

	suite.NoError(err)
	suite.Len(schemas[0].Fields, 3) // ID + PostIDs + Posts

	// Check PostIDs array
	postIDsField := schemas[0].Fields[1]
	suite.Equal("postIDs", postIDsField.Name)

	arrayType, ok := postIDsField.ZodType.(zoddef.ZodArrayType)
	suite.True(ok)
	suite.Equal(zoddef.ZodTypeNumber, arrayType.ElementType)

	// Check Posts array (pluralized)
	postsField := schemas[0].Fields[2]
	suite.Equal("posts", postsField.Name)
}

func (suite *CompileEntitiesTestSuite) TestMorpheEntityToZodSchemas_RelationToUUIDEntity() {
	r := registry.NewRegistry()

	r.SetModel("Organization", yaml.Model{
		Name: "Organization",
		Fields: map[string]yaml.ModelField{
			"ID":   {Type: yaml.ModelFieldTypeUUID},
			"Name": {Type: yaml.ModelFieldTypeString},
		},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.ModelRelation{
			"Invoice": {Type: "HasMany"},
		},
	})

	r.SetEntity("Organization", yaml.Entity{
		Name: "Organization",
		Fields: map[string]yaml.EntityField{
			"ID":   {Type: "Organization.ID"},
			"Name": {Type: "Organization.Name"},
		},
		Identifiers: map[string]yaml.EntityIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.EntityRelation{},
	})

	r.SetModel("Invoice", yaml.Model{
		Name: "Invoice",
		Fields: map[string]yaml.ModelField{
			"ID":            {Type: yaml.ModelFieldTypeUUID},
			"InvoiceNumber": {Type: yaml.ModelFieldTypeString},
		},
		Identifiers: map[string]yaml.ModelIdentifier{
			"primary": {Fields: []string{"ID"}},
		},
		Related: map[string]yaml.ModelRelation{
			"Organization": {Type: "ForOne"},
		},
	})

	entity := yaml.Entity{
		Name: "Invoice",
		Fields: map[string]yaml.EntityField{
			"ID":            {Type: "Invoice.ID"},
			"InvoiceNumber": {Type: "Invoice.InvoiceNumber"},
		},
		Identifiers: map[string]yaml.EntityIdentifier{
			"primary":       {Fields: []string{"ID"}},
			"invoiceNumber": {Fields: []string{"InvoiceNumber"}},
		},
		Related: map[string]yaml.EntityRelation{
			"Organization": {Type: "ForOne"},
		},
	}

	schemas, err := compile.MorpheEntityToZodSchemas(entity, r, cfg.CasingCamel)

	suite.NoError(err)
	suite.Len(schemas[0].Fields, 4) // ID + InvoiceNumber + OrganizationID + Organization

	// OrganizationID FK should be z.string() since Organization uses UUID
	orgIDField := schemas[0].Fields[2]
	suite.Equal("organizationID", orgIDField.Name)
	suite.Equal(zoddef.ZodTypeString, orgIDField.ZodType,
		"FK to UUID entity should produce z.string(), not z.number()")
}
