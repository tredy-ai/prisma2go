package generator

import (
	"github.com/tredy-ai/prisma2go/internal/config"
	"github.com/tredy-ai/prisma2go/internal/dmmf"
)

// TypeMapper handles Prisma to Go type conversions.
type TypeMapper struct {
	cfg       *config.Config
	enumNames map[string]bool
}

// NewTypeMapper creates a new TypeMapper with the given config and known enum names.
func NewTypeMapper(cfg *config.Config, enums []dmmf.Enum) *TypeMapper {
	enumNames := make(map[string]bool)
	for _, e := range enums {
		enumNames[e.Name] = true
	}
	return &TypeMapper{cfg: cfg, enumNames: enumNames}
}

// Import represents a Go import.
type Import struct {
	Path  string
	Alias string // Optional alias
}

// MapType converts a Prisma field to a Go type string and returns any required imports.
func (tm *TypeMapper) MapType(field *dmmf.Field) (goType string, imports []Import) {
	baseType, baseImports := tm.mapBaseType(field)

	// Handle arrays
	if field.IsList {
		return "[]" + baseType, baseImports
	}

	// Handle nullable (pointer for scalars and enums, slice for relations is already handled)
	if !field.IsRequired && !field.IsRelation() {
		return "*" + baseType, baseImports
	}

	// Relations are always pointers for 1:1, slices for 1:N/N:M
	if field.IsRelation() && !field.IsList {
		return "*" + baseType, baseImports
	}

	return baseType, baseImports
}

// mapBaseType returns the base Go type (without pointer/slice) for a Prisma field.
func (tm *TypeMapper) mapBaseType(field *dmmf.Field) (string, []Import) {
	// Check for composite types first (these become json.RawMessage)
	if field.Kind == "object" && !tm.enumNames[field.Type] {
		// This is either a relation or a composite type
		// Relations are handled by the model name directly
		return field.Type, nil
	}

	// Enums
	if field.IsEnum() {
		return field.Type, nil
	}

	// Check for native type overrides
	if len(field.NativeType) > 0 {
		if nativeType, ok := field.NativeType[0].(string); ok {
			if goType, imports := tm.mapNativeType(nativeType, field.Type); goType != "" {
				return goType, imports
			}
		}
	}

	// Standard scalar type mapping
	return tm.mapScalarType(field.Type)
}

// mapScalarType converts a Prisma scalar type to a Go type.
func (tm *TypeMapper) mapScalarType(prismaType string) (string, []Import) {
	switch prismaType {
	case "String":
		return "string", nil
	case "Int":
		return "int32", nil
	case "BigInt":
		return "int64", nil
	case "Float":
		return "float64", nil
	case "Boolean":
		return "bool", nil
	case "DateTime":
		return "time.Time", []Import{{Path: "time"}}
	case "Json":
		return "json.RawMessage", []Import{{Path: "encoding/json"}}
	case "Bytes":
		return "[]byte", nil
	case "Decimal":
		if tm.cfg.DecimalPackage != "" {
			return "decimal.Decimal", []Import{{Path: tm.cfg.DecimalPackage}}
		}
		return "string", nil
	default:
		// Unknown type, use json.RawMessage as fallback
		return "json.RawMessage", []Import{{Path: "encoding/json"}}
	}
}

// mapNativeType handles @db.X native type annotations.
func (tm *TypeMapper) mapNativeType(nativeType, baseType string) (string, []Import) {
	switch nativeType {
	case "Uuid":
		if tm.cfg.UUIDPackage != "" {
			return "uuid.UUID", []Import{{Path: tm.cfg.UUIDPackage}}
		}
		return "string", nil
	case "JsonB", "Json":
		return "json.RawMessage", []Import{{Path: "encoding/json"}}
	case "Text", "VarChar", "Char", "Citext":
		return "string", nil
	case "SmallInt", "Integer", "Int2", "Int4":
		return "int32", nil
	case "BigInt", "Int8":
		return "int64", nil
	case "Real", "Float4":
		return "float32", nil
	case "DoublePrecision", "Float8":
		return "float64", nil
	case "Decimal", "Numeric", "Money":
		if tm.cfg.DecimalPackage != "" {
			return "decimal.Decimal", []Import{{Path: tm.cfg.DecimalPackage}}
		}
		return "string", nil
	case "Timestamp", "Timestamptz", "Date", "Time", "Timetz":
		return "time.Time", []Import{{Path: "time"}}
	case "ByteA":
		return "[]byte", nil
	default:
		// Unknown native type, fall back to base type mapping
		return "", nil
	}
}
