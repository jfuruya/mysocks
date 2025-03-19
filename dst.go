package mysocks

import (
	"encoding/binary"
	"net"
	"strconv"
)

type dst struct {
	atyp byte
	addr []byte
	port []byte // 2 bytes
}

func (d *dst) destAddress() string {
	var host string
	if d.atyp == atypDomain {
		host = string(d.addr)
	} else {
		host = net.IP(d.addr).String()
	}
	port := strconv.Itoa(int(binary.BigEndian.Uint16(d.port)))
	return net.JoinHostPort(host, port)
}
