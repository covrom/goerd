package schema

import (
	"strings"
	"testing"
)

func TestPatchSchema_BuildDropAndNew(t *testing.T) {

	t.Run("1", func(t *testing.T) {
		from := &Schema{
			CurrentSchema: "public",
			Tables: []*Table{
				{
					Name: "table_old",
					Columns: []*Column{
						{
							Name:       "column1",
							Type:       "uuid",
							PrimaryKey: true,
						},
					},
				},
			},
		}

		to := &Schema{
			CurrentSchema: "public",
			Tables: []*Table{
				{
					Name: "table1",
					Columns: []*Column{
						{
							Name:       "column1",
							Type:       "uuid",
							PrimaryKey: true,
						},
						{
							Name: "column2",
							Type: "uuid",
						},
						{
							Name: "column3",
							Type: "uuid",
						},
					},
					Indexes: []*Index{
						{
							Name:    "table1_col2",
							Columns: []string{"column2"},
						},
						{
							Name:    "table1_col3",
							Columns: []string{"column3"},
						},
					},
					Constraints: []*Constraint{
						{
							Name:  "table1_constraint_check",
							Check: "true",
						},
					},
				},
			},
		}

		s := &PatchSchema{}
		s.Build(from, to)
		qs := s.GenerateSQL()
		qss := strings.Join(qs, "\n")
		if qss != `DROP TABLE IF EXISTS table_old
CREATE TABLE table1 (
column1 uuid NOT NULL PRIMARY KEY,
column2 uuid NOT NULL,
column3 uuid NOT NULL,
CONSTRAINT table1_constraint_check CHECK (true))
CREATE INDEX table1_col2 ON table1(column2)
CREATE INDEX table1_col3 ON table1(column3)` {
			t.Error(qss)
		}
	})

}

func TestPatchSchema_BuildAddColIdx(t *testing.T) {

	t.Run("1", func(t *testing.T) {
		from := &Schema{
			CurrentSchema: "public",
			Tables: []*Table{
				{
					Name: "table1",
					Columns: []*Column{
						{
							Name:       "column1",
							Type:       "uuid",
							PrimaryKey: true,
						},
						{
							Name: "column2",
							Type: "uuid",
						},
					},
					Indexes: []*Index{
						{
							Name:    "table1_col2",
							Columns: []string{"column2"},
						},
					},
				},
			},
		}

		to := &Schema{
			CurrentSchema: "public",
			Tables: []*Table{
				{
					Name: "table1",
					Columns: []*Column{
						{
							Name:       "column1",
							Type:       "uuid",
							PrimaryKey: true,
						},
						{
							Name: "column2",
							Type: "uuid",
						},
						{
							Name: "column3",
							Type: "uuid",
						},
					},
					Indexes: []*Index{
						{
							Name:    "table1_col2",
							Columns: []string{"column2"},
						},
						{
							Name:    "table1_col3",
							Columns: []string{"column3"},
						},
					},
					Constraints: []*Constraint{
						{
							Name:  "table1_constraint_check",
							Check: "true",
						},
					},
				},
			},
		}

		s := &PatchSchema{}
		s.Build(from, to)
		qs := s.GenerateSQL()
		qss := strings.Join(qs, "\n")
		if qss != `ALTER TABLE table1 ADD COLUMN column3 uuid NOT NULL
CREATE INDEX table1_col3 ON table1(column3)
ALTER TABLE table1 ADD CONSTRAINT table1_constraint_check CHECK (true)` {
			t.Error(qss)
		}
	})

}

func TestPatchSchema_BuildEq(t *testing.T) {

	t.Run("1", func(t *testing.T) {
		from := &Schema{
			CurrentSchema: "public",
			Tables: []*Table{
				{
					Name: "table1",
					Columns: []*Column{
						{
							Name:       "column1",
							Type:       "uuid",
							PrimaryKey: true,
						},
						{
							Name: "column2",
							Type: "uuid",
						},
					},
					Indexes: []*Index{
						{
							Name:    "table1_col2",
							Columns: []string{"column2"},
						},
					},
				},
			},
		}

		to := &Schema{
			CurrentSchema: "public",
			Tables: []*Table{
				{
					Name: "table1",
					Columns: []*Column{
						{
							Name:       "column1",
							Type:       "uuid",
							PrimaryKey: true,
						},
						{
							Name: "column2",
							Type: "uuid",
						},
					},
					Indexes: []*Index{
						{
							Name:    "table1_col2",
							Columns: []string{"column2"},
						},
					},
				},
			},
		}

		s := &PatchSchema{}
		s.Build(from, to)
		qs := s.GenerateSQL()
		qss := strings.Join(qs, "\n")
		if qss != `` {
			t.Error(qss)
		}
	})

}

