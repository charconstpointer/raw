package main

import (
	"log"

	"github.com/charconstpointer/raw/pkg/raw"
	"github.com/common-nighthawk/go-figure"
)

func main() {
	figure.NewColorFigure("agent", "slant", "green", true).Print()

	c, err := raw.NewClient("ec2-3-10-24-173.eu-west-2.compute.amazonaws.com:7000", ":25565")
	if err != nil {
		log.Fatal(err.Error())
	}
	c.Run()
}
