package main

import (
	"flag"
	"log"
	"os"

	"github.com/covrom/goerd/datasource"
)

var (
	dsn = flag.String("dsn", "", "postgresql database DSN")
	yml = flag.String("o", "schema.yaml", "output yaml filename")
)

func main() {
	flag.Parse()
	if *dsn == "" || *yml == "" {
		flag.Usage()
		return
	}
	wr, err := os.OpenFile(*yml, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	s, err := datasource.Analyze(*dsn)
	if err != nil {
		wr.Close()
		log.Fatal(err)
	}
	if err := s.SaveYaml(wr); err != nil {
		wr.Close()
		log.Fatal(err)
	}
	wr.Close()
}
