package goerd

import (
	"database/sql"
	"io"

	"github.com/covrom/goerd/drivers/postgres"
	"github.com/covrom/goerd/schema"
)

// SchemaFromPostgresDB reads database schema from postgres *sql.DB
func SchemaFromPostgresDB(db *sql.DB) (*schema.Schema, error) {
	s := &schema.Schema{}
	driver := postgres.New(db)
	err := driver.Analyze(s)
	return s, err
}

// GenerateMigrationSQL generates an array of SQL DDL queries
// for postgres that modify database tables, columns, indexes, etc.
func GenerateMigrationSQL(sfrom, sto *schema.Schema) []string {
	ptch := &schema.PatchSchema{CurrentSchema: sfrom.CurrentSchema}
	ptch.Build(sfrom, sto)
	return ptch.GenerateSQL()
}

// SchemaToYAML saves the schema to a yaml file
func SchemaToYAML(s *schema.Schema, w io.Writer) error {
	return s.SaveYaml(w)
}

// SchemaFromYAML loads the schema from the yaml file
func SchemaFromYAML(r io.Reader) (*schema.Schema, error) {
	s := &schema.Schema{}
	if err := s.LoadYaml(r); err != nil {
		return nil, err
	}
	return s, nil
}
