package raw

import (
	"log"
	"net"
)

type Downstream struct {
	downaddr string
	conn     net.Conn
	sendCh   chan Message
	stopCh   chan struct{}
	ID       uint32
}

func NewDownstream(downaddr string, sendCh chan Message, id uint32) *Downstream {
	d := Downstream{
		downaddr: downaddr,
		sendCh:   sendCh,
		stopCh:   make(chan struct{}),
		ID:       id,
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

	d.conn = conn
}

func (d *Downstream) recv() {
	for {
		select {
		case _ = <-d.stopCh:
			d.conn.Close()
			return
		default:
			h := Header(make([]byte, HeaderSize))
			b := make([]byte, 1096)
			n, _ := d.conn.Read(b)
			if n > 0 {
				h.Encode(TICK, uint32(n), d.ID)
				msg := Message{
					Header:  h,
					Payload: b[:n],
				}
				d.sendCh <- msg
			}
		}
	}
}
