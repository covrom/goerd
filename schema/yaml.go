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
	Indexes     map[string]*YamlIndex      `yaml:"indexes,omitempty"`
	Constraints map[string]*YamlConstraint `yaml:"constraints,omitempty"`
	Relations   map[string]*YamlRelation   `yaml:"relations,omitempty"` // key = parent table
	Def         string                     `yaml:"def,omitempty"`
}

type YamlRelation struct {
	Columns       []string `yaml:"columns,flow"`
	ParentColumns []string `yaml:"parentColumns,flow"`
	OnDelete      string   `yaml:"onDelete,omitempty"`
}

type YamlConstraint struct {
	Type             string   `yaml:"type,omitempty"`
	Check            string   `json:"check,omitempty"`
	OnDelete         string   `json:"onDelete,omitempty"`
	ReferenceTable   string   `yaml:"referenceTable,omitempty"`
	Columns          []string `yaml:"columns,flow"`
	ReferenceColumns []string `yaml:"referenceColumns,flow,omitempty"`
}

type YamlIndex struct {
	IsPrimary    bool     `yaml:"isPrimary,omitempty"`
	IsUnique     bool     `yaml:"isUnique,omitempty"`
	IsClustered  bool     `yaml:"isClustered,omitempty"`
	Concurrently bool     `yaml:"concurrently,omitempty"`
	MethodName   string   `yaml:"method,omitempty"`
	Columns      []string `yaml:"columns,flow"`
	ColDef       string   `yaml:"coldef,omitempty"`
	With         string   `yaml:"with,omitempty"`
	Tablespace   string   `yaml:"tablespace,omitempty"`
	Where        string   `yaml:"where,omitempty"`
}

type YamlColumn struct {
	Type       string  `yaml:"type"`
	Nullable   bool    `yaml:"nullable,omitempty"`
	PrimaryKey bool    `yaml:"pk,omitempty"`
	Default    *string `yaml:"default,omitempty"`
}

func (s *Schema) MarshalYAML() ([]byte, error) {
	ys := &YamlSchema{
		Name:   s.Name,
		Schema: s.CurrentSchema,
		Tables: make(map[string]*YamlTable, len(s.Tables)),
	}
	for _, t := range s.Tables {
		yt := &YamlTable{
			Def:         t.Def,
			Columns:     make(map[string]*YamlColumn, len(t.Columns)),
			Constraints: make(map[string]*YamlConstraint, len(t.Constraints)),
			Indexes:     make(map[string]*YamlIndex, len(t.Indexes)),
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
				Type:       c.Type,
				Default:    defval,
				Nullable:   c.Nullable,
				PrimaryKey: c.PrimaryKey,
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
			if cs.Type == TypePK {
				// present as 'pk: true' in columns
				continue
			}
			ycs := &YamlConstraint{
				Type:             cs.Type,
				Check:            cs.Check,
				OnDelete:         cs.OnDelete,
				Columns:          cs.Columns,
				ReferenceColumns: cs.ReferenceColumns,
			}
			if cs.ReferenceTable != nil {
				ycs.ReferenceTable = *cs.ReferenceTable
			} else {
				ycs.ReferenceColumns = nil
			}
			yt.Constraints[cs.Name] = ycs
		}
		for _, r := range s.Relations {
			if r.Table.Name != t.Name {
				continue
			}
			yr := &YamlRelation{
				OnDelete:      r.OnDelete,
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
			Name:        name,
			Type:        yt.Type,
			Def:         yt.Def,
			Columns:     make([]*Column, 0, len(yt.Columns)),
			Indexes:     make([]*Index, 0, len(yt.Indexes)),
			Constraints: make([]*Constraint, 0, len(yt.Constraints)),
		}

		// TODO:

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
