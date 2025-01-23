package mysocks

import (
	"log"
	"net"
)

type socksConnection struct {
	tcpConn *net.Conn
}

func newSocksConnection(tcpConn *net.Conn) *socksConnection {
	return &socksConnection{
		tcpConn: tcpConn,
	}
}

func (sockesConnection *socksConnection) handle() {
	defer func() {
		(*sockesConnection.tcpConn).Close()
		log.Printf("The client has disconnected.")
	}()

	log.Printf("A new client has connected.")

	_, err := newNegotiationRequestFrom(*sockesConnection.tcpConn)
	if err != nil {
		if err == errNegotiationMethodNotSupported {
			negotiationReply := newNegotiationReply(noAcceptable)
			if _, err := negotiationReply.WriteTo(*sockesConnection.tcpConn); err != nil {
				log.Printf("Failed to write the negotiation reply.")
				return
			}
		}

		log.Printf("Failed to read the negotiation request.")
		return
	}

	negotiationReply := newNegotiationReply(SupportedMethod)
	if _, err := negotiationReply.WriteTo(*sockesConnection.tcpConn); err != nil {
		log.Printf("Failed to write the negotiation reply.")
		return
	}

	request, err := newRequestFrom(sockesConnection)
	if err != nil {
		if err == errRequestCmdNotSupported {
			reply := newErrorReply(repCmdNotSupported, atypeIPv4)
			if _, err := reply.WriteTo(*sockesConnection.tcpConn); err != nil {
				log.Printf("Failed to write the reply.")
				return
			}
		}

		if err == errRequestAtypNotSupported {
			reply := newErrorReply(repAddrNotSupported, atypeIPv4)
			if _, err := reply.WriteTo(*sockesConnection.tcpConn); err != nil {
				log.Printf("Failed to write the reply.")
				return
			}
		}

		log.Printf("Failed to read the request.")
		return
	}

	err = request.processCmd()
	if err != nil {
		if err == errRequestNotReacheble {
			reply := newErrorReply(repHostUnreach, atypeIPv4)
			if _, err := reply.WriteTo(*sockesConnection.tcpConn); err != nil {
				log.Printf("Failed to write the reply.")
				return
			}
		}
	}
}
