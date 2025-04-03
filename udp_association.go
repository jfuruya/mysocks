package mysocks

import "net"

type udpAssociation struct {
	clientAddr               *net.UDPAddr
	clientAddrForAccessLimit *net.UDPAddr
	association              chan byte
	destConn                 *net.UDPConn
}

func newUDPAssociation(clientAddrForAccessLimit *net.UDPAddr) *udpAssociation {
	return &udpAssociation{
		clientAddrForAccessLimit: clientAddrForAccessLimit,
		association:              make(chan byte),
	}
}

func (udpAssociation *udpAssociation) end() {
	close(udpAssociation.association)
	udpAssociation.destConn.Close()
}
