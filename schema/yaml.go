package schema

import (
	"io"

	"github.com/goccy/go-yaml"
)

type YamlSchema struct {
	Name   string                `yaml:"name"`
	Schema string                `yaml:"schema"`
	Tables map[string]*YamlTable `yaml:"tables"`
}

type YamlTable struct {
	Type        string                     `yaml:"type,omitempty"`
	Columns     map[string]*YamlColumn     `yaml:"columns"`
	Indexes     map[string]*YamlIndex      `yaml:"indexes"`
	Constraints map[string]*YamlConstraint `yaml:"constraints"`
	Relations   map[string]*YamlRelation   `yaml:"relations"` // key = parent table
	Triggers    map[string]string          `yaml:"triggers"`
	Def         string                     `yaml:"def,omitempty"`
}

type YamlRelation struct {
	Columns       []string `yaml:"columns,flow"`
	ParentColumns []string `yaml:"parentColumns,flow"`
	Def           string   `yaml:"def,omitempty"`
}

type YamlConstraint struct {
	Type             string   `yaml:"type,omitempty"`
	Def              string   `yaml:"def,omitempty"`
	Table            *string  `yaml:"table,omitempty"`
	ReferenceTable   *string  `yaml:"referenceTable,omitempty"`
	Columns          []string `yaml:"columns,flow"`
	ReferenceColumns []string `yaml:"referenceColumns,flow"`
}

type YamlIndex struct {
	IsPrimary    bool     `yaml:"isPrimary,omitempty"`
	IsUnique     bool     `yaml:"isUnique,omitempty"`
	IsClustered  bool     `yaml:"isClustered,omitempty"`
	Concurrently bool     `yaml:"concurrently,omitempty"`
	MethodName   string   `yaml:"methodName,omitempty"`
	Columns      []string `yaml:"columns,flow"`
	ColDef       string   `yaml:"coldef,omitempty"`
	With         string   `yaml:"with,omitempty"`
	Tablespace   string   `yaml:"tablespace,omitempty"`
	Where        string   `yaml:"where,omitempty"`
}

type YamlColumn struct {
	Type     string  `yaml:"type"`
	Nullable bool    `yaml:"nullable,omitempty"`
	Default  *string `yaml:"default,omitempty"`
}

func (s *Schema) MarshalYAML() ([]byte, error) {
	ys := &YamlSchema{
		Name:   s.Name,
		Schema: s.Driver.Meta.CurrentSchema,
		Tables: make(map[string]*YamlTable, len(s.Tables)),
	}
	for _, t := range s.Tables {
		yt := &YamlTable{
			Def:         t.Def,
			Columns:     make(map[string]*YamlColumn, len(t.Columns)),
			Constraints: make(map[string]*YamlConstraint, len(t.Constraints)),
			Indexes:     make(map[string]*YamlIndex, len(t.Indexes)),
			Triggers:    make(map[string]string, len(t.Triggers)),
			Relations:   make(map[string]*YamlRelation, len(t.Constraints)),
			Type:        t.Type,
		}
		var defval *string
		for _, c := range t.Columns {
			defval = nil
			if c.Default.Valid {
				defval = &(c.Default.String)
			}
			yt.Columns[c.Name] = &YamlColumn{
				Type:     c.Type,
				Default:  defval,
				Nullable: c.Nullable,
			}
		}
		for _, idx := range t.Indexes {
			yt.Indexes[idx.Name] = &YamlIndex{
				IsClustered:  idx.IsClustered,
				IsPrimary:    idx.IsPrimary,
				IsUnique:     idx.IsUnique,
				MethodName:   idx.MethodName,
				Columns:      idx.Columns,
				With:         idx.With,
				Where:        idx.Where,
				Concurrently: idx.Concurrently,
				ColDef:       idx.ColDef,
				Tablespace:   idx.Tablespace,
			}
		}
		for _, cs := range t.Constraints {
			yt.Constraints[cs.Name] = &YamlConstraint{
				Type:             cs.Type,
				Def:              cs.Def,
				Table:            cs.Table,
				ReferenceTable:   cs.ReferenceTable,
				Columns:          cs.Columns,
				ReferenceColumns: cs.ReferenceColumns,
			}
		}
		for _, tr := range t.Triggers {
			yt.Triggers[tr.Name] = tr.Def
		}
		for _, r := range s.Relations {
			if r.Table.Name != t.Name {
				continue
			}
			yr := &YamlRelation{
				Def:           r.Def,
				Columns:       make([]string, len(r.Columns)),
				ParentColumns: make([]string, len(r.ParentColumns)),
			}
			for j, v := range r.Columns {
				yr.Columns[j] = v.Name
			}
			for j, v := range r.ParentColumns {
				yr.ParentColumns[j] = v.Name
			}
			yt.Relations[r.ParentTable.Name] = yr
		}
		ys.Tables[t.Name] = yt
	}

	return yaml.Marshal(ys)
}

func (s *Schema) UnmarshalYAML(data []byte) error {
	ys := &YamlSchema{}
	if err := yaml.Unmarshal(data, &ys); err != nil {
		return err
	}
	*s = Schema{}
	s.Tables = make([]*Table, 0, len(ys.Tables))
	for name, yt := range ys.Tables {
		t := &Table{
			Name: name,
			Type: yt.Type,
			Def:  yt.Def,
		}

		s.Tables = append(s.Tables, t)
	}
	return nil
}

// YAML struct
type YAML struct{}

// OutputSchema output YAML format for full relation.
func (j *YAML) OutputSchema(wr io.Writer, s *Schema) error {
	encoder := yaml.NewEncoder(wr)
	err := encoder.Encode(s)
	if err != nil {
		return err
	}
	return nil
}

// OutputTable output YAML format for table.
func (j *YAML) OutputTable(wr io.Writer, t *Table) error {
	encoder := yaml.NewEncoder(wr)
	err := encoder.Encode(t)
	if err != nil {
		return err
	}
	return nil
}

func (s *Schema) SaveYaml(wr io.Writer) error {
	o := new(YAML)
	return o.OutputSchema(wr, s)
}
