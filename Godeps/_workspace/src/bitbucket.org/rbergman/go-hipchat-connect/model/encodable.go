package model

import (
	"bufio"
	"bytes"
	"io"
)

// Encodable is an interface for types that implement the ability to JSON-encode
// their contents.
type Encodable interface {
	Encode(w io.Writer) error
}

// EncodeToBytes encodes an Encodable as a JSON []byte.
func EncodeToBytes(encodable Encodable) ([]byte, error) {
	buf := bytes.Buffer{}
	w := bufio.NewWriter(&buf)
	err := encodable.Encode(w)
	if err != nil {
		return []byte{}, err
	}

	err = w.Flush()
	if err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
}

// NewReader creates an io.Reader from an Encodable's encoded form.
func NewReader(encodable Encodable) (io.Reader, error) {

	b, err := EncodeToBytes(encodable)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}
