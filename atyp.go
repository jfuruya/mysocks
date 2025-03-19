package mysocks

import (
	"fmt"
	"io"
)

const (
	atypIPv4   byte = 0x01
	atypDomain byte = 0x03
	atypIPv6   byte = 0x04
)

func supportedAtyp(atyp byte) bool {
	return atyp == atypIPv4 || atyp == atypIPv6 || atyp == atypDomain
}

func readDestAddr(reader io.Reader, atyp byte) ([]byte, error) {
	var dstAddr []byte
	if atyp == atypIPv4 {
		dstAddr = make([]byte, 4)
		if _, err := io.ReadFull(reader, dstAddr); err != nil {
			return nil, err
		}
	}
	if atyp == atypDomain {
		dstAddrLengthBytes := make([]byte, 1)
		if _, err := io.ReadFull(reader, dstAddrLengthBytes); err != nil {
			return nil, err
		}
		dstAddrLenth := dstAddrLengthBytes[0]
		if dstAddrLenth == 0 {
			return nil, fmt.Errorf("the value of the first byte of ATYPE field in the request is invalid: %d", dstAddrLenth)
		}
		dstAddr = make([]byte, int(dstAddrLenth))
		if _, err := io.ReadFull(reader, dstAddr); err != nil {
			return nil, err
		}
	}
	if atyp == atypIPv6 {
		dstAddr = make([]byte, 16)
		if _, err := io.ReadFull(reader, dstAddr); err != nil {
			return nil, err
		}
	}
	return dstAddr, nil
}

func readDestAddrFromBytes(bytes []byte, atyp byte) (dstAddr []byte, dstAddrLength int, error error) {
	if atyp == atypIPv4 {
		dstAddrLength = 4
	}
	if atyp == atypDomain {
		dstAddrLength = int(bytes[0])
		if dstAddrLength == 0 {
			error = fmt.Errorf("the value of the first byte of ATYPE field in the request is invalid: %d", dstAddrLength)
			return
		}
	}
	if atyp == atypIPv6 {
		dstAddrLength = 16
	}
	dstAddr = bytes[:dstAddrLength]
	return
}
