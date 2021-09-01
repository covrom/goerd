package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/covrom/goerd/datasource"
	"github.com/covrom/goerd/schema"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var (
	dsn     = flag.String("dsn", "", "Build a DSN e.g. postgres://username:password@url:port/dbName")
	inyml   = flag.String("iy", "", "input yaml filename")
	yml     = flag.String("oy", "schema.yaml", "output yaml filename")
	pml     = flag.String("op", "schema.puml", "output plant uml filename")
	dist    = flag.Int("opdist", 2, "distance for plant uml")
	fromyml = flag.String("from", "", "source schema yaml filename")
	toyml   = flag.String("to", "", "destination schema yaml filename")
)

func main() {
	flag.Parse()
	if (*dsn == "" && *inyml == "" && *fromyml == "" && *toyml == "") ||
		(*yml == "" && *pml == "" && *fromyml == "" && *toyml == "") {
		flag.Usage()
		return
	}

	s := &schema.Schema{}

	if *dsn != "" {
		var err error
		s, err = datasource.Analyze(*dsn)
		if err != nil {
			log.Fatal(err)
		}
	} else if *inyml != "" {
		f, err := os.Open(*inyml)
		if err != nil {
			log.Fatal(err)
		}
		if err := s.LoadYaml(f); err != nil {
			f.Close()
			log.Fatal(err)
		}
		f.Close()
	}

	if *yml != "" {
		wr, err := os.OpenFile(*yml, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatal(err)
		}

		if err := s.SaveYaml(wr); err != nil {
			wr.Close()
			log.Fatal(err)
		}
		wr.Close()
	}

	if *pml != "" {
		wr, err := os.OpenFile(*pml, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatal(err)
		}

		if err := s.SavePlantUml(wr, *dist); err != nil {
			wr.Close()
			log.Fatal(err)
		}
		wr.Close()
	}

	if *fromyml != "" || *toyml != "" {
		sfrom := &schema.Schema{}
		if *fromyml != "" {
			ffrom, err := os.Open(*fromyml)
			if err != nil {
				log.Fatal(err)
			}
			if err := sfrom.LoadYaml(ffrom); err != nil {
				ffrom.Close()
				log.Fatal(err)
			}
			ffrom.Close()
		}
		sto := &schema.Schema{}
		if *toyml != "" {
			fto, err := os.Open(*toyml)
			if err != nil {
				log.Fatal(err)
			}
			if err := sto.LoadYaml(fto); err != nil {
				fto.Close()
				log.Fatal(err)
			}
			fto.Close()
		}

		ptch := &schema.PatchSchema{CurrentSchema: sfrom.CurrentSchema}
		ptch.Build(sfrom, sto)
		qs := ptch.GenerateSQL()
		for _, q := range qs {
			fmt.Println(q)
		}
	}
}
