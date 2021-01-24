package main

import (
	"context"
	"log"
	"net"

	"github.com/charconstpointer/raw/pkg/raw"
)

func main() {
	conn, err := net.Dial("tcp", ":5555")
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Println("pipes connected")

	pipe, err := raw.NewPipe(conn)
	if err := pipe.Run(context.Background()); err != nil {
		log.Fatal(err.Error())
	}
}
