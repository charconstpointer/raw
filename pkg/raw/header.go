package raw

import (
	"encoding/binary"
)

type Header []byte

const HeaderSize = 12

const ProtoVersion = 1

type MessageType uint16

const (
	INIT MessageType = iota
	TICK
	TERM
)

func (h Header) Encode(mtype MessageType, len uint32, id uint32) {
	binary.BigEndian.PutUint16(h[:2], ProtoVersion)
	binary.BigEndian.PutUint16(h[2:4], uint16(mtype))
	binary.BigEndian.PutUint32(h[4:8], len)
	binary.BigEndian.PutUint32(h[8:], id)
}

func (h Header) ProtoVersion() uint16 {
	return binary.BigEndian.Uint16(h[:2])
}

func (h Header) MessageType() MessageType {
	value := binary.BigEndian.Uint16(h[2:4])
	return MessageType(value)
}

func (h Header) Next() uint32 {
	return binary.BigEndian.Uint32(h[4:8])
}

func (h Header) ID() uint32 {
	return binary.BigEndian.Uint32(h[8:])
}
