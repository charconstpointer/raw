package raw

import (
	"log"
	"net"
)

type Downstream struct {
	downaddr string
	conn     net.Conn
	sendCh   chan Message
	ID       uint32
}

func NewDownstream(downaddr string, sendCh chan Message) *Downstream {
	d := Downstream{
		downaddr: downaddr,
		sendCh:   sendCh,
	}
	return &d
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
	d.ID = getID(conn.RemoteAddr().String())
	d.conn = conn
}

func (d *Downstream) recv() {
	h := Header(make([]byte, HeaderSize))
	b := make([]byte, 1096)
	for {
		n, _ := d.conn.Read(b)
		if n > 0 {
			h.Encode(uint32(n), d.ID)
			msg := Message{
				Header:  h,
				Payload: b[:n],
			}
			d.sendCh <- msg
		}
	}
}
