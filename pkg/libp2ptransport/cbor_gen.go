package libp2ptransport

import (
	"fmt"
	"io"
	"math"
	"sort"

	"github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
	"golang.org/x/xerrors"
)

var _ = xerrors.Errorf
var _ = cid.Undef
var _ = math.E
var _ = sort.Sort

var lengthBufTransportMessage = []byte{129}

func (t *TransportMessage) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}

	cw := cbg.NewCborWriter(w)

	if _, err := cw.Write(lengthBufTransportMessage); err != nil {
		return err
	}

	// t.Payload ([]uint8) (slice)
	if len(t.Payload) > cbg.ByteArrayMaxLen {
		return xerrors.Errorf("Byte array in field t.Payload was too long")
	}

	if err := cw.WriteMajorTypeHeader(cbg.MajByteString, uint64(len(t.Payload))); err != nil {
		return err
	}

	if _, err := cw.Write(t.Payload[:]); err != nil {
		return err
	}
	return nil
}

func (t *TransportMessage) UnmarshalCBOR(r io.Reader) (err error) {
	*t = TransportMessage{}

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

	if extra != 1 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.Payload ([]uint8) (slice)

	maj, extra, err = cr.ReadHeader()
	if err != nil {
		return err
	}

	if extra > cbg.ByteArrayMaxLen {
		return fmt.Errorf("t.Payload: byte array too large (%d)", extra)
	}
	if maj != cbg.MajByteString {
		return fmt.Errorf("expected byte array")
	}

	if extra > 0 {
		t.Payload = make([]uint8, extra)
	}

	if _, err := io.ReadFull(cr, t.Payload[:]); err != nil {
		return err
	}
	return nil
}