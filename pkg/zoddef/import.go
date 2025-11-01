package zoddef

// SchemaImport represents an import statement in a Zod schema file
type SchemaImport struct {
	Names      []string // Named imports (e.g., ["z"] from "zod")
	ImportPath string   // Import path (e.g., "zod", "../enums/nationality")
}

func (i SchemaImport) DeepClone() SchemaImport {
	names := make([]string, len(i.Names))
	copy(names, i.Names)
	return SchemaImport{
		Names:      names,
		ImportPath: i.ImportPath,
	}
}
