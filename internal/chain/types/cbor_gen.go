package types

import (
	"fmt"
	"io"

	"github.com/filecoin-project/go-state-types/abi"
	cbg "github.com/whyrusleeping/cbor-gen"
	"golang.org/x/xerrors"
)

var lengthBufMessage = []byte{138}

func (t *Message) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}

	cw := cbg.NewCborWriter(w)

	if _, err := cw.Write(lengthBufMessage); err != nil {
		return err
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajUnsignedInt, uint64(t.Version)); err != nil {
		return err
	}

	if err := t.To.MarshalCBOR(cw); err != nil {
		return err
	}

	if err := t.From.MarshalCBOR(cw); err != nil {
		return err
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajUnsignedInt, uint64(t.Nonce)); err != nil {
		return err
	}

	if err := t.Value.MarshalCBOR(cw); err != nil {
		return err
	}

	if t.GasLimit >= 0 {
		if err := cw.WriteMajorTypeHeader(cbg.MajUnsignedInt, uint64(t.GasLimit)); err != nil {
			return err
		}
	} else {
		if err := cw.WriteMajorTypeHeader(cbg.MajNegativeInt, uint64(-t.GasLimit-1)); err != nil {
			return err
		}
	}

	if err := t.GasFeeCap.MarshalCBOR(cw); err != nil {
		return err
	}

	if err := t.GasPremium.MarshalCBOR(cw); err != nil {
		return err
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajUnsignedInt, uint64(t.Method)); err != nil {
		return err
	}

	if len(t.Params) > 2097152 {
		return xerrors.Errorf("Byte array in field t.Params was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajByteString, uint64(len(t.Params))); err != nil {
		return err
	}

	if _, err := cw.Write(t.Params); err != nil {
		return err
	}

	return nil
}

func (t *Message) UnmarshalCBOR(r io.Reader) (err error) {
	*t = Message{}

	cr := cbg.NewCborReader(r)

	maj, extra, err := cr.ReadHeader()
	if err != nil {
		return err
	}
	defer func() {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
	}()

	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 10 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	{
		maj, extra, err = cr.ReadHeader()
		if err != nil {
			return err
		}
		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("wrong type for uint64 field")
		}
		t.Version = uint64(extra)
	}

	{
		if err := t.To.UnmarshalCBOR(cr); err != nil {
			return xerrors.Errorf("unmarshaling t.To: %w", err)
		}
	}

	{
		if err := t.From.UnmarshalCBOR(cr); err != nil {
			return xerrors.Errorf("unmarshaling t.From: %w", err)
		}
	}

	{
		maj, extra, err = cr.ReadHeader()
		if err != nil {
			return err
		}
		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("wrong type for uint64 field")
		}
		t.Nonce = uint64(extra)
	}

	{
		if err := t.Value.UnmarshalCBOR(cr); err != nil {
			return xerrors.Errorf("unmarshaling t.Value: %w", err)
		}
	}

	{
		maj, extra, err := cr.ReadHeader()
		if err != nil {
			return err
		}
		var extraI int64
		switch maj {
		case cbg.MajUnsignedInt:
			extraI = int64(extra)
			if extraI < 0 {
				return fmt.Errorf("int64 positive overflow")
			}
		case cbg.MajNegativeInt:
			extraI = int64(extra)
			if extraI < 0 {
				return fmt.Errorf("int64 negative overflow")
			}
			extraI = -1 - extraI
		default:
			return fmt.Errorf("wrong type for int64 field: %d", maj)
		}

		t.GasLimit = int64(extraI)
	}

	{
		if err := t.GasFeeCap.UnmarshalCBOR(cr); err != nil {
			return xerrors.Errorf("unmarshaling t.GasFeeCap: %w", err)
		}
	}

	{
		if err := t.GasPremium.UnmarshalCBOR(cr); err != nil {
			return xerrors.Errorf("unmarshaling t.GasPremium: %w", err)
		}
	}

	{
		maj, extra, err = cr.ReadHeader()
		if err != nil {
			return err
		}
		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("wrong type for uint64 field")
		}
		t.Method = abi.MethodNum(extra)
	}

	maj, extra, err = cr.ReadHeader()
	if err != nil {
		return err
	}

	if extra > 2097152 {
		return fmt.Errorf("t.Params: byte array too large (%d)", extra)
	}
	if maj != cbg.MajByteString {
		return fmt.Errorf("expected byte array")
	}

	if extra > 0 {
		t.Params = make([]uint8, extra)
	}

	if _, err := io.ReadFull(cr, t.Params); err != nil {
		return err
	}

	return nil
}

var lengthBufSignedMessage = []byte{130}

func (t *SignedMessage) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}

	cw := cbg.NewCborWriter(w)

	if _, err := cw.Write(lengthBufSignedMessage); err != nil {
		return err
	}

	if err := t.Message.MarshalCBOR(cw); err != nil {
		return err
	}

	if err := t.Signature.MarshalCBOR(cw); err != nil {
		return err
	}
	return nil
}

func (t *SignedMessage) UnmarshalCBOR(r io.Reader) (err error) {
	*t = SignedMessage{}

	cr := cbg.NewCborReader(r)

	maj, extra, err := cr.ReadHeader()
	if err != nil {
		return err
	}
	defer func() {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
	}()

	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 2 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	{
		if err := t.Message.UnmarshalCBOR(cr); err != nil {
			return xerrors.Errorf("unmarshaling t.Message: %w", err)
		}
	}

	{
		if err := t.Signature.UnmarshalCBOR(cr); err != nil {
			return xerrors.Errorf("unmarshaling t.Signature: %w", err)
		}
	}
	return nil
}
