package mysocks

import (
	"log"
	"net"
	"os"
	"strconv"
	"sync"
)

type Server struct {
	port             int
	hostName         string
	ready            chan struct{}
	tcpListener      *net.Listener
	udpConn          *net.UDPConn
	socksConnections socksConnections
}

func NewServer() *Server {
	return &Server{
		port:             portFromEnv(),
		ready:            make(chan struct{}),
		socksConnections: *newSocksConnections(),
		hostName:         hostNameFromEnv(),
	}
}

func (server *Server) Start() error {
	log.SetOutput(os.Stdout)

	var waitGroup sync.WaitGroup

	waitGroup.Add(2)

	tcpListener, err := net.Listen("tcp", ":"+strconv.Itoa(server.port))
	if err != nil {
		return err
	}
	defer tcpListener.Close()

	server.tcpListener = &tcpListener

	log.Printf("TCP server has been started on port %d.", server.port)

	go func() {
		for {
			conn, err := tcpListener.Accept()
			if err != nil {
				log.Printf("Failed to accept TCP connection: %v", err)
				break
			}

			log.Printf("A new TCP connection has been received from: %v", conn.RemoteAddr())

			socksConnection := newSocksConnection(&conn, server)

			server.socksConnections.add(socksConnection)

			go func() {
				socksConnection.handle()
				defer server.socksConnections.remove(socksConnection)
			}()
		}
		server.socksConnections.closeAll()
		waitGroup.Done()
	}()

	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{Port: server.port})
	if err != nil {
		return err
	}
	defer udpConn.Close()

	server.udpConn = udpConn

	log.Printf("UDP server has been started on port %d.", server.port)

	go func() {
		for {
			buf := make([]byte, 65507) // 65507 is the maximum UDP payload size
			n, addr, err := udpConn.ReadFromUDP(buf)
			if err != nil {
				log.Printf("Failed to read UDP datagram: %v", err)
				break
			}

			log.Printf("A UDP data received from %s: %v", addr.String(), buf[:n])

			socksConnection := server.socksConnections.get(addr.IP)
			if socksConnection == nil {
				log.Printf("Error: There is no UDP association related to this remote address: %s", addr.String())
				continue
			}

			socksConnection.udpAssociation.clientAddr = addr

			datagram, err := newDatagramFrom(buf[:n])
			if err != nil {
				log.Printf("Failed to create socks5 datagram: %v", err)
				continue
			}

			log.Printf("A UDP datagram received from %s: %v", addr.String(), datagram)

			go socksConnection.handleUDP(datagram)
		}

		waitGroup.Done()
	}()

	close(server.ready)

	waitGroup.Wait()

	return nil
}

func (server *Server) Ready() <-chan struct{} {
	return server.ready
}

func (server *Server) Close() {
	var err error

	err = (*server.tcpListener).Close()
	if err != nil {
		log.Printf("Faild to close TCP listener: %v", err)
	}

	err = server.udpConn.Close()
	if err != nil {
		log.Printf("Failed to close UDP listener: %v", err)
	}
}
