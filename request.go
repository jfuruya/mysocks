package mysocks

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

var (
	errRequestNotReacheble     = fmt.Errorf("the destination is not reachable")
	errRequestCmdNotSupported  = fmt.Errorf("the command is not supported")
	errRequestAtypNotSupported = fmt.Errorf("the address type is not supported")
)

type request struct {
	ver byte
	cmd byte
	rsv byte // 0x00
	dst
	socksConnection *socksConnection
}

const tcpTimeout = 60

func newRequestFrom(socksConnection *socksConnection) (*request, error) {
	reader := *socksConnection.clientTCPConn
	verBytes := make([]byte, 1)
	if _, err := io.ReadFull(reader, verBytes); err != nil {
		return nil, err
	}
	ver := verBytes[0]
	if ver != fiexedVer {
		return nil, fmt.Errorf("the value of the VER field in the negotiation request is invalid: %d %v", ver, socksConnection)
	}

	cmdBytes := make([]byte, 1)
	if _, err := io.ReadFull(reader, cmdBytes); err != nil {
		return nil, err
	}
	cmd := cmdBytes[0]
	if !supportedCmd(cmd) {
		return nil, errRequestCmdNotSupported
	}

	rsvBytes := make([]byte, 1)
	if _, err := io.ReadFull(reader, rsvBytes); err != nil {
		return nil, err
	}
	rsv := rsvBytes[0]
	if rsv != fixedRsv {
		return nil, fmt.Errorf("the value of the RSV field in the request is invalid: %d %v", rsv, socksConnection)
	}

	atypBytes := make([]byte, 1)
	if _, err := io.ReadFull(reader, atypBytes); err != nil {
		return nil, err
	}
	atyp := atypBytes[0]
	if !supportedAtyp(atyp) {
		return nil, errRequestAtypNotSupported
	}

	dstAddr, err := readDestAddr(reader, atyp)
	if err != nil {
		return nil, err
	}

	dstPort := make([]byte, 2)
	if _, err := io.ReadFull(reader, dstPort); err != nil {
		return nil, err
	}

	socksConnection.logWithLevel(logLevelInfo,
		fmt.Sprintf("A request has been received. "+
			"VER: %#v CMD: %#v RSV: %#v ATYP: %#v DST.ADDR: %#v DST.PORT: %#v",
			ver, cmd, rsv, atyp, dstAddr, dstPort))

	return &request{
		ver: ver,
		cmd: cmd,
		rsv: rsv,
		dst: dst{
			atyp: atyp,
			addr: dstAddr,
			port: dstPort,
		},
		socksConnection: socksConnection,
	}, nil
}

func (request *request) processCmd() error {
	switch request.cmd {
	case cmdConnect:
		return request.handleConnect()
	case cmdAssociate:
		return request.handleUDPAssociate()
	default:
		return errRequestCmdNotSupported
	}
}

func (request *request) handleConnect() error {
	var err error

	conn, err := request.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	localAddrAsTCP := conn.LocalAddr().(*net.TCPAddr)
	err = request.replySuccess(localAddrAsTCP.IP, localAddrAsTCP.Port)
	if err != nil {
		return err
	}

	clientConn := *request.socksConnection.clientTCPConn

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

func (request *request) replySuccess(ip net.IP, port int) error {
	var atype byte
	if ip.To4() != nil {
		atype = atypIPv4
	} else if ip.To16() != nil {
		atype = atypIPv6
	} else {
		atype = atypDomain
	}

	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(port))

	reply := newReply(repSucceeded, atype, ip, portBytes, request.socksConnection)
	if _, err := reply.WriteTo(*request.socksConnection.clientTCPConn); err != nil {
		request.socksConnection.logWithLevel(logLevelError, "Failed to write the reply.")
		return err
	}
	return nil
}

func (request *request) connect() (net.Conn, error) {
	conn, err := net.Dial("tcp", request.destAddress())
	if err != nil {
		return nil, errRequestNotReacheble

	}
	request.socksConnection.logWithLevel(logLevelInfo,
		fmt.Sprintf("A TCP connection has been established to: %s", request.destAddress()))

	return conn, nil
}

func (request *request) handleUDPAssociate() error {
	var clientAddrForAccessLimit *net.UDPAddr
	var err error
	if bytes.Equal(request.dst.port, []byte{0x00, 0x00}) {
		clientAddrForAccessLimit, err = net.ResolveUDPAddr("udp", (*request.socksConnection.clientTCPConn).RemoteAddr().String())
	} else {
		clientAddrForAccessLimit, err = net.ResolveUDPAddr("udp", request.destAddress())
	}
	if err != nil {
		return errRequestNotReacheble
	}

	serverAddrAsUDP := (*request.socksConnection.server.udpConn).LocalAddr().(*net.UDPAddr)
	err = request.replySuccess(net.IP(request.socksConnection.server.hostName), serverAddrAsUDP.Port)
	if err != nil {
		return err
	}

	udpAssociation := *newUDPAssociation(clientAddrForAccessLimit)
	defer udpAssociation.end()

	request.socksConnection.udpAssociation = &udpAssociation

	io.Copy(io.Discard, *request.socksConnection.clientTCPConn)

	request.socksConnection.logWithLevel(logLevelInfo,
		fmt.Sprintf("A TCP connection that associated with UDP has been closed. %s", clientAddrForAccessLimit))

	return nil
}
