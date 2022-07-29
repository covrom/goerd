package schema

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

const (
	TypeFK = "FOREIGN KEY"
	TypePK = "PRIMARY KEY"
	TypeUQ = "UNIQUE"
)

// Table is the struct for database table
type Table struct {
	Name        string        `json:"name"`
	Type        string        `json:"type"`
	Comment     string        `json:"comment"`
	Columns     []*Column     `json:"columns"`
	Indexes     []*Index      `json:"indexes"`
	Constraints []*Constraint `json:"constraints"`
	Def         string        `json:"def"`
}

func (t *Table) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("table name not defined")
	}
	if len(t.Columns) == 0 {
		return fmt.Errorf("table columns not defined")
	}
	for _, c := range t.Columns {
		if err := c.Validate(); err != nil {
			return err
		}
	}
	for _, idx := range t.Indexes {
		if err := idx.Validate(); err != nil {
			return err
		}
	}
	for _, c := range t.Constraints {
		if err := c.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Index is the struct for database index
type Index struct {
	Name         string `json:"name"`
	IsPrimary    bool
	IsUnique     bool
	IsClustered  bool
	MethodName   string
	Def          string   `json:"def"`
	Table        *string  `json:"table"`
	Columns      []string `json:"columns"`
	Concurrently bool     `json:"concurrently,omitempty"`
	ColDef       string   `json:"coldef,omitempty"`
	With         string   `json:"with,omitempty"`
	Tablespace   string   `json:"tablespace,omitempty"`
	Where        string   `json:"where,omitempty"`
	Comment      string   `json:"comment"`
}

func (idx *Index) Validate() error {
	if idx.Name == "" {
		return fmt.Errorf("index name not defined")
	}
	if idx.Table == nil {
		return fmt.Errorf("index table not defined")
	}
	if idx.ColDef == "" && len(idx.Columns) == 0 {
		return fmt.Errorf("index columns not defined")
	}
	return nil
}

// Constraint is the struct for database constraint
type Constraint struct {
	Name             string   `json:"name"`
	Type             string   `json:"type"`
	Def              string   `json:"def"`
	Check            string   `json:"check"`
	OnDelete         string   `json:"onDelete"`
	Table            *string  `json:"table"`
	ReferenceTable   *string  `json:"reference_table" yaml:"referenceTable"`
	Columns          []string `json:"columns"`
	ReferenceColumns []string `json:"reference_columns" yaml:"referenceColumns"`
	Comment          string   `json:"comment"`
}

func (c *Constraint) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("constraint name not defined")
	}
	if c.Table == nil {
		return fmt.Errorf("constraint table not defined")
	}
	if len(c.Check) > 0 {
		return nil
	}
	if c.Type == "" {
		return fmt.Errorf("constraint type not defined")
	}
	if len(c.Columns) == 0 {
		return fmt.Errorf("relation columns not defined")
	}
	if c.Type == TypeFK && len(c.ReferenceColumns) == 0 {
		return fmt.Errorf("FK-relation reference columns not defined")
	}
	if c.Type == TypeFK && c.ReferenceTable == nil {
		return fmt.Errorf("FK-relation reference table not defined")
	}
	return nil
}

// Column is the struct for table column
type Column struct {
	Name            string         `json:"name"`
	Type            string         `json:"type"`
	Nullable        bool           `json:"nullable"`
	PrimaryKey      bool           `json:"pk"`
	Default         sql.NullString `json:"default"`
	Comment         string         `json:"comment"`
	ParentRelations []*Relation    `json:"-"`
	ChildRelations  []*Relation    `json:"-"`
}

func (c *Column) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("column name not defined")
	}
	if c.Type == "" {
		return fmt.Errorf("column type not defined")
	}
	return nil
}

// Relation is the struct for table relation
type Relation struct {
	Name          string    `json:"name"`
	Table         *Table    `json:"table"`
	Columns       []*Column `json:"columns"`
	ParentTable   *Table    `json:"parent_table" yaml:"parentTable"`
	ParentColumns []*Column `json:"parent_columns" yaml:"parentColumns"`
	OnDelete      string    `json:"onDelete"`
	Def           string    `json:"def"`
}

