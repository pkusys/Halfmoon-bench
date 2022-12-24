package main

import "fmt"

func Append(b []byte, data byte) {
	b = append(b, data)
}

func Copy(b []byte, data byte) {
	copy(b, []byte{data})
}

func main() {
	b := make([]byte, 0, 1)
	Append(b, 1)
	fmt.Println(b, len(b), cap(b))
	Copy(b, 1)
	fmt.Println(b, len(b), cap(b))
}
