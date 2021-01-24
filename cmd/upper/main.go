package main

import (
	"context"
	"log"
	"net"

	"github.com/charconstpointer/raw/pkg/raw"
)

func main() {
	up, err := net.Listen("tcp", ":5555")
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Println("listening on :5555")
	p, err := up.Accept()
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Println("accepted new conn, starting pipe")
	pipe, err := raw.NewPipe(p)
	go pipe.Run(context.Background())
	s, err := net.Listen("tcp", ":6000")
	if err != nil {
		log.Fatal(err.Error())
	}
	for {
		log.Println("listening for new conns")
		client, err := s.Accept()
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Println("adding new pipe")
		err = pipe.AddS(client)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}
