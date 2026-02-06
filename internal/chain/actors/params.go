package actors

import (
	"bytes"
	"fmt"
	"io"
)

type cborMarshaler interface {
	MarshalCBOR(io.Writer) error
}

func SerializeParams(p interface{}) ([]byte, error) {
	if p == nil {
		return nil, nil
	}

	m, ok := p.(cborMarshaler)
	if !ok {
		return nil, fmt.Errorf("params type %T does not support cbor", p)
	}

	var buf bytes.Buffer
	if err := m.MarshalCBOR(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
