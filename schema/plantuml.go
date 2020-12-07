package schema

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/covrom/goerd/dict"
	"github.com/pkg/errors"
)

// PlantUML struct
type PlantUML struct {
	Dict     *dict.Dict
	Distance int
}

// New return PlantUML
func NewPlantUML(d *dict.Dict, distance int) *PlantUML {
	return &PlantUML{
		Dict:     d,
		Distance: distance,
	}
}

// TODO: add constraints and triggers

func (p *PlantUML) schemaTemplate() string {
	return `@startuml
hide methods
hide stereotypes

skinparam class {
	BackgroundColor White
	BorderColor #6E6E6E
	ArrowColor #6E6E6E
}

' tables
{{- range $i, $t := .Schema.Tables }}
rectangle "{{ $t.Name }}" {
	{{- if ne $t.Type "VIEW" }}
	entity {{ $t.Name }} as "{{ $t.Name }}" << (T,#5DBCD2) >> {
	{{- else }}
	entity {{ $t.Name }} as "{{ $t.Name }}" << (V,#C6EDDB) >> {
	{{- end }}
	{{- range $ii, $c := $t.Columns }}
		{{ $c.Name | html }} <font color="#666666">[{{ $c.Type | html }}]</font>
	{{- end }}
	}
	{{- range $ii, $c := $t.Indexes }}
	entity {{ $c.Name }} as "{{ $c.Name }}" << (I,#D25D8A) >> {
		{{- range $iii, $cc := $c.Columns }}
		{{ $cc | html }}
		{{- end }}
	}
	"{{ $c.Name }}" -- "{{ $t.Name }}" : "{{if $c.IsPrimary}}PRIMARY KEY{{else}}{{if $c.IsUnique}}UNIQUE{{end}} {{$c.MethodName}}{{end}} {{if $c.IsClustered}}CLUSTERED{{end}}"
	{{- end }}
}
{{- end }}

' relations
{{- range $j, $r := .Schema.Relations }}
"{{ $r.Table.Name }}" }-- "{{ $r.ParentTable.Name }}" : "{{ $r.OnDelete | html }}"
{{- end }}

@enduml
`
}

// OutputSchema output dot format for full relation.
func (p *PlantUML) OutputSchema(wr io.Writer, s *Schema) error {
	for _, t := range s.Tables {
		err := addPrefix(t)
		if err != nil {
			return err
		}
	}

	ts := p.schemaTemplate()

	tmpl := template.Must(template.New(s.Name).Funcs(Funcs(p.Dict)).Parse(ts))
	err := tmpl.Execute(wr, map[string]interface{}{
		"Schema":      s,
		"showComment": false,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func addPrefix(t *Table) error {
	// PRIMARY KEY
	for _, i := range t.Indexes {
		if strings.Index(i.Def, "PRIMARY") < 0 {
			continue
		}
		for _, c := range i.Columns {
			column, err := t.FindColumnByName(c)
			if err != nil {
				return err
			}
			column.Name = fmt.Sprintf("+ %s", column.Name)
		}
	}
	// Foreign Key (Relations)
	for _, c := range t.Columns {
		if len(c.ParentRelations) > 0 && strings.Index(c.Name, "+") < 0 {
			c.Name = fmt.Sprintf("# %s", c.Name)
		}
	}
	return nil
}

func contains(rs []*Relation, e *Relation) bool {
	for _, r := range rs {
		if e == r {
			return true
		}
	}
	return false
}

func Funcs(d *dict.Dict) map[string]interface{} {
	return template.FuncMap{
		"nl2br": func(text string) string {
			r := strings.NewReplacer("\r\n", "<br>", "\n", "<br>", "\r", "<br>")
			return r.Replace(text)
		},
		"nl2br_slash": func(text string) string {
			r := strings.NewReplacer("\r\n", "<br />", "\n", "<br />", "\r", "<br />")
			return r.Replace(text)
		},
		"nl2mdnl": func(text string) string {
			r := strings.NewReplacer("\r\n", "  \n", "\n", "  \n", "\r", "  \n")
			return r.Replace(text)
		},
		"nl2space": func(text string) string {
			r := strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ")
			return r.Replace(text)
		},
		"escape_nl": func(text string) string {
			r := strings.NewReplacer("\r\n", "\\n", "\n", "\\n", "\r", "\\n")
			return r.Replace(text)
		},
		"lookup": func(text string) string {
			return d.Lookup(text)
		},
	}
}

func (s *Schema) SavePlantUml(wr io.Writer, distance int) error {
	o := NewPlantUML(s.Driver.Meta.Dict, distance)
	return o.OutputSchema(wr, s)
}