func (r *Relation) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("relation name not defined")
	}
	if r.Table == nil {
		return fmt.Errorf("relation table not defined")
	}
	if len(r.Columns) == 0 {
		return fmt.Errorf("relation columns not defined")
	}
	if r.ParentTable == nil {
		return fmt.Errorf("relation parent table not defined")
	}
	if len(r.ParentColumns) == 0 {
		return fmt.Errorf("relation parent columns not defined")
	}
	return nil
}

// Schema is the struct for database schema
type Schema struct {
	Name          string      `json:"name"`
	Desc          string      `json:"desc"`
	Tables        []*Table    `json:"tables"`
	Relations     []*Relation `json:"relations"`
	CurrentSchema string      `json:"currentSchema"`
	SearchPaths   []string    `json:"searchPaths,omitempty"`
}

func (s *Schema) Validate() error {
	for _, t := range s.Tables {
		if err := t.Validate(); err != nil {
			return fmt.Errorf("table %q validation error: %w", t.Name, err)
		}
	}
	for _, r := range s.Relations {
		if err := r.Validate(); err != nil {
			return fmt.Errorf("relation %q validation error: %w", r.Name, err)
		}
	}
	return nil
}

func (s *Schema) NormalizeTableName(name string) string {
	if !strings.Contains(name, ".") {
		return fmt.Sprintf("%s.%s", s.CurrentSchema, name)
	}
	return name
}

func (s *Schema) NormalizeTableNames(names []string) []string {
	for i, n := range names {
		names[i] = s.NormalizeTableName(n)
	}
	return names
}

// FindTableByName find table by table name
func (s *Schema) FindTableByName(name string) (*Table, error) {
	for _, t := range s.Tables {
		if s.NormalizeTableName(t.Name) == s.NormalizeTableName(name) {
			return t, nil
		}
	}
	return nil, errors.WithStack(fmt.Errorf("not found table '%s'", name))
}

// FindRelation ...
func (s *Schema) FindRelation(tblName string, cs, pcs []*Column) (*Relation, error) {
L:
	for _, r := range s.Relations {
		if len(r.Columns) != len(cs) || len(r.ParentColumns) != len(pcs) || r.Table.Name != tblName {
			continue
		}
		for _, rc := range r.Columns {
			exist := false
			for _, cc := range cs {
				if rc.Name == cc.Name {
					exist = true
				}
			}
			if !exist {
				continue L
			}
		}
		for _, rc := range r.ParentColumns {
			exist := false
			for _, cc := range pcs {
				if rc.Name == cc.Name {
					exist = true
				}
			}
			if !exist {
				continue L
			}
		}
		return r, nil
	}
	return nil, errors.WithStack(fmt.Errorf("not found relation '%v, %v'", cs, pcs))
}

// FindColumnByName find column by column name
func (t *Table) FindColumnByName(name string) (*Column, error) {
	for _, c := range t.Columns {
		if c.Name == name {
			return c, nil
		}
	}
	return nil, errors.WithStack(fmt.Errorf("not found column '%s' on table '%s'", name, t.Name))
}

// FindIndexByName find index by index name
func (t *Table) FindIndexByName(name string) (*Index, error) {
	for _, i := range t.Indexes {
		if i.Name == name {
			return i, nil
		}
	}
	return nil, errors.WithStack(fmt.Errorf("not found index '%s' on table '%s'", name, t.Name))
}

// FindConstraintByName find constraint by constraint name
func (t *Table) FindConstraintByName(name string) (*Constraint, error) {
	for _, c := range t.Constraints {
		if c.Name == name {
			return c, nil
		}
	}
	return nil, errors.WithStack(fmt.Errorf("not found constraint '%s' on table '%s'", name, t.Name))
}

// FindConstrainsByColumnName find constraint by column name
func (t *Table) FindConstrainsByColumnName(name string) []*Constraint {
	cts := []*Constraint{}
	for _, ct := range t.Constraints {
		for _, ctc := range ct.Columns {
			if ctc == name {
				cts = append(cts, ct)
			}
		}
	}
	return cts
}

