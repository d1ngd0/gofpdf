package geh

import (
	"bytes"
	"encoding/gob"
)

func EncodeMany(v ...interface{}) ([]byte, error) {
	var err error
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)

	for x := 0; x < len(v); x++ {
		if err == nil {
			err = encoder.Encode(v[x])
		}
	}

	return w.Bytes(), err
}

// GobDecode decodes the specified byte buffer into the receiving template.
func DecodeMany(buf []byte, v ...interface{}) error {
	r := bytes.NewBuffer(buf)

	decoder := gob.NewDecoder(r)
	for x := 0; x < len(v); x++ {
		if err := decoder.Decode(v[x]); err != nil {
			return err
		}
	}

	return nil
}
