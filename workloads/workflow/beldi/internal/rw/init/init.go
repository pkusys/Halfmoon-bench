package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/eniac/Beldi/internal/rw/utils"
	"github.com/eniac/Beldi/pkg/beldilib"
)

const table = "rw"

var nKeys = 10000
var valueSize = 256 // bytes
var value []byte

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
	beldilib.DeleteTable(fmt.Sprintf("b%s", table))
	beldilib.WaitUntilDeleted(fmt.Sprintf("b%s", table))
}

func create() {
	beldilib.CreateBaselineTable(fmt.Sprintf("b%s", table))
	beldilib.WaitUntilActive(fmt.Sprintf("b%s", table))
}

func populate() {
	for i := 0; i < nKeys; i++ {
		beldilib.Populate(table, strconv.Itoa(i), value, true)
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
