package mysocks

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"
)

var (
	errRequestNotReacheble     = fmt.Errorf("the destination is not reachable")
	errRequestCmdNotSupported  = fmt.Errorf("the command is not supported")
	errRequestAtypNotSupported = fmt.Errorf("the address type is not supported")
)

type request struct {
	ver             byte
	cmd             byte
	rsv             byte // 0x00
	atyp            byte
	dstAddr         []byte
	dstPort         []byte // 2 bytes
	socksConnection *socksConnection
}

const tcpTimeout = 60

func newRequestFrom(socksConnection *socksConnection) (*request, error) {
	reader := *socksConnection.tcpConn
	verBytes := make([]byte, 1)
	if _, err := io.ReadFull(reader, verBytes); err != nil {
		return nil, err
	}
	ver := verBytes[0]
	if ver != fiexedVer {
		return nil, fmt.Errorf("the value of the VER field in the negotiation request is invalid: %d", ver)
	}

	cmdBytes := make([]byte, 1)
	if _, err := io.ReadFull(reader, cmdBytes); err != nil {
		return nil, err
	}
	cmd := cmdBytes[0]
	if cmd != supportedCmd {
		return nil, errRequestCmdNotSupported
	}

	rsvBytes := make([]byte, 1)
	if _, err := io.ReadFull(reader, rsvBytes); err != nil {
		return nil, err
	}
	rsv := rsvBytes[0]
	if rsv != fixedRsv {
		return nil, fmt.Errorf("the value of the RSV field in the request is invalid: %d", rsv)
	}

	atypBytes := make([]byte, 1)
	if _, err := io.ReadFull(reader, atypBytes); err != nil {
		return nil, err
	}
	atyp := atypBytes[0]
	if atyp != atypeIPv4 && atyp != atypeIPv6 && atyp != atypeDomain {
		return nil, errRequestAtypNotSupported
	}

	var dstAddr []byte
	if atyp == atypeIPv4 {
		dstAddr = make([]byte, 4)
		if _, err := io.ReadFull(reader, dstAddr); err != nil {
			return nil, err
		}
	}
	if atyp == atypeDomain {
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
	if atyp == atypeIPv6 {
		dstAddr = make([]byte, 16)
		if _, err := io.ReadFull(reader, dstAddr); err != nil {
			return nil, err
		}
	}

	dstPort := make([]byte, 2)
	if _, err := io.ReadFull(reader, dstPort); err != nil {
		return nil, err
	}

	log.Printf("A request has been received. "+
		"VER: %#v CMD: %#v RSV: %#v ATYP: %#v DST.ADDR: %#v DST.PORT: %#v\n",
		ver, cmd, rsv, atyp, dstAddr, dstPort)

	return &request{
		ver:             ver,
		cmd:             cmd,
		rsv:             rsv,
		atyp:            atyp,
		dstAddr:         dstAddr,
		dstPort:         dstPort,
		socksConnection: socksConnection,
	}, nil
}

func (request *request) processCmd() error {
	switch request.cmd {
	case cmdConnect:
		return request.handleConnect()
	default:
		return errRequestCmdNotSupported
	}
}

func (request *request) handleConnect() error {
	conn, err := request.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	addr := conn.LocalAddr().(*net.TCPAddr).IP
	var atype byte
	if addr.To4() != nil {
		atype = atypeIPv4
	} else {
		atype = atypeIPv6
	}
	port := make([]byte, 2)
	binary.BigEndian.PutUint16(port, uint16(conn.LocalAddr().(*net.TCPAddr).Port))

	reply := newReply(repSucceeded, atype, addr, port)
	if _, err := reply.WriteTo(*request.socksConnection.tcpConn); err != nil {
		log.Printf("Failed to write the reply.")
		return err
	}

	clientConn := *request.socksConnection.tcpConn

	go func() {
		var bf [1024 * 2]byte
		for {
			if err := conn.SetDeadline(time.Now().Add(time.Duration(tcpTimeout) * time.Second)); err != nil {
				return
			}
			i, err := conn.Read(bf[:])
			if err != nil {
				return
			}
			if _, err := clientConn.Write(bf[0:i]); err != nil {
				return
			}
		}
	}()

	var bf [1024 * 2]byte
	for {
		if err := clientConn.SetDeadline(time.Now().Add(time.Duration(tcpTimeout) * time.Second)); err != nil {
			return err
		}
		i, err := clientConn.Read(bf[:])
		if err != nil {
			return err
		}
		if _, err := conn.Write(bf[0:i]); err != nil {
			return err
		}
	}
}

func (request *request) connect() (net.Conn, error) {
	conn, err := net.Dial("tcp", request.destAddress())
	if err != nil {
		return nil, errRequestNotReacheble
	}
	return conn, nil
}

func (r *request) destAddress() string {
	var host string
	if r.atyp == atypeDomain {
		host = string(r.dstAddr)
	} else {
		host = net.IP(r.dstAddr).String()
	}
	port := strconv.Itoa(int(binary.BigEndian.Uint16(r.dstPort)))
	return net.JoinHostPort(host, port)
}
