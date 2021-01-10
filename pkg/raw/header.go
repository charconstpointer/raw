package raw

import (
	"encoding/binary"
)

type Header []byte

const HeaderSize = 10

const ProtoVersion = 1

func (h Header) Encode(len uint32, id uint32) {
	binary.BigEndian.PutUint16(h[:2], ProtoVersion)
	binary.BigEndian.PutUint32(h[2:6], len)
	binary.BigEndian.PutUint32(h[6:], id)
}

func (h Header) ProtoVersion() uint16 {
	return binary.BigEndian.Uint16(h[:2])
}

func (h Header) Next() uint32 {
	return binary.BigEndian.Uint32(h[2:6])
}

func (h Header) ID() uint32 {
	return binary.BigEndian.Uint32(h[6:])
}
