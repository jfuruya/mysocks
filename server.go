package mysocks

import (
	"log"
	"net"
	"os"
	"strconv"
)

type Server struct {
	port     int
	ready    chan struct{}
	listener *net.Listener
}

func NewServer() *Server {
	return &Server{
		port:  portFromEnv(),
		ready: make(chan struct{}),
	}
}

func (s *Server) Start() error {
	log.SetOutput(os.Stdout)

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(s.port))
	if err != nil {
		return err
	}
	defer listener.Close()

	s.listener = &listener

	log.Printf("INFO: サーバーが %d ポートで起動しました。", s.port)

	close(s.ready)

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		socksConnection := newSocksConnection(&conn)

		go socksConnection.handle()
	}
}

func (s *Server) Ready() <-chan struct{} {
	return s.ready
}

func (s *Server) Close() {
	(*s.listener).Close()
}
