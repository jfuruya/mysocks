package mysocks

import (
	"fmt"
	"net"
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
	var waitGroup sync.WaitGroup

	waitGroup.Add(2)

	tcpListener, err := net.Listen("tcp", ":"+strconv.Itoa(server.port))
	if err != nil {
		return err
	}
	defer tcpListener.Close()

	server.tcpListener = &tcpListener

	logInfo(fmt.Sprintf("TCP server has been started on port %d.", server.port), nil)

	go func() {
		for {
			conn, err := tcpListener.Accept()
			if err != nil {
				logError(fmt.Sprintf("Failed to accept TCP connection: %v", err), nil)
				break
			}

			logInfo(fmt.Sprintf("A new TCP connection has been received from: %v", conn.RemoteAddr()), nil)

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

	logInfo(fmt.Sprintf("UDP server has been started on port %d.", server.port), nil)

	go func() {
		for {
			buf := make([]byte, 65507) // 65507 is the maximum UDP payload size
			n, addr, err := udpConn.ReadFromUDP(buf)
			if err != nil {
				logError(fmt.Sprintf("Failed to read UDP datagram: %v", err), nil)
				break
			}

			logInfo(fmt.Sprintf("A UDP data received from %s: %v", addr.String(), buf[:n]), nil)

			socksConnection := server.socksConnections.get(addr.IP)
			if socksConnection == nil {
				logError(fmt.Sprintf("There is no UDP association related to this remote address: %s", addr.String()), nil)
				continue
			}

			socksConnection.udpAssociation.clientAddr = addr

			datagram, err := newDatagramFrom(buf[:n])
			if err != nil {
				socksConnection.logWithLevel(logLevelError, fmt.Sprintf("Failed to create socks5 datagram: %v", err))
				continue
			}

			socksConnection.logWithLevel(logLevelInfo, fmt.Sprintf("A UDP datagram received from %s: %v", addr.String(), datagram))

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
		logError(fmt.Sprintf("Faild to close TCP listener: %v", err), nil)
	}

	err = server.udpConn.Close()
	if err != nil {
		logError(fmt.Sprintf("Failed to close UDP listener: %v", err), nil)
	}
}
