package main

import (
	"os"
	"strconv"

	"github.com/eniac/Beldi/pkg/cayonlib"
)

const table = "singleop"

var nKeys = 10000
var value = 1

func init() {
	if nk, err := strconv.Atoi(os.Getenv("NUM_KEYS")); err == nil {
		nKeys = nk
	} else {
		panic("invalid NUM_KEYS")
	}
}

func clean() {
	cayonlib.DeleteLambdaTables(table)
	cayonlib.WaitUntilDeleted(table)
}

func create() {
	cayonlib.CreateLambdaTables(table)
	cayonlib.WaitUntilActive(table)
}

func populate() {
	for i := 0; i < nKeys; i++ {
		cayonlib.Populate(table, strconv.Itoa(i), value, false)
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
