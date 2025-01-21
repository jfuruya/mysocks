package mysocks

import (
	"log"
	"net"
)

type SocksConnection struct {
	tcpConn *net.Conn
}

func NewSocksConnection(tcpConn *net.Conn) *SocksConnection {
	return &SocksConnection{
		tcpConn: tcpConn,
	}
}

func (sockesConnection *SocksConnection) Handle() {
	defer func() {
		(*sockesConnection.tcpConn).Close()
		log.Printf("The client has disconnected.")
	}()

	log.Printf("A new client has connected.")

	_, err := NewNegotiationRequestFrom(*sockesConnection.tcpConn)
	if err != nil {
		if err == ErrNegotiationMethodNotSupported {
			negotiationReply := NewNegotiationReply(NoAcceptable)
			if _, err := negotiationReply.WriteTo(*sockesConnection.tcpConn); err != nil {
				log.Printf("Failed to write the negotiation reply.")
				return
			}
		}

		log.Printf("Failed to read the negotiation request.")
		return
	}

	negotiationReply := NewNegotiationReply(SupportedMethod)
	if _, err := negotiationReply.WriteTo(*sockesConnection.tcpConn); err != nil {
		log.Printf("Failed to write the negotiation reply.")
		return
	}

	request, err := NewRequestFrom(sockesConnection)
	if err != nil {
		if err == ErrRequestCmdNotSupported {
			reply := NewErrorReply(RepCmdNotSupported, ATYPIPv4)
			if _, err := reply.WriteTo(*sockesConnection.tcpConn); err != nil {
				log.Printf("Failed to write the reply.")
				return
			}
		}

		if err == ErrRequestAtypNotSupported {
			reply := NewErrorReply(RepAddrNotSupported, ATYPIPv4)
			if _, err := reply.WriteTo(*sockesConnection.tcpConn); err != nil {
				log.Printf("Failed to write the reply.")
				return
			}
		}

		log.Printf("Failed to read the request.")
		return
	}

	err = request.ProcessCmd()
	if err != nil {
		if err == ErrRequestNotReacheble {
			reply := NewErrorReply(RepHostUnreach, ATYPIPv4)
			if _, err := reply.WriteTo(*sockesConnection.tcpConn); err != nil {
				log.Printf("Failed to write the reply.")
				return
			}
		}
	}
}
