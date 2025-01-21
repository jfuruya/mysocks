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
	ErrRequestNotReacheble     = fmt.Errorf("the destination is not reachable")
	ErrRequestCmdNotSupported  = fmt.Errorf("the command is not supported")
	ErrRequestAtypNotSupported = fmt.Errorf("the address type is not supported")
)

type Request struct {
	Ver             byte
	Cmd             byte
	Rsv             byte // 0x00
	Atyp            byte
	DstAddr         []byte
	DstPort         []byte // 2 bytes
	SocksConnection *SocksConnection
}

const tcpTimeout = 60

func NewRequestFrom(socksConnection *SocksConnection) (*Request, error) {
	reader := *socksConnection.tcpConn
	verBytes := make([]byte, 1)
	if _, err := io.ReadFull(reader, verBytes); err != nil {
		return nil, err
	}
	ver := verBytes[0]
	if ver != Ver {
		return nil, fmt.Errorf("the value of the VER field in the negotiation request is invalid: %d", ver)
	}

	cmdBytes := make([]byte, 1)
	if _, err := io.ReadFull(reader, cmdBytes); err != nil {
		return nil, err
	}
	cmd := cmdBytes[0]
	if ver != Ver {
		return nil, fmt.Errorf("the value of the VER field in the negotiation request is invalid: %d", ver)
	}
	if cmd != SupportedCmd {
		return nil, ErrRequestCmdNotSupported
	}

	rsvBytes := make([]byte, 1)
	if _, err := io.ReadFull(reader, rsvBytes); err != nil {
		return nil, err
	}
	rsv := rsvBytes[0]
	if rsv != Rsv {
		return nil, fmt.Errorf("the value of the RSV field in the request is invalid: %d", rsv)
	}

	atypBytes := make([]byte, 1)
	if _, err := io.ReadFull(reader, atypBytes); err != nil {
		return nil, err
	}
	atyp := atypBytes[0]
	if atyp != ATYPIPv4 && atyp != ATYPIPv6 && atyp != ATYPDomain {
		return nil, ErrRequestAtypNotSupported
	}

	var dstAddr []byte
	if atyp == ATYPIPv4 {
		dstAddr = make([]byte, 4)
		if _, err := io.ReadFull(reader, dstAddr); err != nil {
			return nil, err
		}
	}
	if atyp == ATYPDomain {
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
	if atyp == ATYPIPv6 {
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

	return &Request{
		Ver:             ver,
		Cmd:             cmd,
		Rsv:             rsv,
		Atyp:            atyp,
		DstAddr:         dstAddr,
		DstPort:         dstPort,
		SocksConnection: socksConnection,
	}, nil
}

func (request *Request) ProcessCmd() error {
	switch request.Cmd {
	case CmdConnect:
		return request.handleConnect()
	default:
		return ErrRequestCmdNotSupported
	}
}

func (request *Request) handleConnect() error {
	conn, err := request.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	addr := conn.LocalAddr().(*net.TCPAddr).IP
	var atype byte
	if addr.To4() != nil {
		atype = ATYPIPv4
	} else {
		atype = ATYPIPv6
	}
	port := make([]byte, 2)
	binary.BigEndian.PutUint16(port, uint16(conn.LocalAddr().(*net.TCPAddr).Port))

	reply := NewReply(RepSucceeded, atype, addr, port)
	if _, err := reply.WriteTo(*request.SocksConnection.tcpConn); err != nil {
		log.Printf("Failed to write the reply.")
		return err
	}

	clientConn := *request.SocksConnection.tcpConn

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

func (request *Request) connect() (net.Conn, error) {
	conn, err := net.Dial("tcp", request.DestAddress())
	if err != nil {
		return nil, ErrRequestNotReacheble
	}
	return conn, nil
}

func (r *Request) DestAddress() string {
	var host string
	if r.Atyp == ATYPDomain {
		host = string(r.DstAddr)
	} else {
		host = net.IP(r.DstAddr).String()
	}
	port := strconv.Itoa(int(binary.BigEndian.Uint16(r.DstPort)))
	return net.JoinHostPort(host, port)
}
