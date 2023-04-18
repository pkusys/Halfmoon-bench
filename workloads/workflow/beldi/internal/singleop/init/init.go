package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/eniac/Beldi/pkg/beldilib"
)

var table = "singleop"
var baseline = false

var nKeys = 10000
var value = 1

func init() {
	if nk, err := strconv.Atoi(os.Getenv("NUM_KEYS")); err == nil {
		nKeys = nk
	} else {
		panic("invalid NUM_KEYS")
	}
	if beldilib.TYPE == "BASELINE" {
		table = "b" + table
		baseline = true
	}
}

func clean() {
	if baseline {
		beldilib.DeleteTable(table)
		beldilib.WaitUntilDeleted(table)
	} else {
		beldilib.DeleteLambdaTables(table)
		beldilib.WaitUntilDeleted(table)
		beldilib.WaitUntilDeleted(fmt.Sprintf("%s-log", table))
		beldilib.WaitUntilDeleted(fmt.Sprintf("%s-collector", table))
	}
}

func create() {
	if baseline {
		beldilib.CreateBaselineTable(table)
		time.Sleep(10 * time.Second)
		beldilib.WaitUntilActive(table)
	} else {
		for _, lambda := range []string{table, "nop"} {
			beldilib.CreateLambdaTables(lambda)
			time.Sleep(10 * time.Second)
			beldilib.WaitUntilActive(lambda)
			beldilib.WaitUntilActive(fmt.Sprintf("%s-log", lambda))
			beldilib.WaitUntilActive(fmt.Sprintf("%s-collector", lambda))
		}
	}
}

func populate() {
	for i := 0; i < nKeys; i++ {
		beldilib.Populate(table, strconv.Itoa(i), value, baseline)
	}
}

func main() {
	option := os.Args[1]
	if option == "clean" {
		clean()
	} else if option == "create" {
		create()
	} else if option == "populate" {
		populate()
	} else {
		panic("unkown option: " + option)
	}
}
