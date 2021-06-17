// graphite-metric-test is a verifier for Graphite Plaintext Protocol format.
//
// Usage:
//
//	graphite-metric-test [-f rule] [file ...]
//
// The rule option is a file contains rules with metric path patterns and metric value ranges.
// graphite-metric-test reads the rules, reads metrics (from stdin by default),
// verify them and reports mismatches.
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
// Finally, the operators are '<=', '<', '>=' and '>'.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
	rules, err := graphitemetrictest.Parse(f)
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
	var metrics []*graphitemetrictest.Metric

	f := bufio.NewScanner(r)
	for f.Scan() {
		s := strings.TrimSpace(f.Text())
		if s == "" {
			continue
		}
		a := strings.Fields(s)
		if len(a) != 3 {
			logf("%s: a metric must be constructed three fields\n", s)
			continue
		}
		value, err := strconv.ParseFloat(a[1], 64)
		if err != nil {
			logf("%s: %v\n", s, err)
			continue
		}
		tick, err := strconv.ParseInt(a[2], 10, 64)
		if err != nil {
			logf("%s: %v\n", s, err)
			continue
		}
		metrics = append(metrics, &graphitemetrictest.Metric{
			Path:      a[0],
			Value:     value,
			Timestamp: tick,
		})
	}
	if err := f.Err(); err != nil {
		logf("an error occurs while reading: %v\n", err)
		return
	}

	diffs := graphitemetrictest.Diff(rules, metrics)
	for _, d := range diffs {
		if d.Rule != nil && d.Metric != nil {
			logf("metric %v is violated for rule %v\n", d.Metric, d.Rule)
		} else if d.Rule == nil {
			logf("found unexpected metric %v\n", d.Metric)
		} else {
			logf("rule %v is not matched any metrics\n", d.Rule)
		}
	}
}