// Sort schema tables, columns, relations, and constrains
func (s *Schema) Sort() {
	for _, t := range s.Tables {
		for _, c := range t.Columns {
			sort.SliceStable(c.ParentRelations, func(i, j int) bool {
				return c.ParentRelations[i].Table.Name < c.ParentRelations[j].Table.Name
			})
			sort.SliceStable(c.ChildRelations, func(i, j int) bool {
				return c.ChildRelations[i].Table.Name < c.ChildRelations[j].Table.Name
			})
		}
		sort.SliceStable(t.Columns, func(i, j int) bool {
			return t.Columns[i].Name < t.Columns[j].Name
		})
		sort.SliceStable(t.Indexes, func(i, j int) bool {
			return t.Indexes[i].Name < t.Indexes[j].Name
		})
		for _, idx := range t.Indexes {
			sort.Strings(idx.Columns)
		}
		sort.SliceStable(t.Constraints, func(i, j int) bool {
			return t.Constraints[i].Name < t.Constraints[j].Name
		})
		for _, cs := range t.Constraints {
			sort.Strings(cs.Columns)
			sort.Strings(cs.ReferenceColumns)
		}
	}
	sort.SliceStable(s.Tables, func(i, j int) bool {
		return s.Tables[i].Name < s.Tables[j].Name
	})
	sort.SliceStable(s.Relations, func(i, j int) bool {
		return s.Relations[i].Table.Name < s.Relations[j].Table.Name
	})
	for _, r := range s.Relations {
		sort.SliceStable(r.Columns, func(i, j int) bool {
			return r.Columns[i].Name < r.Columns[j].Name
		})
		sort.SliceStable(r.ParentColumns, func(i, j int) bool {
			return r.ParentColumns[i].Name < r.ParentColumns[j].Name
		})
	}
}

// Repair column relations
func (s *Schema) Repair() error {
	for _, t := range s.Tables {
		if len(t.Columns) == 0 {
			t.Columns = nil
		}
		if len(t.Indexes) == 0 {
			t.Indexes = nil
		}
		if len(t.Constraints) == 0 {
			t.Constraints = nil
		}
	}

	for _, r := range s.Relations {
		t, err := s.FindTableByName(r.Table.Name)
		if err != nil {
			return errors.Wrap(err, "failed to repair relation")
		}
		for i, rc := range r.Columns {
			c, err := t.FindColumnByName(rc.Name)
			if err != nil {
				return errors.Wrap(err, "failed to repair relation")
			}
			c.ParentRelations = append(c.ParentRelations, r)
			r.Columns[i] = c
		}
		r.Table = t
		pt, err := s.FindTableByName(r.ParentTable.Name)
		if err != nil {
			return errors.Wrap(err, "failed to repair relation")
		}
		for i, rc := range r.ParentColumns {
			pc, err := pt.FindColumnByName(rc.Name)
			if err != nil {
				return errors.Wrap(err, "failed to repair relation")
			}
			pc.ChildRelations = append(pc.ChildRelations, r)
			r.ParentColumns[i] = pc
		}
		r.ParentTable = pt
	}
	return nil
}

func (t *Table) CollectTablesAndRelations(distance int, root bool) ([]*Table, []*Relation, error) {
	tables := []*Table{}
	relations := []*Relation{}
	tables = append(tables, t)
	if distance == 0 {
		return tables, relations, nil
	}
	distance = distance - 1
	for _, c := range t.Columns {
		for _, r := range c.ParentRelations {
			relations = append(relations, r)
			ts, rs, err := r.ParentTable.CollectTablesAndRelations(distance, false)
			if err != nil {
				return nil, nil, err
			}
			tables = append(tables, ts...)
			relations = append(relations, rs...)
		}
		for _, r := range c.ChildRelations {
			relations = append(relations, r)
			ts, rs, err := r.Table.CollectTablesAndRelations(distance, false)
			if err != nil {
				return nil, nil, err
			}
			tables = append(tables, ts...)
			relations = append(relations, rs...)
		}
	}

	if !root {
		return tables, relations, nil
	}

	uTables := []*Table{}
	encounteredT := make(map[string]bool)
	for _, t := range tables {
		if !encounteredT[t.Name] {
			encounteredT[t.Name] = true
			uTables = append(uTables, t)
		}
	}

	uRelations := []*Relation{}
	encounteredR := make(map[*Relation]bool)
	for _, r := range relations {
		if !encounteredR[r] {
			encounteredR[r] = true
			if !encounteredT[r.ParentTable.Name] || !encounteredT[r.Table.Name] {
				continue
			}
			uRelations = append(uRelations, r)
		}
	}

	return uTables, uRelations, nil
}
