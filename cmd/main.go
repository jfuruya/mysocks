package main

import (
	"github.com/jfuruya/mysocks"
)

func main() {
	socksServer := mysocks.NewServer()
	err := socksServer.Start()
	if err != nil {
		panic(err)
	}
}
