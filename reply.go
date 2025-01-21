package mysocks

import (
	"io"
	"log"
	"net"
)

// Reply is the reply packet
type Reply struct {
	Ver  byte
	Rep  byte
	Rsv  byte // 0x00
	Atyp byte
	// CONNECT socks server's address which used to connect to dst addr
	// BIND ...
	// UDP socks server's address which used to connect to dst addr
	BndAddr []byte
	// CONNECT socks server's port which used to connect to dst addr
	// BIND ...
	// UDP socks server's port which used to connect to dst addr
	BndPort []byte // 2 bytes
}

const (
	RepSucceeded        byte = 0x00
	RepGeneral          byte = 0x01
	RepDenied           byte = 0x02
	RepNetUnreach       byte = 0x03
	RepHostUnreach      byte = 0x04
	RepConnRefused      byte = 0x05
	RepTTLExpired       byte = 0x06
	RepCmdNotSupported  byte = 0x07
	RepAddrNotSupported byte = 0x08
)

func NewReply(rep byte, atype byte, bndAddr []byte, bndPort []byte) *Reply {
	return &Reply{
		Ver:     Ver,
		Rep:     rep,
		Rsv:     0x00,
		Atyp:    atype,
		BndAddr: bndAddr,
		BndPort: bndPort,
	}
}

func NewErrorReply(rep byte, atype byte) *Reply {
	var bndAddr []byte
	if atype == ATYPDomain || atype == ATYPIPv4 {
		bndAddr = []byte{0x00, 0x00, 0x00, 0x00}
	} else {
		bndAddr = []byte(net.IPv6zero)
	}

	bndPort := []byte{0x00, 0x00}

	return &Reply{
		Ver:     Ver,
		Rep:     rep,
		Rsv:     0x00,
		Atyp:    atype,
		BndAddr: bndAddr,
		BndPort: bndPort,
	}
}

func (r *Reply) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(append(append([]byte{r.Ver, r.Rep, r.Rsv, r.Atyp}, r.BndAddr...), r.BndPort...))
	if err != nil {
		return 0, err
	}

	log.Printf("Reply sent. VER: %#v REP: %#v RSV: %#v ATYPE: %#v BND.ARRR: %#v BND.PORT: %#v\n",
		r.Ver, r.Rep, r.Rsv, r.Atyp, r.BndAddr, r.BndPort)

	return int64(n), nil
}
