package raw

import (
	"encoding/binary"
)

type header []byte

const HeaderSize = 6

func (h header) encode(len uint32, id uint16) {
	binary.BigEndian.PutUint32(h[:4], len)
	binary.BigEndian.PutUint16(h[4:6], id)
}

func (h header) Next() uint32 {
	return binary.BigEndian.Uint32(h[:4])
}

func (h header) ID() uint16 {
	return binary.BigEndian.Uint16(h[4:6])
}
