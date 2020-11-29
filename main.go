package main

import (
	"flag"
	"log"
	"os"

	"github.com/covrom/goerd/datasource"
)

var (
	dsn  = flag.String("dsn", "", "postgresql database DSN")
	yml  = flag.String("oy", "schema.yaml", "output yaml filename")
	pml  = flag.String("op", "schema.puml", "output plant uml filename")
	dist = flag.Int("opdist", 2, "distance for plant uml")
)

func main() {
	flag.Parse()
	if *dsn == "" || (*yml == "" && *pml == "") {
		flag.Usage()
		return
	}

	s, err := datasource.Analyze(*dsn)
	if err != nil {
		log.Fatal(err)
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
