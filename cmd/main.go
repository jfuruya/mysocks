package main

import (
	"github.com/jfuruya/mysocks"
)

func main() {
	socksServer := mysocks.NewServer(58080)
	err := socksServer.Start()
	if err != nil {
		panic(err)
	}
}
