package mysocks

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/txthinking/socks5"
)

const portOfTestServer = 9000

var proxyAddress = "127.0.0.1:" + strconv.Itoa(portOfTestServer)

var server *Server

func StartServer() {
	os.Setenv("MYSOCKS_PORT", strconv.Itoa(portOfTestServer))

	server = NewServer()
	go func() {
		err := server.Start()
		if err != nil {
			log.Printf("Error(Ignored): %v", err)
		}
	}()
	<-server.Ready()
}

func StopServer() {
	server.Close()
}

func TestConnect(t *testing.T) {
	StartServer()
	defer StopServer()

	socks5.Debug = true

	client, err := socks5.NewClient(proxyAddress, "", "", 0, 60)
	if err != nil {
		log.Println(err)
		return
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return client.Dial(network, addr)
			},
		},
	}
	res, err := httpClient.Get("https://ifconfig.co")
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

func TestUDPAssociate(t *testing.T) {
	StartServer()
	defer StopServer()

	socks5.Debug = true

	client, err := socks5.NewClient(proxyAddress, "", "", 0, 60)
	if err != nil {
		t.Fatal(err)
	}

	conn, err := client.Dial("udp", "8.8.8.8:53")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	message := []byte{
		0x12, 0x34, // ID
		0x01, 0x00, // Flags
		0x00, 0x01, // Questions
		0x00, 0x00, // Answer RRs
		0x00, 0x00, // Authority RRs
		0x00, 0x00, // Additional RRs
		0x03, 'w', 'w', 'w',
		0x06, 'g', 'o', 'o', 'g', 'l', 'e',
		0x03, 'c', 'o', 'm', 0x00, // "www.google.com"
		0x00, 0x01, // Type: A
		0x00, 0x01, // Class: IN
	}

	_, err = conn.Write(message)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}

	buffer := make([]byte, 512)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(buffer)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	if err := parseDNSResponse(buffer[:n]); err != nil {
		log.Fatalf("DNS response validation failed: %v", err)
	}
}

func parseDNSResponse(response []byte) error {
	// ヘッダー解析
	if len(response) < 12 {
		return fmt.Errorf("response too short")
	}
	transactionID := binary.BigEndian.Uint16(response[0:2])
	flags := binary.BigEndian.Uint16(response[2:4])
	questions := binary.BigEndian.Uint16(response[4:6])
	answerRRs := binary.BigEndian.Uint16(response[6:8])

	// ヘッダー情報を検証
	if transactionID != 0x1234 {
		return fmt.Errorf("unexpected transaction ID: %x", transactionID)
	}
	if flags&0x8000 == 0 {
		return fmt.Errorf("not a response packet")
	}
	if questions != 1 {
		return fmt.Errorf("unexpected number of questions: %d", questions)
	}
	if answerRRs == 0 {
		return fmt.Errorf("no answers in the response")
	}

	// クエッションセクションをスキップ
	offset := 12
	for response[offset] != 0 {
		offset += int(response[offset]) + 1
	}
	offset += 5 // 終端バイト + QTYPE (2) + QCLASS (2)

	// アンサーセクションをパース
	for i := 0; i < int(answerRRs); i++ {
		// 名前をスキップ
		if response[offset] == 0xc0 { // ポインタの場合
			offset += 2
		} else {
			for response[offset] != 0 {
				offset += int(response[offset]) + 1
			}
			offset++
		}

		// TYPE, CLASS, TTL, RDLENGTHを読み取る
		rType := binary.BigEndian.Uint16(response[offset : offset+2])
		// rClass := binary.BigEndian.Uint16(response[offset+2 : offset+4])
		// ttl := binary.BigEndian.Uint32(response[offset+4 : offset+8])
		rdLength := binary.BigEndian.Uint16(response[offset+8 : offset+10])
		offset += 10

		// Aレコードの場合、IPアドレスを検証
		if rType == 0x01 { // TYPE = A
			if rdLength != 4 {
				return fmt.Errorf("invalid A record length: %d", rdLength)
			}
			ip := net.IP(response[offset : offset+4])
			fmt.Printf("Found A record: %s\n", ip)
		}

		// レコードデータをスキップ
		offset += int(rdLength)
	}

	return nil
}
