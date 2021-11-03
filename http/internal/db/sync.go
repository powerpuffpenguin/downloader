package db

import (
	"encoding/gob"
	"encoding/json"
	"io"
)

type Decoder interface {
	Decode(e interface{}) error
}
type Encoder interface {
	Encode(e interface{}) error
}

func NewDecoder(r io.Reader, j bool) Decoder {
	if j {
		return json.NewDecoder(r)
	}
	return gob.NewDecoder(r)
}
func NewEncoder(w io.Writer, j bool) Encoder {
	if j {
		return json.NewEncoder(w)
	}
	return gob.NewEncoder(w)
}
