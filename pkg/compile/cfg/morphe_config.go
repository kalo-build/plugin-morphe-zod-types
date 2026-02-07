package cfg

import "fmt"

// MorpheConfig contains configuration for all Morphe type categories
// This is a simplified version without language-specific details
type MorpheConfig struct {
	// Configuration for different type categories
	Enums      EnumConfig
	Models     ModelConfig
	Structures StructureConfig
	Entities   EntityConfig

	// FieldCasing specifies the casing for field names in generated schemas.
	// Valid values: "camel" (default), "snake", "pascal"
	FieldCasing Casing
}

// EnumConfig contains configuration specific to enum generation
type EnumConfig struct {
	// TODO: Add format-specific enum configuration
	// Examples:
	// - EnumStyle string (e.g., "constant", "class", "string-literal")
	// - GenerateHelpers bool (toString, fromString methods)
	// - Prefix string (e.g., "E_" for all enum names)
}

// ModelConfig contains configuration specific to model generation
type ModelConfig struct {
	// TODO: Add format-specific model configuration
	// Examples:
	// - GenerateGettersSetters bool
	// - GenerateConstructor bool
	// - ImplementInterfaces []string
}

// StructureConfig contains configuration specific to structure generation
type StructureConfig struct {
	// TODO: Add format-specific structure configuration
	// Examples:
	// - TreatAsValueType bool
	// - GenerateEquality bool
}

// EntityConfig contains configuration specific to entity generation
type EntityConfig struct {
	// TODO: Add format-specific entity configuration
	// Examples:
	// - GenerateRepository bool
	// - IncludeRelationships bool
	// - GenerateValidation bool
}

// Validate checks if the configuration is valid
func (config MorpheConfig) Validate() error {
	if !config.FieldCasing.IsValid() {
		return fmt.Errorf("invalid fieldCasing value %q, must be one of: camel, snake, pascal, or empty", config.FieldCasing)
	}
	return nil
}
