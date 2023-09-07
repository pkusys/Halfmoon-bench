package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/eniac/Beldi/internal/utils"
	"github.com/eniac/Beldi/pkg/beldilib"
)

const table = "rw"

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
	if beldilib.TYPE == "BASELINE" {
		beldilib.DeleteTable(fmt.Sprintf("b%s", table))
		beldilib.WaitUntilDeleted(fmt.Sprintf("b%s", table))
	} else {
		beldilib.DeleteLambdaTables(fmt.Sprintf("%s", table))
		beldilib.WaitUntilDeleted(fmt.Sprintf("%s", table))
		beldilib.WaitUntilDeleted(fmt.Sprintf("%s-log", table))
		beldilib.WaitUntilDeleted(fmt.Sprintf("%s-collector", table))
	}
}

func create() {
	if beldilib.TYPE == "BASELINE" {
		beldilib.CreateBaselineTable(fmt.Sprintf("b%s", table))
		beldilib.WaitUntilActive(fmt.Sprintf("b%s", table))
	} else {
		beldilib.CreateLambdaTables(fmt.Sprintf("%s", table))
		time.Sleep(10 * time.Second)
		beldilib.WaitUntilActive(fmt.Sprintf("%s", table))
		beldilib.WaitUntilActive(fmt.Sprintf("%s-log", table))
		beldilib.WaitUntilActive(fmt.Sprintf("%s-collector", table))
	}
}

func populate() {
	baseline := beldilib.TYPE == "BASELINE"
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
