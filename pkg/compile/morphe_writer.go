package compile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kalo-build/go-util/strcase"
)

// MorpheWriter handles writing compiled Zod schemas to files
type MorpheWriter struct {
	OutputPath string
}

// NewMorpheWriter creates a new MorpheWriter instance
func NewMorpheWriter(outputPath string) *MorpheWriter {
	return &MorpheWriter{
		OutputPath: outputPath,
	}
}

// ensureDir creates a directory if it doesn't exist
func (w *MorpheWriter) ensureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// writeFile writes content to a file
func (w *MorpheWriter) writeFile(path string, content []byte) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := w.ensureDir(dir); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	return os.WriteFile(path, content, 0644)
}

// WriteEnum writes a single enum definition to a file
func (w *MorpheWriter) WriteEnum(enumName string, content []byte) error {
	fileName := toFileName(enumName) + ".ts"
	filePath := filepath.Join(w.OutputPath, "enums", fileName)
	return w.writeFile(filePath, content)
}

// WriteModel writes a single model definition to a file
func (w *MorpheWriter) WriteModel(modelName string, content []byte) error {
	fileName := toFileName(modelName) + ".ts"
	filePath := filepath.Join(w.OutputPath, "models", fileName)
	return w.writeFile(filePath, content)
}

// WriteStructure writes a single structure definition to a file
func (w *MorpheWriter) WriteStructure(structureName string, content []byte) error {
	fileName := toFileName(structureName) + ".ts"
	filePath := filepath.Join(w.OutputPath, "structures", fileName)
	return w.writeFile(filePath, content)
}

// WriteEntity writes a single entity definition to a file
func (w *MorpheWriter) WriteEntity(entityName string, content []byte) error {
	fileName := toFileName(entityName) + ".ts"
	filePath := filepath.Join(w.OutputPath, "entities", fileName)
	return w.writeFile(filePath, content)
}

// WriteAllEnums writes multiple enum definitions
func (w *MorpheWriter) WriteAllEnums(enumContents map[string][]byte) error {
	for enumName, content := range enumContents {
		if err := w.WriteEnum(enumName, content); err != nil {
			return err
		}
	}
	return nil
}

// WriteAllModels writes multiple model definitions
func (w *MorpheWriter) WriteAllModels(modelContents map[string][]byte) error {
	for modelName, content := range modelContents {
		if err := w.WriteModel(modelName, content); err != nil {
			return err
		}
	}
	return nil
}

// WriteAllStructures writes multiple structure definitions
func (w *MorpheWriter) WriteAllStructures(structureContents map[string][]byte) error {
	for structureName, content := range structureContents {
		if err := w.WriteStructure(structureName, content); err != nil {
			return err
		}
	}
	return nil
}

// WriteAllEntities writes multiple entity definitions
func (w *MorpheWriter) WriteAllEntities(entityContents map[string][]byte) error {
	for entityName, content := range entityContents {
		if err := w.WriteEntity(entityName, content); err != nil {
			return err
		}
	}
	return nil
}

// Helper function to convert type names to file names (kebab-case for TypeScript/Zod)
func toFileName(typeName string) string {
	// ToKebabCase converts PascalCase to kebab-case, then lowercase
	// (e.g., "ContactInfo" -> "contact-info", "Person" -> "person")
	return strings.ToLower(strcase.ToKebabCase(typeName))
}
