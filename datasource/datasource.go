package datasource

import (
	"database/sql"

	"github.com/covrom/goerd/drivers/postgres"
	"github.com/covrom/goerd/schema"
	"github.com/pkg/errors"
)

// Analyze database
func Analyze(urlstr string) (*schema.Schema, error) {
	s := &schema.Schema{}
	db, err := sql.Open("pgx", urlstr)
	if err != nil {
		return s, errors.WithStack(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return s, errors.WithStack(err)
	}

	driver := postgres.New(db)
	err = driver.Analyze(s)
	if err != nil {
		return s, err
	}

	return s, nil
}
