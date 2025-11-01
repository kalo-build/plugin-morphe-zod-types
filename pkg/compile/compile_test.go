package compile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/kalo-build/go-util/assertfile"
	rcfg "github.com/kalo-build/morphe-go/pkg/registry/cfg"
	"github.com/kalo-build/plugin-morphe-zod-types/internal/testutils"
	"github.com/kalo-build/plugin-morphe-zod-types/pkg/compile"
)

type CompileTestSuite struct {
	assertfile.FileSuite

	TestDirPath            string
	TestGroundTruthDirPath string

	EnumsDirPath      string
	ModelsDirPath     string
	StructuresDirPath string
	EntitiesDirPath   string
}

func TestCompileTestSuite(t *testing.T) {
	suite.Run(t, new(CompileTestSuite))
}

func (suite *CompileTestSuite) SetupTest() {
	suite.TestDirPath = testutils.GetTestDirPath()
	suite.TestGroundTruthDirPath = filepath.Join(suite.TestDirPath, "ground-truth", "compile-minimal")

	suite.EnumsDirPath = filepath.Join(suite.TestDirPath, "registry", "minimal", "enums")
	suite.ModelsDirPath = filepath.Join(suite.TestDirPath, "registry", "minimal", "models")
	suite.StructuresDirPath = filepath.Join(suite.TestDirPath, "registry", "minimal", "structures")
	suite.EntitiesDirPath = filepath.Join(suite.TestDirPath, "registry", "minimal", "entities")
}

func (suite *CompileTestSuite) TearDownTest() {
	suite.TestDirPath = ""
}

func (suite *CompileTestSuite) TestMorpheToZod() {
	workingDirPath := suite.TestDirPath + "/working"
	os.RemoveAll(workingDirPath) // Clean up first
	suite.Nil(os.Mkdir(workingDirPath, 0755))
	defer os.RemoveAll(workingDirPath)

	config := compile.MorpheCompileConfig{
		MorpheLoadRegistryConfig: rcfg.MorpheLoadRegistryConfig{
			RegistryEnumsDirPath:      suite.EnumsDirPath,
			RegistryStructuresDirPath: suite.StructuresDirPath,
			RegistryModelsDirPath:     suite.ModelsDirPath,
			RegistryEntitiesDirPath:   suite.EntitiesDirPath,
		},
		OutputPath: workingDirPath,
		FormatConfig: compile.ZodConfig{
			IndentSize:    2,
			UseStrictMode: false,
		},
	}

	compileErr := compile.MorpheToZod(config)
	suite.NoError(compileErr)

	// Test enums
	enumsDirPath := workingDirPath + "/enums"
	gtEnumsDirPath := suite.TestGroundTruthDirPath + "/enums"
	suite.DirExists(enumsDirPath)

	enumPath0 := enumsDirPath + "/nationality.ts"
	gtEnumPath0 := gtEnumsDirPath + "/nationality.ts"
	suite.FileExists(enumPath0)
	suite.FileEquals(enumPath0, gtEnumPath0)

	enumPath1 := enumsDirPath + "/universal-number.ts"
	gtEnumPath1 := gtEnumsDirPath + "/universal-number.ts"
	suite.FileExists(enumPath1)
	suite.FileEquals(enumPath1, gtEnumPath1)

	// Test models
	modelsDirPath := workingDirPath + "/models"
	gtModelsDirPath := suite.TestGroundTruthDirPath + "/models"
	suite.DirExists(modelsDirPath)

	modelPath0 := modelsDirPath + "/company.ts"
	gtModelPath0 := gtModelsDirPath + "/company.ts"
	suite.FileExists(modelPath0)
	suite.FileEquals(modelPath0, gtModelPath0)

	modelPath1 := modelsDirPath + "/contact-info.ts"
	gtModelPath1 := gtModelsDirPath + "/contact-info.ts"
	suite.FileExists(modelPath1)
	suite.FileEquals(modelPath1, gtModelPath1)

	modelPath2 := modelsDirPath + "/person.ts"
	gtModelPath2 := gtModelsDirPath + "/person.ts"
	suite.FileExists(modelPath2)
	suite.FileEquals(modelPath2, gtModelPath2)

	// Test structures
	structuresDirPath := workingDirPath + "/structures"
	gtStructuresDirPath := suite.TestGroundTruthDirPath + "/structures"
	suite.DirExists(structuresDirPath)

	structurePath0 := structuresDirPath + "/address.ts"
	gtStructurePath0 := gtStructuresDirPath + "/address.ts"
	suite.FileExists(structurePath0)
	suite.FileEquals(structurePath0, gtStructurePath0)

	// Test entities
	entitiesDirPath := workingDirPath + "/entities"
	gtEntitiesDirPath := suite.TestGroundTruthDirPath + "/entities"
	suite.DirExists(entitiesDirPath)

	entityPath0 := entitiesDirPath + "/company.ts"
	gtEntityPath0 := gtEntitiesDirPath + "/company.ts"
	suite.FileExists(entityPath0)
	suite.FileEquals(entityPath0, gtEntityPath0)

	entityPath1 := entitiesDirPath + "/person.ts"
	gtEntityPath1 := gtEntitiesDirPath + "/person.ts"
	suite.FileExists(entityPath1)
	suite.FileEquals(entityPath1, gtEntityPath1)
}
