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

// Bytes encodes an Encodable as a JSON []byte.
func Bytes(encodable Encodable) ([]byte, error) {
	buf := bytes.Buffer{}
	w := bufio.NewWriter(&buf)
	if err := encodable.Encode(w); err != nil {
		return []byte{}, err
	}
	if err := w.Flush(); err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}

// String converts an Encodable to a JSON string.
func String(encodable Encodable) (string, error) {
	encd := bytes.Buffer{}
	if err := encodable.Encode(&encd); err != nil {
		return "", err
	}
	return encd.String(), nil
}

// MustString converts an Encodable to a JSON string, panicking on failure.
func MustString(encodable Encodable) string {
	s, err := String(encodable)
	if err != nil {
		panic(err)
	}
	return s
}

// NewReader creates an io.Reader from an Encodable's encoded form.
func NewReader(encodable Encodable) (io.Reader, error) {
	b, err := Bytes(encodable)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}
