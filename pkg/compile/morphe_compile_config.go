package compile

import (
	"path"

	rcfg "github.com/kalo-build/morphe-go/pkg/registry/cfg"
	"github.com/kalo-build/plugin-morphe-zod-types/pkg/compile/cfg"
)

// MorpheCompileConfig contains all configuration for compiling Morphe to Zod schemas
type MorpheCompileConfig struct {
	// Registry loading configuration
	rcfg.MorpheLoadRegistryConfig

	// Output path for generated files
	OutputPath string

	// Format-specific configuration
	FormatConfig ZodConfig
}

// ZodConfig contains Zod-specific configuration options
type ZodConfig struct {
	// IndentSize specifies the number of spaces for indentation (default: 2)
	IndentSize int

	// UseStrictMode enables strict mode for Zod schemas (default: false)
	UseStrictMode bool

	// FieldCasing specifies the casing for field names in generated schemas.
	// Valid values: "camel" (default), "snake", "pascal"
	FieldCasing cfg.Casing
}

// DefaultMorpheCompileConfig creates a default configuration
func DefaultMorpheCompileConfig(
	yamlRegistryPath string,
	baseOutputDirPath string,
) MorpheCompileConfig {
	return MorpheCompileConfig{
		MorpheLoadRegistryConfig: rcfg.MorpheLoadRegistryConfig{
			RegistryEnumsDirPath:      path.Join(yamlRegistryPath, "enums"),
			RegistryModelsDirPath:     path.Join(yamlRegistryPath, "models"),
			RegistryStructuresDirPath: path.Join(yamlRegistryPath, "structures"),
			RegistryEntitiesDirPath:   path.Join(yamlRegistryPath, "entities"),
		},
		OutputPath: baseOutputDirPath,
		FormatConfig: ZodConfig{
			IndentSize:    2,
			UseStrictMode: false,
		},
	}
}

// Validate checks if the configuration is valid
func (config MorpheCompileConfig) Validate() error {
	// Validate registry paths
	if err := config.MorpheLoadRegistryConfig.Validate(); err != nil {
		return err
	}

	// Validate Zod-specific configuration
	if config.FormatConfig.IndentSize < 0 {
		config.FormatConfig.IndentSize = 2
	}

	if !config.FormatConfig.FieldCasing.IsValid() {
		return cfg.ErrInvalidFieldCasing(config.FormatConfig.FieldCasing)
	}

	return nil
}
