// graphdef2rule generates rules from GraphDef JSON.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type GraphDefMetric struct {
	Name string `json:"name"`
}

type GraphDef struct {
	Unit    string            `json:"unit"`
	Metrics []*GraphDefMetric `json:"metrics"`
}

type GraphDefs struct {
	Graphs map[string]*GraphDef `json:"graphs"`
}

var argv0 = filepath.Base(os.Args[0])

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "usage: %s [file ...]\n", argv0)
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(fmt.Sprintf("%s: ", argv0))
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 0 {
		GenerateRule(os.Stdin)
	} else {
		for _, file := range flag.Args() {
			f, err := os.Open(file)
			if err != nil {
				log.Fatalf("cannot open %s: %v\n", file, err)
			}
			log.SetPrefix(fmt.Sprintf("%s: %s: ", argv0, file))
			GenerateRule(f)
			f.Close()
		}
	}
}

func GenerateRule(r io.Reader) {
	var graphDefs GraphDefs

	if err := json.NewDecoder(r).Decode(&graphDefs); err != nil {
		log.Fatalf("cannot decode a JSON: %v", err)
	}
	for key, g := range graphDefs.Graphs {
		for _, m := range g.Metrics {
			exprs := ">=0"
			if g.Unit == "percentage" {
				exprs = ">=0, <=100"
			}
			fmt.Printf("%s.%s\t%s\n", key, m.Name, exprs)
		}
	}
}
