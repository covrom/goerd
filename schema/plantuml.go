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

// TODO: add indexes, constraints and triggers

func (p *PlantUML) schemaTemplate() string {
	return `@startuml
{{ $sc := .showComment -}}
!define table(name, desc) entity name as "desc" << (T,#5DBCD2) >>
!define view(name, desc) entity name as "desc" << (V,#C6EDDB) >>
!define column(name, type, desc) name <font color="#666666">[type]</font><font color="#333333">desc</font>
hide methods
hide stereotypes

skinparam class {
	BackgroundColor White
	BorderColor #6E6E6E
	ArrowColor #6E6E6E
}

' tables
{{- range $i, $t := .Schema.Tables }}
{{- if ne $t.Type "VIEW" }}
table("{{ $t.Name }}", "{{ $t.Name }}{{ if $sc }}{{ if ne $t.Comment "" }}\n{{ $t.Comment | html | escape_nl }}{{ end }}{{ end }}") {
{{- else }}
view("{{ $t.Name }}", "{{ $t.Name }}{{ if $sc }}{{ if ne $t.Comment "" }}\n{{ $t.Comment | html | escape_nl }}{{ end }}{{ end }}") {
{{- end }}
{{- range $ii, $c := $t.Columns }}
	column("{{ $c.Name | html }}", "{{ $c.Type | html }}", "{{ if $sc }}{{ if ne $c.Comment "" }} {{ $c.Comment | html | nl2space }}{{ end }}{{ end }}")
{{- end }}
}
{{- end }}

' relations
{{- range $j, $r := .Schema.Relations }}
"{{ $r.Table.Name }}" }-- "{{ $r.ParentTable.Name }}" : "{{ $r.Def | html }}"
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
