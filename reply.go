package mysocks

import (
	"io"
	"log"
	"net"
)

// reply is the reply packet
type reply struct {
	ver  byte
	rep  byte
	rsv  byte // 0x00
	atyp byte
	// CONNECT socks server's address which used to connect to dst addr
	// BIND ...
	// UDP socks server's address which used to connect to dst addr
	bndAddr []byte
	// CONNECT socks server's port which used to connect to dst addr
	// BIND ...
	// UDP socks server's port which used to connect to dst addr
	bndPort         []byte // 2 bytes
	socksConnection *socksConnection
}

const (
	repSucceeded        byte = 0x00
	repGeneral          byte = 0x01
	repDenied           byte = 0x02
	repNetUnreach       byte = 0x03
	repHostUnreach      byte = 0x04
	repConnRefused      byte = 0x05
	repTTLExpired       byte = 0x06
	repCmdNotSupported  byte = 0x07
	repAddrNotSupported byte = 0x08
)

func newReply(rep byte, atype byte, bndAddr []byte, bndPort []byte, socksConnection *socksConnection) *reply {
	if atype == atypDomain {
		bndAddr = append([]byte{byte(len(bndAddr))}, bndAddr...)
	}
	return &reply{
		ver:             fiexedVer,
		rep:             rep,
		rsv:             fixedRsv,
		atyp:            atype,
		bndAddr:         bndAddr,
		bndPort:         bndPort,
		socksConnection: socksConnection,
	}
}

func newErrorReply(rep byte, atype byte, socksConnection *socksConnection) *reply {
	var bndAddr []byte
	if atype == atypDomain || atype == atypIPv4 {
		bndAddr = []byte{0x00, 0x00, 0x00, 0x00}
	} else {
		bndAddr = []byte(net.IPv6zero)
	}

	bndPort := []byte{0x00, 0x00}

	return &reply{
		ver:             fiexedVer,
		rep:             rep,
		rsv:             fixedRsv,
		atyp:            atype,
		bndAddr:         bndAddr,
		bndPort:         bndPort,
		socksConnection: socksConnection,
	}
}

func (reply *reply) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(append(append([]byte{reply.ver, reply.rep, reply.rsv, reply.atyp}, reply.bndAddr...), reply.bndPort...))
	if err != nil {
		return 0, err
	}

	log.Printf("Reply sent. VER: %#v REP: %#v RSV: %#v ATYPE: %#v BND.ARRR: %#v BND.PORT: %#v %v\n",
		reply.ver, reply.rep, reply.rsv, reply.atyp, reply.bndAddr, reply.bndPort, reply.socksConnection)

	return int64(n), nil
}
