package datasource

import (
	"fmt"
	"strings"

	"github.com/covrom/goerd/drivers"
	"github.com/covrom/goerd/drivers/postgres"
	"github.com/covrom/goerd/schema"
	"github.com/pkg/errors"
	"github.com/xo/dburl"
)

// Analyze database
func Analyze(urlstr string) (*schema.Schema, error) {
	s := &schema.Schema{}
	u, err := dburl.Parse(urlstr)
	if err != nil {
		return s, errors.WithStack(err)
	}
	splitted := strings.Split(u.Short(), "/")
	if len(splitted) < 2 {
		return s, errors.WithStack(fmt.Errorf("invalid DSN: parse %s -> %#v", urlstr, u))
	}

	db, err := dburl.Open(urlstr)
	defer db.Close()
	if err != nil {
		return s, errors.WithStack(err)
	}
	if err := db.Ping(); err != nil {
		return s, errors.WithStack(err)
	}

	var driver drivers.Driver

	switch u.Driver {
	case "postgres":
		s.Name = splitted[1]
		driver = postgres.New(db)
	default:
		return s, errors.WithStack(fmt.Errorf("unsupported driver '%s'", u.Driver))
	}
	err = driver.Analyze(s)
	if err != nil {
		return s, err
	}
	return s, nil
}
