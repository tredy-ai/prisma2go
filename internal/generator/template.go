package generator

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/tredy-ai/prisma2go/internal/config"
	"github.com/tredy-ai/prisma2go/internal/dmmf"
)

// TemplateData holds all data needed to render the template.
type TemplateData struct {
	Package string
	Imports []Import
	Enums   []EnumData
	Models  []ModelData
}

// EnumData holds processed enum data for templating.
type EnumData struct {
	Name     string
	Values   []EnumValueData
	Trimmed  string // Name with common prefix removed for enumer
}

// EnumValueData holds a single enum value for templating.
type EnumValueData struct {
	ConstName string // e.g., RoleAdmin
	RawName   string // e.g., ADMIN (original Prisma name)
}

// ModelData holds processed model data for templating.
type ModelData struct {
	Name          string
	TableName     string // Database table name
	Documentation string
	Fields        []FieldData
	Indexes       []string // Index descriptions for comment
	HasRelations  bool
}

// FieldData holds processed field data for templating.
type FieldData struct {
	Name          string
	GoType        string
	JSONTag       string
	DBTag         string
	Comment       string // inline comment (e.g., "primary key")
	IsRelation    bool
	Documentation string
}

// Render generates the Go source code from DMMF data.
func Render(datamodel *dmmf.Datamodel, cfg *config.Config) ([]byte, error) {
	data := prepareTemplateData(datamodel, cfg)

	tmpl, err := template.New("models").Parse(modelTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// prepareTemplateData converts DMMF data into template-ready data.
func prepareTemplateData(datamodel *dmmf.Datamodel, cfg *config.Config) *TemplateData {
	mapper := NewTypeMapper(cfg, datamodel.Enums)
	importSet := make(map[string]Import)

	// Process enums
	enums := make([]EnumData, 0, len(datamodel.Enums))
	for _, e := range datamodel.Enums {
		enumName := ToPascalCase(e.Name)
		values := make([]EnumValueData, 0, len(e.Values))
		for _, v := range e.Values {
			values = append(values, EnumValueData{
				ConstName: ToEnumConstName(enumName, v.Name),
				RawName:   v.Name,
			})
		}
		enums = append(enums, EnumData{
			Name:    enumName,
			Values:  values,
			Trimmed: enumName,
		})
	}

	// Process models
	models := make([]ModelData, 0, len(datamodel.Models))
	for _, m := range datamodel.Models {
		fields := make([]FieldData, 0, len(m.Fields))
		hasRelations := false

		for _, f := range m.Fields {
			goType, imports := mapper.MapType(&f)
			for _, imp := range imports {
				importSet[imp.Path] = imp
			}

			// Build tags
			var jsonTag, dbTag string
			if cfg.JSONTags {
				jsonTag = buildJSONTag(&f, cfg)
			}
			if cfg.DBTags {
				dbTag = buildDBTag(&f)
			}

			// Build comment
			var comment string
			if f.IsID {
				comment = "primary key"
			}

			if f.IsRelation() {
				hasRelations = true
			}

			var docComment string
			if f.Documentation != nil {
				docComment = *f.Documentation
			}

			fields = append(fields, FieldData{
				Name:          ToPascalCase(f.Name),
				GoType:        goType,
				JSONTag:       jsonTag,
				DBTag:         dbTag,
				Comment:       comment,
				IsRelation:    f.IsRelation(),
				Documentation: docComment,
			})
		}

		// Build index descriptions
		indexes := buildIndexDescriptions(&m)

		var docComment string
		if m.Documentation != nil {
			docComment = *m.Documentation
		}

		models = append(models, ModelData{
			Name:          ToPascalCase(m.Name),
			TableName:     m.GetDBName(),
			Documentation: docComment,
			Fields:        fields,
			Indexes:       indexes,
			HasRelations:  hasRelations,
		})
	}

	// Sort imports
	imports := make([]Import, 0, len(importSet))
	for _, imp := range importSet {
		imports = append(imports, imp)
	}
	sort.Slice(imports, func(i, j int) bool {
		return imports[i].Path < imports[j].Path
	})

	return &TemplateData{
		Package: cfg.Package,
		Imports: imports,
		Enums:   enums,
		Models:  models,
	}
}

// buildJSONTag creates the json struct tag for a field.
func buildJSONTag(f *dmmf.Field, cfg *config.Config) string {
	tag := f.Name

	// Relations always get omitempty since they're manually populated
	if f.IsRelation() {
		return fmt.Sprintf(`json:"%s,omitempty"`, tag)
	}

	if !f.IsRequired && cfg.JSONOmitempty {
		tag += ",omitempty"
	}
	if f.HasDefaultValue && cfg.JSONOmitempty {
		if !strings.Contains(tag, "omitempty") {
			tag += ",omitempty"
		}
	}
	return fmt.Sprintf(`json:"%s"`, tag)
}

// buildDBTag creates the db struct tag for a field.
func buildDBTag(f *dmmf.Field) string {
	if f.IsRelation() {
		return `db:"-"`
	}
	return fmt.Sprintf(`db:"%s"`, f.GetDBName())
}

// buildIndexDescriptions creates human-readable index descriptions.
func buildIndexDescriptions(m *dmmf.Model) []string {
	var indexes []string

	// Unique indexes
	for _, idx := range m.UniqueIndexes {
		name := "unique"
		if idx.Name != nil {
			name = *idx.Name
		}
		indexes = append(indexes, fmt.Sprintf("%s (%s)", name, strings.Join(idx.Fields, ", ")))
	}

	// Composite unique fields (@@unique without explicit name)
	for _, fields := range m.UniqueFields {
		if len(fields) > 0 {
			indexes = append(indexes, fmt.Sprintf("unique (%s)", strings.Join(fields, ", ")))
		}
	}

	return indexes
}

const modelTemplate = `// Code generated by prisma2go. DO NOT EDIT.
package {{.Package}}
{{if .Imports}}
import (
{{- range .Imports}}
	"{{.Path}}"
{{- end}}
)
{{end}}
{{- if .Enums}}

// Enums
{{range $enum := .Enums}}
//go:generate enumer -type={{$enum.Name}} -json -sql -text -trimprefix={{$enum.Trimmed}}
type {{$enum.Name}} int

const (
{{- range $i, $v := $enum.Values}}
{{- if eq $i 0}}
	{{$v.ConstName}} {{$enum.Name}} = iota
{{- else}}
	{{$v.ConstName}}
{{- end}}
{{- end}}
)
{{end}}
{{- end}}
{{- if .Models}}

// Models
{{range .Models}}
{{- if .Documentation}}
// {{.Name}} - {{.Documentation}}
{{- else}}
// {{.Name}} represents the {{.TableName}} table.
{{- end}}
{{- range .Indexes}}
// Index: {{.}}
{{- end}}
type {{.Name}} struct {
{{- $hasNonRelation := false}}
{{- range .Fields}}
{{- if not .IsRelation}}
{{- $hasNonRelation = true}}
{{- if .Documentation}}
	// {{.Documentation}}
{{- end}}
	{{.Name}} {{.GoType}} ` + "`" + `{{.JSONTag}} {{.DBTag}}` + "`" + `{{if .Comment}} // {{.Comment}}{{end}}
{{- end}}
{{- end}}
{{- if .HasRelations}}

	// Relations (not persisted, populate manually)
{{- range .Fields}}
{{- if .IsRelation}}
	{{.Name}} {{.GoType}} ` + "`" + `{{.JSONTag}} {{.DBTag}}` + "`" + `
{{- end}}
{{- end}}
{{- end}}
}
{{end}}
{{- end}}
`
