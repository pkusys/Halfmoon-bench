package main

import (
	"encoding/json"
	"fmt"

	"github.com/eniac/Beldi/internal/utils"
	"github.com/golang/snappy"
)

func main() {
	valueSize := 256
	value := utils.RandomString(valueSize)
	fmt.Println(len(value))
	jsonData := map[string]interface{}{
		"V": string(value),
	}
	serialized, err := json.Marshal(jsonData)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(serialized))
	encoded := snappy.Encode(nil, serialized)
	fmt.Println(len(encoded))
	deserialized := map[string]interface{}{}
	err = json.Unmarshal(serialized, &deserialized)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%T, %v\n", deserialized["V"], deserialized["V"] == string(value))
	fmt.Println(len(deserialized["V"].(string)))
}
