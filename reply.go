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
	bndPort []byte // 2 bytes
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

func newReply(rep byte, atype byte, bndAddr []byte, bndPort []byte) *reply {
	return &reply{
		ver:     fiexedVer,
		rep:     rep,
		rsv:     fixedRsv,
		atyp:    atype,
		bndAddr: bndAddr,
		bndPort: bndPort,
	}
}

func newErrorReply(rep byte, atype byte) *reply {
	var bndAddr []byte
	if atype == atypeDomain || atype == atypeIPv4 {
		bndAddr = []byte{0x00, 0x00, 0x00, 0x00}
	} else {
		bndAddr = []byte(net.IPv6zero)
	}

	bndPort := []byte{0x00, 0x00}

	return &reply{
		ver:     fiexedVer,
		rep:     rep,
		rsv:     fixedRsv,
		atyp:    atype,
		bndAddr: bndAddr,
		bndPort: bndPort,
	}
}

func (r *reply) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(append(append([]byte{r.ver, r.rep, r.rsv, r.atyp}, r.bndAddr...), r.bndPort...))
	if err != nil {
		return 0, err
	}

	log.Printf("Reply sent. VER: %#v REP: %#v RSV: %#v ATYPE: %#v BND.ARRR: %#v BND.PORT: %#v\n",
		r.ver, r.rep, r.rsv, r.atyp, r.bndAddr, r.bndPort)

	return int64(n), nil
}
