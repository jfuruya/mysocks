package mysocks

import (
	"log"
	"net"
)

type socksConnections struct {
	connections map[string]*socksConnection
}

func newSocksConnections() *socksConnections {
	return &socksConnections{
		connections: make(map[string]*socksConnection),
	}
}

func (socksConnections *socksConnections) add(sc *socksConnection) {
	log.Printf("Remote IP remembered: %v", sc.remoteIP())
	socksConnections.connections[sc.remoteIP().String()] = sc
}

func (socksConnections *socksConnections) remove(sc *socksConnection) {
	log.Printf("Remote IP forgotten: %v", sc.remoteIP())
	delete(socksConnections.connections, sc.remoteIP().String())
}

func (socksConnections *socksConnections) get(remoteIP net.IP) *socksConnection {
	return socksConnections.connections[remoteIP.String()]
}

func (socksConnections *socksConnections) closeAll() {
	for _, sc := range socksConnections.connections {
		(*sc.clientTCPConn).Close()
	}
}
