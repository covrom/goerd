package goerd

import (
	"database/sql"
	"io"

	"github.com/covrom/goerd/drivers/postgres"
	"github.com/covrom/goerd/schema"
)

func SchemaFromPostgresDB(db *sql.DB) (*schema.Schema, error) {
	s := &schema.Schema{}
	driver := postgres.New(db)
	err := driver.Analyze(s)
	return s, err
}

func GenerateMigrationSQL(sfrom, sto *schema.Schema) []string {
	ptch := &schema.PatchSchema{CurrentSchema: sfrom.CurrentSchema}
	ptch.Build(sfrom, sto)
	return ptch.GenerateSQL()
}

func SchemaToYAML(s *schema.Schema, w io.Writer) error {
	return s.SaveYaml(w)
}

func SchemaFromYAML(r io.Reader) (*schema.Schema, error) {
	s := &schema.Schema{}
	if err := s.LoadYaml(r); err != nil {
		return nil, err
	}
	return s, nil
}
