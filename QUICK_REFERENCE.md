# Quick Reference - Morphe Plugin Development

## 🚀 30-Second Start

1. **Clone & Rename**
   ```bash
   cp -r plugin-morphe-types-template plugin-morphe-python-types
   ```

2. **Global Replace** (in order)
   - `_FORMAT_` → `Python`
   - `_FORMAT_EXTENSION_` → `.py`
   - `MorpheTo_FORMAT_` → `MorpheToPython`

3. **Update Types** (`pkg/formatdef/types.go`)
   ```go
   TypeString  = BasicType{Name: "str"}
   TypeInteger = BasicType{Name: "int"}
   ```

4. **Build & Test**
   ```bash
   go build ./cmd/plugin
   ./plugin '{"inputPath":"./test","outputPath":"./out"}'
   ```

## 📁 Key Files to Modify

| File | Purpose | Priority |
|------|---------|----------|
| `pkg/formatdef/types.go` | Define your type system | 🔴 High |
| `pkg/typemap/morphe_fields.go` | Map Morphe → Your types | 🔴 High |
| `pkg/compile/compile_*.go` | Implement generators | 🟡 Medium |
| `pkg/compile/morphe_writer.go` | File naming & organization | 🟢 Low |

## 🎯 Working Examples Already Included

The template includes working generators that produce output. You can:
1. Run it immediately to see the default output
2. Modify the output format incrementally
3. Test with real Morphe schemas

## 💡 Common Patterns

### Python
```go
// types.go
TypeString  = BasicType{Name: "str"}
TypeInteger = BasicType{Name: "int"}
TypeBoolean = BasicType{Name: "bool"}
TypeFloat   = BasicType{Name: "float"}
TypeDate    = BasicType{Name: "datetime"}

// writer.go
FileExtension: ".py"
toFileName: PascalCase → snake_case
```

### TypeScript
```go
// types.go  
TypeString  = BasicType{Name: "string"}
TypeInteger = BasicType{Name: "number"}
TypeBoolean = BasicType{Name: "boolean"}
TypeDate    = BasicType{Name: "Date"}

// writer.go
FileExtension: ".ts"
toFileName: PascalCase → kebab-case
```

### Java
```go
// types.go
TypeString  = BasicType{Name: "String"}
TypeInteger = BasicType{Name: "Integer"}
TypeBoolean = BasicType{Name: "Boolean"}
TypeDate    = BasicType{Name: "LocalDateTime"}

// writer.go
FileExtension: ".java"
toFileName: Keep PascalCase
```

## 🛠️ Helper Functions Available

```go
// Content generation
cb := formatdef.NewContentBuilder("  ")
cb.Line("class %s:", name)
cb.Indent()
cb.Line("def __init__(self):")

// Name conversions
ToPascalCase("user_name")  // UserName
ToCamelCase("user_name")   // userName  
ToSnakeCase("UserName")    // user_name

// Formatting
FormatList(items, ", ")
QuoteString("value")
```

## ✅ Checklist

- [ ] Replace all `_FORMAT_` placeholders
- [ ] Update type mappings in `types.go`
- [ ] Set correct file extension
- [ ] Run and verify output compiles in target language
- [ ] Test with complex Morphe schemas
- [ ] Update README with format-specific notes

## 🎉 You're Done When...

1. `go build ./cmd/plugin` succeeds
2. Plugin generates files with correct extension
3. Generated code is valid in your target language
4. All Morphe types are properly mapped

## 💬 Tips

- Start with enums (simplest)
- Use SAMPLE_OUTPUT.md for examples
- The default output already works - just needs syntax adjustments
- Test frequently with small schemas first
