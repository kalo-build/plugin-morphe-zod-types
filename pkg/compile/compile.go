package compile

import (
	"fmt"

	"github.com/kalo-build/morphe-go/pkg/registry"
)

// MorpheToZod compiles a Morphe registry to Zod schemas
func MorpheToZod(config MorpheCompileConfig) error {
	// Load the Morphe registry
	r, rErr := registry.LoadMorpheRegistry(registry.LoadMorpheRegistryHooks{}, config.MorpheLoadRegistryConfig)
	if rErr != nil {
		return fmt.Errorf("failed to load morphe registry: %w", rErr)
	}

	// Initialize the writer
	writer := NewMorpheWriter(config.OutputPath)

	// Process enums if present
	if r.HasEnums() {
		fmt.Println("Compiling enums...")
		if err := CompileAllEnums(config, r, writer); err != nil {
			return fmt.Errorf("failed to compile enums: %w", err)
		}
	}

	// Process models if present
	if r.HasModels() {
		fmt.Println("Compiling models...")
		if err := CompileAllModels(config, r, writer); err != nil {
			return fmt.Errorf("failed to compile models: %w", err)
		}
	}

	// Process structures if present
	if r.HasStructures() {
		fmt.Println("Compiling structures...")
		if err := CompileAllStructures(config, r, writer); err != nil {
			return fmt.Errorf("failed to compile structures: %w", err)
		}
	}

	// Process entities if present
	if r.HasEntities() {
		// Entities depend on models
		if !r.HasModels() {
			return fmt.Errorf("entities compilation requires models to be compiled")
		}

		fmt.Println("Compiling entities...")
		if err := CompileAllEntities(config, r, writer); err != nil {
			return fmt.Errorf("failed to compile entities: %w", err)
		}
	}

	return nil
}
