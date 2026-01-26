package dmmf

// Datamodel represents the root datamodel section of DMMF.
type Datamodel struct {
	Enums  []Enum  `json:"enums"`
	Models []Model `json:"models"`
	Types  []Type  `json:"types"` // Composite types (we don't fully support these)
}

// Enum represents a Prisma enum.
type Enum struct {
	Name   string      `json:"name"`
	Values []EnumValue `json:"values"`
	DBName *string     `json:"dbName"`
}

// EnumValue represents a single value in a Prisma enum.
type EnumValue struct {
	Name   string  `json:"name"`
	DBName *string `json:"dbName"`
}

// Model represents a Prisma model.
type Model struct {
	Name          string        `json:"name"`
	DBName        *string       `json:"dbName"`
	Fields        []Field       `json:"fields"`
	PrimaryKey    *PrimaryKey   `json:"primaryKey"`
	UniqueFields  [][]string    `json:"uniqueFields"`
	UniqueIndexes []UniqueIndex `json:"uniqueIndexes"`
	Documentation *string       `json:"documentation"`
}

// Type represents a Prisma composite type.
type Type struct {
	Name   string  `json:"name"`
	Fields []Field `json:"fields"`
}

// Field represents a field in a Prisma model or type.
type Field struct {
	Name            string   `json:"name"`
	Kind            string   `json:"kind"` // "scalar", "enum", "object"
	Type            string   `json:"type"` // The actual type name
	IsList          bool     `json:"isList"`
	IsRequired      bool     `json:"isRequired"`
	IsUnique        bool     `json:"isUnique"`
	IsID            bool     `json:"isId"`
	IsReadOnly      bool     `json:"isReadOnly"`
	HasDefaultValue bool     `json:"hasDefaultValue"`
	Default         any      `json:"default"`
	DBName          *string  `json:"dbName"`
	RelationName    *string  `json:"relationName"`
	RelationFromFields []string `json:"relationFromFields"`
	RelationToFields   []string `json:"relationToFields"`
	Documentation   *string  `json:"documentation"`
	NativeType      []any    `json:"nativeType"` // e.g., ["Uuid"] or ["VarChar", "255"]
}

// PrimaryKey represents a composite primary key (@@id).
type PrimaryKey struct {
	Name   *string  `json:"name"`
	Fields []string `json:"fields"`
}

// UniqueIndex represents a unique constraint (@@unique).
type UniqueIndex struct {
	Name   *string  `json:"name"`
	Fields []string `json:"fields"`
}

// IsRelation returns true if this field is a relation (not a scalar/enum).
func (f *Field) IsRelation() bool {
	return f.Kind == "object"
}

// IsEnum returns true if this field is an enum type.
func (f *Field) IsEnum() bool {
	return f.Kind == "enum"
}

// IsScalar returns true if this field is a scalar type.
func (f *Field) IsScalar() bool {
	return f.Kind == "scalar"
}

// GetDBName returns the database column/table name, falling back to the Prisma name.
func (f *Field) GetDBName() string {
	if f.DBName != nil {
		return *f.DBName
	}
	return f.Name
}

// GetDBName returns the database table name, falling back to the model name.
func (m *Model) GetDBName() string {
	if m.DBName != nil {
		return *m.DBName
	}
	return m.Name
}
