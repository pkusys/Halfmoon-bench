package main

import (
	"os"
	"strconv"

	"github.com/eniac/Beldi/internal/utils"
	"github.com/eniac/Beldi/pkg/cayonlib"
)

const table = "recovery"

var nKeys = 10000
var valueSize = 256 // bytes
var value string

func init() {
	if nk, err := strconv.Atoi(os.Getenv("NUM_KEYS")); err == nil {
		nKeys = nk
	} else {
		panic("invalid NUM_KEYS")
	}
	if vs, err := strconv.Atoi(os.Getenv("VALUE_SIZE")); err == nil {
		valueSize = vs
	} else {
		panic("invalid VALUE_SIZE")
	}
	value = utils.RandomString(valueSize)
}

func clean() {
	cayonlib.DeleteTable(table)
	cayonlib.WaitUntilDeleted(table)
}

func create() {
	cayonlib.CreateMainTable(table)
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