func TestPatchSchema_BuildChangeCol(t *testing.T) {

	t.Run("1", func(t *testing.T) {
		from := &Schema{
			CurrentSchema: "public",
			Tables: []*Table{
				{
					Name: "table1",
					Columns: []*Column{
						{
							Name:       "column1",
							Type:       "uuid",
							PrimaryKey: true,
						},
						{
							Name: "column2",
							Type: "uuid",
						},
					},
					Indexes: []*Index{
						{
							Name:    "table1_col2",
							Columns: []string{"column2"},
						},
					},
				},
			},
		}

		to := &Schema{
			CurrentSchema: "public",
			Tables: []*Table{
				{
					Name: "table1",
					Columns: []*Column{
						{
							Name:       "column1",
							Type:       "uuid",
							PrimaryKey: true,
						},
						{
							Name: "column2",
							Type: "text",
						},
					},
					Indexes: []*Index{
						{
							Name:    "table1_col2",
							Columns: []string{"column2"},
						},
					},
				},
			},
		}

		s := &PatchSchema{}
		s.Build(from, to)
		qs := s.GenerateSQL()
		qss := strings.Join(qs, "\n")
		if qss != `ALTER TABLE table1 ALTER COLUMN column2 TYPE text` {
			t.Error(qss)
		}
	})

}

func TestPatchSchema_BuildChangeIdx(t *testing.T) {

	t.Run("1", func(t *testing.T) {
		from := &Schema{
			CurrentSchema: "public",
			Tables: []*Table{
				{
					Name: "table1",
					Columns: []*Column{
						{
							Name:       "column1",
							Type:       "uuid",
							PrimaryKey: true,
						},
						{
							Name: "column2",
							Type: "uuid",
						},
					},
					Indexes: []*Index{
						{
							Name:    "table1_col2",
							Columns: []string{"column2"},
						},
					},
				},
			},
		}

		to := &Schema{
			CurrentSchema: "public",
			Tables: []*Table{
				{
					Name: "table1",
					Columns: []*Column{
						{
							Name:       "column1",
							Type:       "uuid",
							PrimaryKey: true,
						},
						{
							Name: "column2",
							Type: "uuid",
						},
						{
							Name: "column3",
							Type: "uuid",
						},
					},
					Indexes: []*Index{
						{
							Name:    "table1_col2",
							Columns: []string{"column2,column3"},
						},
					},
				},
			},
		}

		s := &PatchSchema{}
		s.Build(from, to)
		qs := s.GenerateSQL()
		qss := strings.Join(qs, "\n")
		if qss != `ALTER TABLE table1 ADD COLUMN column3 uuid NOT NULL
DROP INDEX IF EXISTS table1_col2
CREATE INDEX table1_col2 ON table1(column2,column3)` {
			t.Error(qss)
		}
	})

}

func TestPatchSchema_BuildChangeIdx2(t *testing.T) {
	from := &Schema{
		CurrentSchema: "public",
		Tables: []*Table{
			{
				Name: "projects",
				Type: "TABLE",
				Columns: []*Column{
					{
						Name:       "id",
						Type:       "uuid",
						PrimaryKey: true,
					},
					{
						Name:     "deleted_at",
						Type:     "timestamptz",
						Nullable: true,
					},
				},
				Indexes: []*Index{
					{
						Name:       "projects_deleted_at",
						MethodName: "btree",
						Columns:    []string{"deleted_at"},
					},
				},
			},
		},
	}

	to := &Schema{
		CurrentSchema: "public",
		Tables: []*Table{
			{
				Name: "projects",
				Type: "TABLE",
				Columns: []*Column{
					{
						Name:       "id",
						Type:       "uuid",
						PrimaryKey: true,
					},
					{
						Name:     "deleted_at",
						Type:     "timestamptz",
						Nullable: true,
					},
				},
				Indexes: []*Index{
					{
						Name:    "projects_deleted_at",
						Columns: []string{"deleted_at"},
					},
				},
			},
		},
	}

	t.Run("1", func(t *testing.T) {
		s := &PatchSchema{}
		s.Build(from, to)
		qs := s.GenerateSQL()
		qss := strings.Join(qs, "\n")
		if qss != `` {
			t.Error(qss)
		}
	})

}
