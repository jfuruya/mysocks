package mysocks

import (
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"testing"

	"golang.org/x/net/proxy"
)

func TestServer(t *testing.T) {
	port := 9000

	server := NewServer(port)

	go func() {
		err := server.Start()
		if err != nil {
			log.Printf("Error(Ignored): %v", err)
		}
	}()

	<-server.Ready()

	proxyAddress := "127.0.0.1:" + strconv.Itoa(port)

	dialer, err := proxy.SOCKS5("tcp", proxyAddress, nil, proxy.Direct)
	if err != nil {
		t.Fatal(err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Dial: dialer.Dial,
		},
	}
	res, err := client.Get("https://ifconfig.co")
	if err != nil {
		log.Println(err)
		return
	}
	defer res.Body.Close()
	responseBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return
	}

	ipString := string(bytes.TrimSpace(responseBytes))
	ip := net.ParseIP(ipString)
	if ip == nil {
		t.Fatalf("Failed to parse the IP address: %s", ipString)
	}
}
