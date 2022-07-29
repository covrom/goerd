package goerd

import (
	"fmt"
	"os"
	"strings"

	"github.com/covrom/goerd/schema"
	"github.com/jmoiron/sqlx"
)

type ObjectField[T any] struct {
	v   *T
	dbn string
	c   *schema.Column // nil == virtual
}

func (f *ObjectField[T]) dbName() string         { return f.dbn }
func (f *ObjectField[T]) column() *schema.Column { return f.c }
func (f *ObjectField[T]) new() any               { return new(T) }
func (f *ObjectField[T]) Column() *schema.Column { return f.c }

func Field[T any](column *schema.Column) *ObjectField[T] {
	return &ObjectField[T]{
		v:   new(T),
		dbn: column.Name,
		c:   column,
	}
}

func FieldVirtual[T any](dbName string) *ObjectField[T] {
	return &ObjectField[T]{
		v:   new(T),
		dbn: dbName,
		c:   nil,
	}
}

type ObjectModel[T any] struct {
	v      *T
	fields map[string]objectField
	tbls   *schema.Table
}

func (m *ObjectModel[T]) objectModel() interface{} {
	return m
}

func (m *ObjectModel[T]) Field(dbName string) objectField {
	return m.fields[dbName]
}

func (m *ObjectModel[T]) SchemaTable() *schema.Table {
	return m.tbls
}

func (m *ObjectModel[T]) WithType(t string) *ObjectModel[T] {
	m.tbls.Type = t
	return m
}

func (m *ObjectModel[T]) WithIndex(idxs ...*schema.Index) *ObjectModel[T] {
	m.tbls.Indexes = append(m.tbls.Indexes, idxs...)
	return m
}

func (m *ObjectModel[T]) WithConstraint(ctrs ...*schema.Constraint) *ObjectModel[T] {
	m.tbls.Constraints = append(m.tbls.Constraints, ctrs...)
	return m
}

func (m *ObjectModel[T]) WithComment(c string) *ObjectModel[T] {
	m.tbls.Comment = c
	return m
}

func (m *ObjectModel[T]) WithDef(def string) *ObjectModel[T] {
	m.tbls.Def = def
	return m
}

func (m *ObjectModel[T]) New() *T {
	return new(T)
}

type objectField interface {
	dbName() string
	column() *schema.Column
	new() any
}

var _ objectField = &ObjectField[int]{}

func Model[T any](table string, fields ...objectField) *ObjectModel[T] {
	t := &schema.Table{
		Name: table,
	}
	fmap := make(map[string]objectField, len(fields))
	for _, fld := range fields {
		if fc := fld.column(); fc != nil {
			t.Columns = append(t.Columns, fc)
		}
		fmap[fld.dbName()] = fld
	}

	ret := &ObjectModel[T]{
		v:      new(T),
		fields: fmap,
		tbls:   t,
	}

	return ret
}

type objectModel interface {
	objectModel() interface{}
	SchemaTable() *schema.Table
}

var _ objectModel = &ObjectModel[struct{}]{}

type ModelSet struct {
	ms   []objectModel
	rels []*schema.Relation
}

func NewModelSet(mds ...objectModel) *ModelSet {
	return &ModelSet{
		ms: mds,
	}
}

func (md *ModelSet) WithRelations(rels ...*schema.Relation) *ModelSet {
	md.rels = append(md.rels, rels...)
	return md
}

func (md *ModelSet) Migrate(d *sqlx.DB, dbSchema string) error {
	migsch := &schema.Schema{
		CurrentSchema: dbSchema,
		Relations:     md.rels,
	}
	for _, m := range md.ms {
		migsch.Tables = append(migsch.Tables, m.SchemaTable())
	}

	if err := migsch.Repair(); err != nil {
		return err
	}

	dbsch, err := SchemaFromPostgresDB(d.DB)
	if err != nil {
		return fmt.Errorf("cannot migrate database: %w", err)
	}

	qs := GenerateMigrationSQL(dbsch, migsch)
	tx, err := d.Begin()
	if err != nil {
		return fmt.Errorf("cannot migrate database: %w", err)
	}
	for i, q := range qs {
		// skip all dropping DDL queries
		if strings.HasPrefix(strings.ToUpper(q), "DROP") {
			fmt.Println(i+1, "skip: ", q)
			continue
		}
		if strings.Contains(strings.ToUpper(q), "DROP COLUMN") {
			fmt.Println(i+1, "skip: ", q)
			continue
		}

		fmt.Println(i+1, q)

		_, err = tx.Exec(q)

		if err != nil {
			_ = tx.Rollback()

			fmt.Println("db schema:")
			dbsch.SaveYaml(os.Stdout)
			fmt.Println("target schema:")
			migsch.SaveYaml(os.Stdout)

			return fmt.Errorf("cannot migrate database: %w", err)
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("cannot migrate database: %w", err)
	}
	return nil
}
