package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
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
	beldilib.DeleteLambdaTables(fmt.Sprintf("%s", table))
	beldilib.WaitUntilDeleted(fmt.Sprintf("%s", table))
	beldilib.WaitUntilDeleted(fmt.Sprintf("%s-log", table))
	beldilib.WaitUntilDeleted(fmt.Sprintf("%s-collector", table))
	// counter
	beldilib.DeleteTable("counter")
}

func create() {
	beldilib.CreateLambdaTables(fmt.Sprintf("%s", table))
	time.Sleep(10 * time.Second)
	beldilib.WaitUntilActive(fmt.Sprintf("%s", table))
	beldilib.WaitUntilActive(fmt.Sprintf("%s-log", table))
	beldilib.WaitUntilActive(fmt.Sprintf("%s-collector", table))
	// counter
	beldilib.CreateCounterTable()
	time.Sleep(3 * time.Second)
	beldilib.WaitUntilActive("counter")
	// beldilib.EnableStream(fmt.Sprintf("%s", table))
}

func populate() {
	beldilib.LibWrite("counter", aws.JSONValue{"K": "counter"}, map[expression.NameBuilder]expression.OperandBuilder{
		expression.Name("V"): expression.Value(1),
	})
	for i := 0; i < nKeys; i++ {
		beldilib.Populate(table, strconv.Itoa(i), value, false)
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
