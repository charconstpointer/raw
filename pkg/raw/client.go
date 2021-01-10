package raw

import (
	"bytes"
	"io"

	log "github.com/sirupsen/logrus"

	"net"

	"github.com/pkg/errors"
)

type Client struct {
	downstreams map[int32]*Downstream
	downstream  net.Conn
	upstream    net.Conn

	downaddr string
	upaddr   string

	sendCh chan Message
}

func NewClient(upaddr string, downaddr string) (*Client, error) {
	upstream, err := net.Dial("tcp", upaddr)
	if err != nil {
		return nil, errors.Wrap(err, "could not connect to upstream")
	}
	log.Infof("🧵connected to upstream at %s", upaddr)
	c := &Client{
		downstreams: make(map[int32]*Downstream),
		sendCh:      make(chan Message),

		upstream: upstream,

		upaddr:   upaddr,
		downaddr: downaddr,
	}

	return c, nil
}
func (c *Client) Run() {
	go c.recv()
	c.send()
}

func (c *Client) recv() {
	for {
		select {
		case msg := <-c.sendCh:
			sent := 0
			for sent < HeaderSize {
				n, _ := c.upstream.Write(msg.Header)
				sent += n
			}

			io.Copy(c.upstream, bytes.NewBuffer(msg.Payload))
		}
	}
}

func (c *Client) send() {
	for {
		h := Header(make([]byte, HeaderSize))
		io.ReadFull(c.upstream, h)
		if h.MessageType() == TERM {
			d := c.downstreams[int32(h.ID())]
			log.Printf("❌terminating downstream connection on port %s", d.conn.LocalAddr().String())
			d.stopCh <- struct{}{}
			continue
		}
		if c.downstreams[int32(h.ID())] == nil {
			log.Println("new downstream conn")
			d := NewDownstream(c.downaddr, c.sendCh, h.ID())
			d.Run()
			c.downstreams[int32(h.ID())] = d
		}

		if h != nil {
			mb := make([]byte, int(h.Next()))
			n, _ := io.ReadFull(c.upstream, mb)
			if n > 0 {
				d := c.downstreams[int32(h.ID())]
				if d == nil {
					continue
				}
				io.Copy(d.conn, bytes.NewBuffer(mb))
			}
		}

	}
}
