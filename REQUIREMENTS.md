# Morphe Plugin Development Requirements

This document outlines the requirements and instructions for creating a new Morphe compile plugin using this template.

## Overview

The `plugin-morphe-types-template` provides a simplified, working starting point for creating new Morphe compilation plugins. It's designed as a minimal, clean boilerplate that compiles to WASM and can be customized for any target language or format.

## Key Features

1. **No Hooks System**: Simplified architecture without complex lifecycle hooks
2. **Generic Type System**: Format-agnostic type definitions ready for customization
3. **Working WASM Compilation**: Compiles out-of-the-box to WASM
4. **Clear TODOs**: Implementation guidance throughout the codebase
5. **Minimal Configuration**: Simple, extensible configuration structure

## Quick Start

### 1. Clone and Setup

```bash
# Clone the template
cp -r plugin-morphe-types-template plugin-morphe-<yourformat>-types

# Update go.mod
cd plugin-morphe-<yourformat>-types
# Edit go.mod to update module name
```

### 2. Global Search and Replace

Replace these placeholders throughout the codebase:

| Placeholder | Replace With | Example |
|------------|--------------|---------|
| `_FORMAT_` | Your format name | `Python`, `Java`, `Ruby` |
| `_FORMAT_EXTENSION_` | File extension | `.py`, `.java`, `.rb` |
| `MorpheTo_FORMAT_` | Main function name | `MorpheToPython` |

### 3. Essential Implementation Steps

#### A. Define Your Type System (`pkg/formatdef/`)

1. Update `types.go` with your format's type system:
   ```go
   var (
       TypeString  = BasicType{Name: "str"}      // Python
       TypeInteger = BasicType{Name: "int"}      // Python
       TypeFloat   = BasicType{Name: "float"}    // Python
       // etc...
   )
   ```

2. Adapt `enum.go` for your format's enum representation
3. Adapt `struct.go` for your format's class/interface/type representation

#### B. Update Type Mappings (`pkg/typemap/`)

Update `morphe_fields.go` to map Morphe types to your format:
```go
var MorpheModelFieldToPythonType = map[yaml.ModelFieldType]formatdef.Type{
    yaml.ModelFieldTypeString: PythonTypeStr,
    yaml.ModelFieldTypeInteger: PythonTypeInt,
    // etc...
}
```

#### C. Implement Content Generation

In each compile file (`compile_*.go`), implement the `generate*Content` functions:

```go
func generateEnumContent(enum *formatdef.Enum, config PythonConfig) []byte {
    // Generate actual Python enum syntax
    var content strings.Builder
    content.WriteString(fmt.Sprintf("class %s(Enum):\n", enum.Name))
    for _, entry := range enum.Entries {
        content.WriteString(fmt.Sprintf("    %s = %v\n", entry.Name, entry.Value))
    }
    return []byte(content.String())
}
```

#### D. Configure Output Options

Update `_FORMAT_Config` in `morphe_compile_config.go`:
```go
type PythonConfig struct {
    UseDataclasses bool
    GenerateTypeHints bool
    PythonVersion string
    // etc...
}
```

### 4. Build and Test

```bash
# Regular build
go build ./cmd/plugin

# WASM build
GOOS=wasip1 GOARCH=wasm go build -o dist/plugin.wasm cmd/plugin/main.go

# Test with sample data
./plugin '{"inputPath":"./test/morphe","outputPath":"./output","verbose":true}'
```

## Implementation Guidelines

### DO:
- ✅ Keep initial implementation simple
- ✅ Follow your target language's conventions
- ✅ Add clear error messages
- ✅ Test with various Morphe schemas
- ✅ Document format-specific behaviors

### DON'T:
- ❌ Add hooks unless absolutely necessary
- ❌ Over-engineer the initial version
- ❌ Ignore the existing TODOs
- ❌ Break WASM compatibility

## Architecture Overview

```
plugin-morphe-<format>-types/
├── cmd/plugin/          # Entry point
├── pkg/
│   ├── compile/         # Main compilation logic
│   ├── formatdef/       # Your format's type definitions
│   └── typemap/         # Morphe → Your format type mappings
└── dist/               # WASM output
```

## Common Patterns

### Single vs Multi-File Output

The template supports both patterns:

**Multi-File** (default):
```
output/
├── enums/
│   └── Status.py
├── models/
│   └── User.py
└── entities/
    └── UserEntity.py
```

**Single-File**:
```
output/
└── types.py  # All definitions in one file
```

### Handling Relationships

For entity relationships, consider your format's patterns:
- Foreign key fields
- Navigation properties
- Lazy loading support
- Circular dependency handling

### Format-Specific Features

Add configuration for format-specific features:
- Decorators/Annotations
- Access modifiers
- Package/Module structure
- Import management
- Serialization support

## Testing Your Plugin

1. Create test Morphe schemas in `testdata/`
2. Run compilation tests
3. Verify output compiles in your target language
4. Test with the Kalo CLI

## Getting Help

- Review existing plugins for patterns (but don't copy complexity)
- Check Morphe spec documentation
- Test with simple schemas first
- Add features incrementally

## Future Enhancements

Once basic functionality works, consider:
- Advanced type mappings
- Custom validation
- Code generation optimizations
- IDE integration files
- Documentation generation
