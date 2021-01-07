package main

import (
	"log"

	"github.com/charconstpointer/raw/pkg/raw"
)

func main() {
	c, err := raw.NewClient(":7000", ":25565")
	if err != nil {
		log.Fatal(err.Error())
	}
	c.Run()
}

// func recv(conn net.Conn, mc net.Conn) {
// 	h := raw.Header(make([]byte, raw.HeaderSize))
// 	b := make([]byte, 1096)
// 	for {
// 		n, _ := mc.Read(b)
// 		h.Encode(uint32(n), 1)
// 		sent := 0
// 		for sent < raw.HeaderSize {
// 			n, _ := conn.Write(h)
// 			sent += n
// 		}
// 		if n > 0 {
// 			log.Println(n)
// 			io.Copy(conn, bytes.NewBuffer(b[:n]))
// 		}
// 	}
// }

// func send(conn net.Conn, mc net.Conn) {
// 	// b := make([]byte, 1096)
// 	h := raw.Header(make([]byte, raw.HeaderSize))
// 	for {
// 		io.ReadFull(conn, h)
// 		if h != nil {
// 			mb := make([]byte, int(h.Next()))
// 			n, _ := io.ReadFull(conn, mb)
// 			// n, _ := conn.Read(b)
// 			if n > 0 {
// 				log.Println(n)
// 				io.Copy(mc, bytes.NewBuffer(mb))
// 			}
// 		}

// 	}
// }
