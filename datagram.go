package mysocks

import (
	"fmt"
)

type datagram struct {
	rsv  []byte // 0x00 0x00
	frag byte
	dst
	data []byte
}

func newDatagramFrom(bytes []byte) (*datagram, error) {
	rsv := bytes[:2]
	if rsv[0] != 0x00 || rsv[1] != 0x00 {
		return nil, fmt.Errorf("the value of the RSV field in the request is invalid: %d", rsv)
	}

	frag := bytes[2]
	// Fragmentation is not supported
	if frag != 0x00 {
		return nil, fmt.Errorf("the value of the FRAG field in the request is invalid: %d", frag)
	}

	atyp := bytes[3]
	if !supportedAtyp(atyp) {
		return nil, errRequestAtypNotSupported
	}

	dstAddr, dstAddrLength, err := readDestAddrFromBytes(bytes[4:], atyp)
	if err != nil {
		return nil, err
	}

	dstPort := bytes[4+dstAddrLength : 4+dstAddrLength+2]

	data := bytes[4+dstAddrLength+2:]

	return &datagram{
		rsv:  rsv,
		frag: frag,
		dst: dst{
			atyp: atyp,
			addr: dstAddr,
			port: dstPort,
		},
		data: data,
	}, nil
}

func newDatagram(dst dst, data []byte) *datagram {
	return &datagram{
		rsv:  []byte{0x00, 0x00},
		frag: 0x00,
		dst:  dst,
		data: data,
	}
}

func (d *datagram) bytes() []byte {
	bytes := make([]byte, 0)
	bytes = append(bytes, d.rsv...)
	bytes = append(bytes, d.frag)
	bytes = append(bytes, d.dst.atyp)
	bytes = append(bytes, d.dst.addr...)
	bytes = append(bytes, d.dst.port...)
	bytes = append(bytes, d.data...)
	return bytes
}
