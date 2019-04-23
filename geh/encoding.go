package geh

import (
	"bytes"
	"encoding/gob"

	"github.com/d1ngd0/gofpdf/bp"
)

func EncodeMany(v ...interface{}) ([]byte, error) {
	var err error
	w := bp.GetBuffer()
	encoder := gob.NewEncoder(w)

	for x := 0; x < len(v); x++ {
		if err == nil {
			err = encoder.Encode(v[x])
		}
	}

	return w.Bytes(), err
}

// GobDecode decodes the specified byte buffer into the receiving template.
func DecodeMany(b []byte, v ...interface{}) error {
	decoder := gob.NewDecoder(bytes.NewReader(b))
	for x := 0; x < len(v); x++ {
		if err := decoder.Decode(v[x]); err != nil {
			return err
		}
	}

	return nil
}
