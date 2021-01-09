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
	sendCh      chan Message
	ID          uint32
}

func NewClient(upaddr string, downaddr string) (*Client, error) {
	// downstream, err := net.Dial("tcp", downaddr)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "could not connect to downstream")
	// }
	// log.Infof("🎠connected to downstream at %s", downaddr)

	upstream, err := net.Dial("tcp", upaddr)
	if err != nil {
		return nil, errors.Wrap(err, "could not connect to upstream")
	}
	log.Infof("🧵connected to upstream at %s", upaddr)
	id := uint32(4444)
	c := &Client{
		downstreams: make(map[int32]*Downstream),
		// downstream:  downstream,
		upstream: upstream,
		sendCh:   make(chan Message),
		ID:       id,
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
	// h := Header(make([]byte, HeaderSize))
	// b := make([]byte, 1096)
	// for {
	// 	n, _ := c.downstream.Read(b)
	// 	h.Encode(uint32(n), 1)
	// 	sent := 0
	// 	for sent < HeaderSize {
	// 		n, _ := c.upstream.Write(h)
	// 		sent += n
	// 	}
	// 	if n > 0 {
	// 		io.Copy(c.upstream, bytes.NewBuffer(b[:n]))
	// 	}
	// }
}

func (c *Client) send() {
	h := Header(make([]byte, HeaderSize))
	for {
		io.ReadFull(c.upstream, h)
		if c.downstreams[int32(h.ID())] == nil {
			d := NewDownstream(":25565", c.sendCh)
			c.downstreams[int32(h.ID())] = d
		}

		if h != nil {
			mb := make([]byte, int(h.Next()))
			n, _ := io.ReadFull(c.upstream, mb)
			log.Printf("from upstream, expecting %d, got %d", h.Next(), n)
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
