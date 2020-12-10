package main

import (
	"flag"
	"log"
	"os"

	"github.com/covrom/goerd/datasource"
	"github.com/covrom/goerd/schema"
)

var (
	dsn   = flag.String("dsn", "", "Build a DSN e.g. postgres://username:password@url:port/dbName")
	inyml = flag.String("iy", "", "input yaml filename")
	yml   = flag.String("oy", "schema.yaml", "output yaml filename")
	pml   = flag.String("op", "schema.puml", "output plant uml filename")
	dist  = flag.Int("opdist", 2, "distance for plant uml")
)

func main() {
	flag.Parse()
	if (*dsn == "" && *inyml == "") || (*yml == "" && *pml == "") {
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
}
