package main

import (
	"fmt"
	"strconv"
)

func main() {
	x, err := strconv.Atoi("")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(x)
}
