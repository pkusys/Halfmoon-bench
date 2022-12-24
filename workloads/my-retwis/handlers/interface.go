package handlers

import (
	"bytes"
	"encoding/gob"
	"io"

	"github.com/golang/snappy"
)

const compress = false

func init() {
	gob.Register(WriteLog{})
}

func CompressData(uncompressed []byte) []byte {
	return snappy.Encode(nil, uncompressed)
}

func DecompressReader(compressed []byte) (io.Reader, error) {
	uncompressed, err := snappy.Decode(nil, compressed)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(uncompressed), nil
}

func Decode(raw []byte) (interface{}, error) {
	var reader io.Reader
	if compress {
		r, err := DecompressReader(raw)
		if err != nil {
			return nil, err
		}
		reader = r
	} else {
		reader = bytes.NewReader(raw)
	}
	dec := gob.NewDecoder(reader)
	var value interface{}
	if err := dec.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

func Encode(value interface{}) ([]byte, error) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(&value); err != nil {
		return nil, err
	}
	if compress {
		return CompressData(buf.Bytes()), nil
	}
	return buf.Bytes(), nil
}

type Updatable interface {
	Update(params WriteOp) (interface{}, error)
}

type WriteOp map[string]interface{}

type WriteLog map[uint64]WriteOp
