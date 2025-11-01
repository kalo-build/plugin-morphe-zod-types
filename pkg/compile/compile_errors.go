package compile

import "fmt"

// Common compilation errors
// TODO: Add format-specific error types as needed

// ErrNoRegistry is returned when registry is nil
var ErrNoRegistry = fmt.Errorf("registry is nil")

// ErrInvalidFieldType is returned when a field type cannot be mapped
func ErrInvalidFieldType(fieldType string) error {
	return fmt.Errorf("invalid or unsupported field type: %s", fieldType)
}

// ErrModelNotFound is returned when a referenced model doesn't exist
func ErrModelNotFound(modelName string) error {
	return fmt.Errorf("model not found: %s", modelName)
}

// ErrEnumNotFound is returned when a referenced enum doesn't exist
func ErrEnumNotFound(enumName string) error {
	return fmt.Errorf("enum not found: %s", enumName)
}

// TODO: Add more error types specific to your format's compilation needs
// Examples:
// - ErrInvalidPackageName(name string) - for package/module validation
// - ErrReservedKeyword(word string) - when names conflict with language keywords
// - ErrCircularDependency(type1, type2 string) - for dependency issues
