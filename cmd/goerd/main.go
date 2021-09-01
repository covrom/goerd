package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/covrom/goerd"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var (
	from    = flag.String("from", "", "source schema filename *.yaml, or source PostgreSQL database DSN, e.g. postgres://username:password@url:port/dbName")
	to      = flag.String("to", "", "target schema filename *.yaml or target PostgreSQL database DSN, or *.puml for save source to plantuml")
	command = flag.String("c", "print", "command: 'print' - stdout print diff queries, 'apply' - apply diff queries to target database ('to')")
	dist    = flag.Int("d", 2, "max relations distance for plant uml")
	drop    = flag.Bool("drop", false, "drop tables or columns when applying migration")
)

func main() {
	flag.Parse()
	if *from == "" || *to == "" {
		flag.Usage()
		return
	}

	srcIsYaml := strings.HasSuffix(strings.ToLower(*from), ".yaml") || strings.HasSuffix(strings.ToLower(*from), ".yml")
	srcIsPg := strings.HasPrefix(strings.ToLower(*from), "postgres://")
	dstIsYaml := strings.HasSuffix(strings.ToLower(*to), ".yaml") || strings.HasSuffix(strings.ToLower(*to), ".yml")
	dstIsPg := strings.HasPrefix(strings.ToLower(*to), "postgres://")
	dstIsPuml := strings.HasSuffix(strings.ToLower(*to), ".puml")
	cmdIsPrint := *command == "print"
	cmdIsApply := *command == "apply"

	switch {
	case srcIsYaml && dstIsYaml:
		f, err := os.Open(*from)
		if err != nil {
			log.Fatal(err)
		}
		src, err := goerd.SchemaFromYAML(f)
		if err != nil {
			f.Close()
			log.Fatal(err)
		}
		f.Close()

		f, err = os.Open(*to)
		if err != nil {
			log.Fatal(err)
		}
		dst, err := goerd.SchemaFromYAML(f)
		if err != nil {
			f.Close()
			log.Fatal(err)
		}
		f.Close()

		qs := goerd.GenerateMigrationSQL(src, dst)
		if cmdIsPrint {
			for _, q := range qs {
				if !*drop {
					if strings.HasPrefix(strings.ToUpper(q), "DROP") {
						fmt.Println("--", q)
						continue
					}
					if strings.Contains(strings.ToUpper(q), "DROP COLUMN") {
						fmt.Println("--", q)
						continue
					}
				}
				fmt.Println(q)
			}
		} else if cmdIsApply {
			log.Fatal("cant apply diffs between two yaml schemas, only print allowed")
		} else {
			log.Fatal("wrong command")
		}

	case srcIsYaml && dstIsPuml:
		f, err := os.Open(*from)
		if err != nil {
			log.Fatal(err)
		}
		src, err := goerd.SchemaFromYAML(f)
		if err != nil {
			f.Close()
			log.Fatal(err)
		}
		f.Close()

		wr, err := os.OpenFile(*to, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatal(err)
		}

		if err := goerd.SchemaToPlantUML(src, wr, *dist); err != nil {
			wr.Close()
			log.Fatal(err)
		}
		wr.Close()

	case srcIsYaml && dstIsPg:
		f, err := os.Open(*from)
		if err != nil {
			log.Fatal(err)
		}
		// here from is destination schema, migrate to it
		dst, err := goerd.SchemaFromYAML(f)
		if err != nil {
			f.Close()
			log.Fatal(err)
		}
		f.Close()

		src, err := goerd.SchemaFromPostgresWithConnect(*to)
		if err != nil {
			log.Fatal(err)
		}

		qs := goerd.GenerateMigrationSQL(src, dst)
		if cmdIsPrint {
			for _, q := range qs {
				if !*drop {
					if strings.HasPrefix(strings.ToUpper(q), "DROP") {
						fmt.Println("--", q)
						continue
					}
					if strings.Contains(strings.ToUpper(q), "DROP COLUMN") {
						fmt.Println("--", q)
						continue
					}
				}
				fmt.Println(q)
			}
		} else if cmdIsApply {
			db, err := sql.Open("pgx", *to)
			if err != nil {
				log.Fatal(err)
			}
			tx, err := db.Begin()
			if err != nil {
				log.Fatal(err)
			}
			for _, q := range qs {
				if !*drop {
					if strings.HasPrefix(strings.ToUpper(q), "DROP") {
						fmt.Println("--", q)
						continue
					}
					if strings.Contains(strings.ToUpper(q), "DROP COLUMN") {
						fmt.Println("--", q)
						continue
					}
				}
				fmt.Println(q)
				_, err = tx.Exec(q)
				if err != nil {
					_ = tx.Rollback()
					db.Close()
					log.Fatal(err)
				}
			}
			if err = tx.Commit(); err != nil {
				db.Close()
				log.Fatal(err)
			}
			db.Close()
		} else {
			log.Fatal("wrong command")
		}
	case srcIsPg && dstIsYaml:
		src, err := goerd.SchemaFromPostgresWithConnect(*from)
		if err != nil {
			log.Fatal(err)
		}

		wr, err := os.OpenFile(*to, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatal(err)
		}

		if err := goerd.SchemaToYAML(src, wr); err != nil {
			wr.Close()
			log.Fatal(err)
		}
		wr.Close()

	case srcIsPg && dstIsPuml:
		src, err := goerd.SchemaFromPostgresWithConnect(*from)
		if err != nil {
			log.Fatal(err)
		}

		wr, err := os.OpenFile(*to, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatal(err)
		}

		if err := goerd.SchemaToPlantUML(src, wr, *dist); err != nil {
			wr.Close()
			log.Fatal(err)
		}
		wr.Close()

	case srcIsPg && dstIsPg:
		// here from is destination schema described by database, migrate to it
		dst, err := goerd.SchemaFromPostgresWithConnect(*from)
		if err != nil {
			log.Fatal(err)
		}
		src, err := goerd.SchemaFromPostgresWithConnect(*to)
		if err != nil {
			log.Fatal(err)
		}
		qs := goerd.GenerateMigrationSQL(src, dst)
		if cmdIsPrint {
			for _, q := range qs {
				if !*drop {
					if strings.HasPrefix(strings.ToUpper(q), "DROP") {
						fmt.Println("--", q)
						continue
					}
					if strings.Contains(strings.ToUpper(q), "DROP COLUMN") {
						fmt.Println("--", q)
						continue
					}
				}
				fmt.Println(q)
			}
		} else if cmdIsApply {
			db, err := sql.Open("pgx", *to)
			if err != nil {
				log.Fatal(err)
			}
			tx, err := db.Begin()
			if err != nil {
				log.Fatal(err)
			}
			for _, q := range qs {
				if !*drop {
					if strings.HasPrefix(strings.ToUpper(q), "DROP") {
						fmt.Println("--", q)
						continue
					}
					if strings.Contains(strings.ToUpper(q), "DROP COLUMN") {
						fmt.Println("--", q)
						continue
					}
				}
				fmt.Println(q)
				_, err = tx.Exec(q)
				if err != nil {
					_ = tx.Rollback()
					db.Close()
					log.Fatal(err)
				}
			}
			if err = tx.Commit(); err != nil {
				db.Close()
				log.Fatal(err)
			}
			db.Close()
		} else {
			log.Fatal("wrong command")
		}
	default:
		flag.Usage()
	}
}
