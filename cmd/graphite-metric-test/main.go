// graphite-metric-test is a verifier for Graphite Plaintext Protocol format.
//
// Usage
//
//	graphite-metric-test [-f rule] [file ...]
//
// graphite-metric-test reads the rules, reads metrics from stdin by default, and verify them.
// Then it ends up reports missing metrics, unexpected metrics or out of range metrics.
//
// Options
//
// The -f option is a file contains rules with metric path patterns and metric value ranges.
//
// The Rules
//
// The rule described in the rule file each lines is a pair of metric path pattern and value range.
//
//	// comment
//	local.random.diceroll	>0, <=6	 // v > 0 && v <= 6
//	local.thermal.*.temp	<=100000 // wildcard (* or #) matches any stem in the path
//	~local.network.tx.bytes	>0 // path starting with ~ is optional
//	local.uptime // no range; it checks path existence but the value is not checked
//
// If you want to check metrics with OR condition, you can put multiple lines with same path pattern.
//
//	local.signal.level		>=0, <2
//	local.signal.level		>=3, <5
//
// The Operators
//
// The operators are '<=', '<', '>=' and '>'.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/lufia/graphitemetrictest"
)

var (
	flagFile = flag.String("f", "metricrules", "a pattern `file` for metrics")

	argv0   = filepath.Base(os.Args[0])
	nerrors int
)

func logf(format string, args ...interface{}) {
	log.Printf(format, args...)
	nerrors++
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "usage: %s [options] [file ...]\n", argv0)
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(fmt.Sprintf("%s: ", argv0))
	flag.Usage = usage
	flag.Parse()

	f, err := os.Open(*flagFile)
	if err != nil {
		log.Fatalf("cannot open %s: %v", *flagFile, err)
	}
	rules, err := graphitemetrictest.ReadRules(f)
	if err != nil {
		log.Fatalf("cannot parse %s: %v", *flagFile, err)
	}
	f.Close()

	if flag.NArg() == 0 {
		log.SetPrefix(fmt.Sprintf("%s: %s: ", argv0, "<stdin>"))
		checkMetrics(rules, os.Stdin)
	} else {
		for _, file := range flag.Args() {
			f, err = os.Open(file)
			if err != nil {
				logf("cannot open %s: %v", file, err)
				continue
			}
			log.SetPrefix(fmt.Sprintf("%s: %s: ", argv0, file))
			checkMetrics(rules, f)
			f.Close()
		}
	}
	if nerrors > 0 {
		os.Exit(1)
	}
}

func checkMetrics(rules []*graphitemetrictest.Rule, r io.Reader) {
	metrics, err := graphitemetrictest.ReadMetrics(r)
	if err != nil {
		logf("cannot parse metrics: %v", err)
		return
	}

	diffs := graphitemetrictest.Diff(rules, metrics)
	for _, d := range diffs {
		if d.Rule != nil && d.Metric != nil {
			logf("metric %v is violated to rule %v\n", d.Metric, d.Rule)
		} else if d.Rule == nil {
			logf("found unexpected metric %v\n", d.Metric)
		} else {
			logf("rule %v is not matched any metrics\n", d.Rule)
		}
	}
}
