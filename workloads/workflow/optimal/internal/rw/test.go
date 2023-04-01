package main

import (
	"encoding/json"
	"fmt"

	"github.com/eniac/Beldi/internal/rw/utils"
)

func main() {
	valueSize := 256
	value := utils.RandomString(valueSize)
	fmt.Println(len(value))
	jsonData := map[string]interface{}{
		"V": value,
	}
	serialized, err := json.Marshal(jsonData)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(serialized))
}
