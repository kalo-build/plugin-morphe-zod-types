package compile_test

import (
	"testing"

	"github.com/kalo-build/morphe-go/pkg/yaml"
	"github.com/kalo-build/plugin-morphe-zod-types/pkg/compile"
	"github.com/kalo-build/plugin-morphe-zod-types/pkg/zoddef"
	"github.com/stretchr/testify/suite"
)

type CompileEnumsTestSuite struct {
	suite.Suite
}

func TestCompileEnumsTestSuite(t *testing.T) {
	suite.Run(t, new(CompileEnumsTestSuite))
}

func (suite *CompileEnumsTestSuite) TestMorpheEnumToZodEnum_String() {
	enum := yaml.Enum{
		Name: "Color",
		Type: yaml.EnumTypeString,
		Entries: map[string]any{
			"Red":   "rgb(255,0,0)",
			"Green": "rgb(0,255,0)",
			"Blue":  "rgb(0,0,255)",
		},
	}

	zodEnum, err := compile.MorpheEnumToZodEnum(enum)

	suite.NoError(err)
	suite.NotNil(zodEnum)
	suite.Equal("Color", zodEnum.Name)
	suite.Equal(zoddef.ZodTypeString, zodEnum.Type)
	suite.Len(zodEnum.Entries, 3)

	// Entries should be sorted
	suite.Equal("Blue", zodEnum.Entries[0].Name)
	suite.Equal("rgb(0,0,255)", zodEnum.Entries[0].Value)
	suite.Equal("Green", zodEnum.Entries[1].Name)
	suite.Equal("rgb(0,255,0)", zodEnum.Entries[1].Value)
	suite.Equal("Red", zodEnum.Entries[2].Name)
	suite.Equal("rgb(255,0,0)", zodEnum.Entries[2].Value)
}

func (suite *CompileEnumsTestSuite) TestMorpheEnumToZodEnum_Integer() {
	enum := yaml.Enum{
		Name: "StatusCode",
		Type: yaml.EnumTypeInteger,
		Entries: map[string]any{
			"OK":                  200,
			"NotFound":            404,
			"InternalServerError": 500,
		},
	}

	zodEnum, err := compile.MorpheEnumToZodEnum(enum)

	suite.NoError(err)
	suite.NotNil(zodEnum)
	suite.Equal("StatusCode", zodEnum.Name)
	suite.Equal(zoddef.ZodTypeNumber, zodEnum.Type)
	suite.Len(zodEnum.Entries, 3)

	// Check values are preserved
	suite.Equal("InternalServerError", zodEnum.Entries[0].Name)
	suite.Equal(500, zodEnum.Entries[0].Value)
}

func (suite *CompileEnumsTestSuite) TestMorpheEnumToZodEnum_Float() {
	enum := yaml.Enum{
		Name: "MathConstant",
		Type: yaml.EnumTypeFloat,
		Entries: map[string]any{
			"Pi":    3.14159,
			"Euler": 2.71828,
		},
	}

	zodEnum, err := compile.MorpheEnumToZodEnum(enum)

	suite.NoError(err)
	suite.NotNil(zodEnum)
	suite.Equal("MathConstant", zodEnum.Name)
	suite.Equal(zoddef.ZodTypeNumber, zodEnum.Type)
	suite.Len(zodEnum.Entries, 2)
}

func (suite *CompileEnumsTestSuite) TestMorpheEnumToZodEnum_EmptyEntries() {
	enum := yaml.Enum{
		Name:    "Empty",
		Type:    yaml.EnumTypeString,
		Entries: map[string]any{},
	}

	zodEnum, err := compile.MorpheEnumToZodEnum(enum)

	suite.NoError(err)
	suite.NotNil(zodEnum)
	suite.Len(zodEnum.Entries, 0)
}

func (suite *CompileEnumsTestSuite) TestMorpheEnumToZodEnum_SingleEntry() {
	enum := yaml.Enum{
		Name: "SingleValue",
		Type: yaml.EnumTypeString,
		Entries: map[string]any{
			"Only": "value",
		},
	}

	zodEnum, err := compile.MorpheEnumToZodEnum(enum)

	suite.NoError(err)
	suite.NotNil(zodEnum)
	suite.Len(zodEnum.Entries, 1)
	suite.Equal("Only", zodEnum.Entries[0].Name)
	suite.Equal("value", zodEnum.Entries[0].Value)
}
