# Morphe Zod Types Plugin

A plugin for the Kalo CLI that converts Morphe registry definitions to Zod schemas for runtime validation and TypeScript type generation.

## Table of Contents

- [Usage](#usage)
  - [Configuration](#configuration)
  - [Output Structure](#output-structure)
- [Error Codes](#error-codes)
- [Development](#development)
  - [Building as WASM (WASI)](#building-as-wasm-wasi)
- [Overview](#overview)
- [Features](#features)
- [Example](#example)
- [License](#license)

## Usage

This plugin is intended to be called by the Kalo CLI. It converts Morphe registry YAML files to Zod validation schemas and TypeScript types.

### Configuration

The plugin accepts a JSON config string with the following parameters:

```json
{
  "inputPath": "/path/to/morphe/registry",
  "outputPath": "/path/to/output/directory",
  "verbose": true,
  "config": {
    "indentSize": 2,
    "useStrictMode": false
  }
}
```

#### Parameters

- `inputPath` (required): Path to the Morphe registry directory.
- `outputPath` (required): Path where Zod schemas will be generated.
- `verbose` (optional): Enable verbose logging for debugging. If not provided, defaults to 'false'.
- `config` (optional): Additional configuration options:
  - `indentSize`: Number of spaces for indentation (default: 2)
  - `useStrictMode`: Enable strict mode for Zod schemas (default: false)

### Output Structure

The plugin generates TypeScript files with Zod schemas in the following structure:

```
outputPath/
  ├── enums/
  │   └── [enum-files].ts
  ├── models/
  │   └── [model-files].ts
  ├── structures/
  │   └── [structure-files].ts
  └── entities/
      └── [entity-files].ts
```

## Error Codes

| Code | Description |
|------|-------------|
| 1 | Compilation failed |
| 3 | Missing config |
| 4 | Invalid config JSON |
| 12 | Input path is required |
| 13 | Output path is required |

## Development

This plugin is designed to work in WASM (WASI) format when called by the Kalo CLI.

### Building as WASM (WASI)

The utility scripts `./scripts/build.bat` and `./scripts/build.sh` can be used to generate a build under `/dist`.

To build manually, run this command from the project root:

```bash
GOOS=wasip1 GOARCH=wasm go build -o ./dist/morphe-zod-types-v1.0.0.wasm ./cmd/plugin/main.go
```

## Overview

This plugin generates Zod schemas and TypeScript types from Morphe specifications by converting:

- Models to Zod object schemas with validation
- Entities to Zod object schemas
- Enums to TypeScript enums with Zod native enum schemas
- Structures to Zod object schemas

## Features

- **Runtime Validation**: Generate Zod schemas for runtime type validation
- **Type Inference**: Automatic TypeScript type generation via `z.infer<>`
- **Comprehensive Support**: All Morphe relationship types supported:
  - HasOne/HasMany/ForOne/ForMany for model relationships
  - Polymorphic relationships (HasOnePoly, HasManyPoly, ForOnePoly, ForManyPoly)
  - Relationship aliasing
- **Field Types**: Handles all primitive types:
  - String, Number, Boolean, Date
  - UUID fields
  - Custom enum types
- **Identifier Schemas**: Generates separate schemas for model identifiers
- **Lazy Loading**: Uses `z.lazy()` for circular references
- **Optional Fields**: Properly handles mandatory vs. optional fields

## Example

Given a Morphe model:

```yaml
name: Person
fields:
  ID:
    type: AutoIncrement
    attributes:
      - mandatory
  FirstName:
    type: String
  LastName:
    type: String
  Nationality:
    type: Nationality
identifiers:
  primary: ID
  name:
    - FirstName
    - LastName
related:
  ContactInfo:
    type: HasOne
  Company:
    type: ForOne
```

The plugin generates:

```typescript
import { z } from 'zod'
import { Nationality, NationalitySchema } from '../enums/nationality'
import { CompanySchema } from './company'
import { ContactInfoSchema } from './contact-info'

export const PersonSchema = z.object({
  firstName: z.string().optional(),
  id: z.number(),
  lastName: z.string().optional(),
  nationality: NationalitySchema.optional(),
  companyID: z.number().optional(),
  company: z.lazy(() => CompanySchema).optional(),
  contactInfoID: z.number().optional(),
  contactInfo: z.lazy(() => ContactInfoSchema).optional()
})

export type Person = z.infer<typeof PersonSchema>

export const PersonIDNameSchema = z.object({
  firstName: z.string().optional(),
  lastName: z.string().optional()
})

export type PersonIDName = z.infer<typeof PersonIDNameSchema>

export const PersonIDPrimarySchema = z.object({
  id: z.number()
})

export type PersonIDPrimary = z.infer<typeof PersonIDPrimarySchema>
```

## Testing

Run the test suite:

```bash
go test ./pkg/compile -v
```

## License

MIT License
