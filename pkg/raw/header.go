package raw

import (
	"encoding/binary"
)

type Header []byte

const HeaderSize = 8

func (h Header) Encode(len uint32, id uint32) {
	binary.BigEndian.PutUint32(h[:4], len)
	binary.BigEndian.PutUint32(h[4:8], id)
}

func (h Header) Next() uint32 {
	return binary.BigEndian.Uint32(h[:4])
}

func (h Header) ID() uint32 {
	return binary.BigEndian.Uint32(h[4:8])
}
