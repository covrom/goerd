package schema

import (
	"io"

	"github.com/goccy/go-yaml"
)

type YamlSchema struct {
	Name      string                `yaml:"name"`
	Tables    map[string]*YamlTable `yaml:"tables"`
	Relations []*YamlRelation       `yaml:"relations"`
}

type YamlTable struct {
	Type        string                     `yaml:"type,omitempty"`
	Columns     map[string]*YamlColumn     `yaml:"columns"`
	Indexes     map[string]*YamlIndex      `yaml:"indexes"`
	Constraints map[string]*YamlConstraint `yaml:"constraints"`
	Triggers    map[string]string          `yaml:"triggers"`
	Def         string                     `yaml:"def,omitempty"`
}

type YamlRelation struct {
	Table         string   `yaml:"table,omitempty"`
	Columns       []string `yaml:"columns,flow"`
	ParentTable   string   `yaml:"parentTable,omitempty"`
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
	Def         string   `yaml:"def,omitempty"`
	IsPrimary   bool     `yaml:"isPrimary,omitempty"`
	IsUnique    bool     `yaml:"isUnique,omitempty"`
	IsClustered bool     `yaml:"isClustered,omitempty"`
	MethodName  string   `yaml:"methodName,omitempty"`
	Columns     []string `yaml:"columns,flow"`
}

type YamlColumn struct {
	Type     string  `yaml:"type"`
	Nullable bool    `yaml:"nullable,omitempty"`
	Default  *string `yaml:"default,omitempty"`
}

func (s *Schema) MarshalYAML() ([]byte, error) {
	ys := &YamlSchema{
		Name:      s.Name,
		Tables:    make(map[string]*YamlTable, len(s.Tables)),
		Relations: make([]*YamlRelation, len(s.Relations)),
	}
	for _, t := range s.Tables {
		yt := &YamlTable{
			Def:         t.Def,
			Columns:     make(map[string]*YamlColumn, len(t.Columns)),
			Constraints: make(map[string]*YamlConstraint, len(t.Constraints)),
			Indexes:     make(map[string]*YamlIndex, len(t.Indexes)),
			Triggers:    make(map[string]string, len(t.Triggers)),
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
				IsClustered: idx.IsClustered,
				IsPrimary:   idx.IsPrimary,
				IsUnique:    idx.IsUnique,
				MethodName:  idx.MethodName,
				Def:         idx.Def,
				Columns:     idx.Columns,
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
		ys.Tables[t.Name] = yt
	}

	for i, r := range s.Relations {
		yr := &YamlRelation{
			Def:           r.Def,
			Columns:       make([]string, len(r.Columns)),
			ParentColumns: make([]string, len(r.ParentColumns)),
		}
		if r.Table != nil {
			yr.Table = r.Table.Name
		}
		if r.ParentTable != nil {
			yr.ParentTable = r.ParentTable.Name
		}
		for j, v := range r.Columns {
			yr.Columns[j] = v.Name
		}
		for j, v := range r.ParentColumns {
			yr.ParentColumns[j] = v.Name
		}
		ys.Relations[i] = yr
	}

	return yaml.Marshal(ys)
}

// // UnmarshalYAML unmarshal YAML to schema.Column
// func (c *Column) UnmarshalYAML(data []byte) error {
// 	s := struct {
// 		Name     string  `yaml:"name"`
// 		Type     string  `yaml:"type"`
// 		Nullable bool    `yaml:"nullable"`
// 		Default  *string `yaml:"default"`
// 		Comment  string  `yaml:"comment"`
// 	}{}
// 	err := yaml.Unmarshal(data, &s)
// 	if err != nil {
// 		return err
// 	}

// 	c.Name = s.Name
// 	c.Type = s.Type
// 	c.Nullable = s.Nullable
// 	if s.Default != nil {
// 		c.Default.Valid = true
// 		c.Default.String = *s.Default
// 	} else {
// 		c.Default.Valid = false
// 		c.Default.String = ""
// 	}
// 	c.Comment = s.Comment
// 	return nil
// }

// // UnmarshalYAML unmarshal YAML to schema.Column
// func (r *Relation) UnmarshalYAML(data []byte) error {
// 	s := struct {
// 		Table         string   `yaml:"table"`
// 		Columns       []string `yaml:"columns"`
// 		ParentTable   string   `yaml:"parentTable"`
// 		ParentColumns []string `yaml:"parentColumns"`
// 		Def           string   `yaml:"def"`
// 	}{}
// 	err := yaml.Unmarshal(data, &s)
// 	if err != nil {
// 		return err
// 	}
// 	r.Table = &Table{
// 		Name: s.Table,
// 	}
// 	r.Columns = []*Column{}
// 	for _, c := range s.Columns {
// 		r.Columns = append(r.Columns, &Column{
// 			Name: c,
// 		})
// 	}
// 	r.ParentTable = &Table{
// 		Name: s.ParentTable,
// 	}
// 	r.ParentColumns = []*Column{}
// 	for _, c := range s.ParentColumns {
// 		r.ParentColumns = append(r.ParentColumns, &Column{
// 			Name: c,
// 		})
// 	}
// 	r.Def = s.Def
// 	return nil
// }

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
