package graphitemetrictest_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/lufia/graphitemetrictest"
)

const rules = `
custom.disks.#.reads.bytes	>=0
custom.disks.#.writes.bytes	>=0
~custom.interfaces.#.tx.packets	>=0
~custom.interfaces.#.rx.packets	>=0
`

func Example() {
	r := strings.NewReader(rules)
	rules, err := graphitemetrictest.ReadRules(r)
	if err != nil {
		log.Fatal(err)
	}

	metrics := []*graphitemetrictest.Metric{
		{
			Path:      "custom.disks.sdC0.reads.bytes",
			Value:     3012.0,
			Timestamp: 1623990692,
		},
		{
			Path:      "custom.disks.sdC0.writes.bytes",
			Value:     24091.0,
			Timestamp: 1623990692,
		},
		{
			Path:      "custom.disks.sdD0.reads.bytes",
			Value:     24520.0,
			Timestamp: 1623990692,
		},
		{
			Path:      "custom.interfaces.ether0.rx.packets",
			Value:     30802.0,
			Timestamp: 1623990692,
		},
	}
	diffs := graphitemetrictest.Diff(rules, metrics)
	fmt.Println(diffs)
	// Output: []
}
