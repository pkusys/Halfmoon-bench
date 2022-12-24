package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
)

type WriteOp struct {
	SeqNum uint64
	OpType int
	Value  interface{}
	Param  interface{}
}

func EncodeWriteOps(ops map[uint64]*WriteOp) []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(ops); err != nil {
		log.Fatalf("[WARN] failed to encode write ops: %s", err.Error())
	}
	return buf.Bytes()
}

func DecodeWriteOp(key uint64, data []byte) *WriteOp {
	dec := gob.NewDecoder(bytes.NewReader(data))
	var ops map[uint64]*WriteOp
	if err := dec.Decode(&ops); err != nil {
		log.Fatalf("[WARN] failed to load write op: %s", err.Error())
	}
	if op, ok := ops[key]; ok {
		return op
	}
	return nil
}

func main() {
	ops := map[uint64]*WriteOp{
		1: &WriteOp{
			SeqNum: 1,
			OpType: 1,
		},
	}
	data := EncodeWriteOps(ops)
	op := DecodeWriteOp(1, data)
	log.Printf("op: %+v", op)

	var str string = "hello"
	buf := bytes.NewBufferString(str)
	fmt.Println(buf.String())
	str = str[:1]
	fmt.Println(buf.String())
	fmt.Printf("%08x\n", 123)
	encoded, err := json.Marshal(map[string]interface{}{})
	if err != nil {
		log.Fatalf("failed to encode json: %s\n", err.Error())
	}
	log.Printf("encoded: %s %v", encoded, encoded)
}
