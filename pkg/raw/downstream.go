package raw

import (
	"log"
	"net"
)

type Downstream struct {
	downaddr string
	conn     net.Conn
	sendCh   chan Message
}

func NewDownstream(downaddr string) *Downstream {
	return nil
}

func (d *Downstream) Run() {
	d.openConn()
	go d.recv()
}

func (d *Downstream) openConn() {
	conn, err := net.Dial("tcp", d.downaddr)
	if err != nil {
		log.Fatal(err.Error())
	}

	d.conn = conn
}

func (d *Downstream) recv() {
	h := Header(make([]byte, HeaderSize))
	b := make([]byte, 1096)
	for {
		n, _ := d.conn.Read(b)
		if n > 0 {
			h.Encode(uint32(n), 1)
			msg := Message{
				Header:  h,
				Payload: b[:n],
			}
			d.sendCh <- msg
		}
	}
}
