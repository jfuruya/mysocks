package mysocks

import (
	"log"
	"net"
	"time"
)

type socksConnection struct {
	clientTCPConn  *net.Conn
	server         *Server
	udpAssociation *udpAssociation
}

const udpTimeout = 60

func newSocksConnection(tcpConn *net.Conn, server *Server) *socksConnection {
	return &socksConnection{
		clientTCPConn: tcpConn,
		server:        server,
	}
}

func (socksConnection *socksConnection) String() string {
	if socksConnection == nil {
		return ""
	}
	return "remote address of TCP connection: " + (*socksConnection.clientTCPConn).RemoteAddr().String() + ", " + socksConnection.udpAssociation.String()
}

func (socksConnection *socksConnection) remoteIP() net.IP {
	return (*socksConnection.clientTCPConn).RemoteAddr().(*net.TCPAddr).IP
}

func (socksConnection *socksConnection) handle() {
	defer func() {
		socksConnectionDesc := socksConnection
		(*socksConnection.clientTCPConn).Close()
		log.Printf("TCP connection has been closed. %v", socksConnectionDesc)
	}()

	_, err := newNegotiationRequestFrom(*socksConnection.clientTCPConn)
	if err != nil {
		if err == errNegotiationMethodNotSupported {
			negotiationReply := newNegotiationReply(noAcceptable, socksConnection)
			if _, err := negotiationReply.WriteTo(*socksConnection.clientTCPConn); err != nil {
				log.Printf("Failed to write the negotiation reply. %v", socksConnection)
				return
			}
		}

		log.Printf("Failed to read the negotiation request. %v", socksConnection)
		return
	}

	negotiationReply := newNegotiationReply(supportedMethod, socksConnection)
	if _, err := negotiationReply.WriteTo(*socksConnection.clientTCPConn); err != nil {
		log.Printf("Failed to write the negotiation reply. %v", socksConnection)
		return
	}

	request, err := newRequestFrom(socksConnection)
	if err != nil {
		if err == errRequestCmdNotSupported {
			reply := newErrorReply(repCmdNotSupported, atypIPv4, socksConnection)
			if _, err := reply.WriteTo(*socksConnection.clientTCPConn); err != nil {
				log.Printf("Failed to write the reply. %v", socksConnection)
				return
			}
		}

		if err == errRequestAtypNotSupported {
			reply := newErrorReply(repAddrNotSupported, atypIPv4, socksConnection)
			if _, err := reply.WriteTo(*socksConnection.clientTCPConn); err != nil {
				log.Printf("Failed to write the reply. %v", socksConnection)
				return
			}
		}

		log.Printf("Failed to read the request. %v", socksConnection)
		return
	}

	err = request.processCmd()
	if err != nil {
		if err == errRequestNotReacheble {
			reply := newErrorReply(repHostUnreach, atypIPv4, socksConnection)
			if _, err := reply.WriteTo(*socksConnection.clientTCPConn); err != nil {
				log.Printf("Failed to write the reply.")
				return
			}
		}
	}
}

func (socksConnection *socksConnection) handleUDP(datagram *datagram) {
	if socksConnection.udpAssociation.destConn == nil {
		conn, err := net.Dial("udp", datagram.destAddress())
		if err != nil {
			log.Printf("Error: %v %v", err, socksConnection)
			return
		}

		log.Printf("A UDP socket has been created to: %s, from: %s\n", datagram.destAddress(), conn.LocalAddr().String())

		socksConnection.udpAssociation.destConn = conn.(*net.UDPConn)

		// Send the datagram from the destination server to the client

		go func() {
			defer func() {
				socksConnection.udpAssociation.destConn.Close()
				socksConnection.udpAssociation.destConn = nil
				log.Printf("UDP connection has been closed. %v", socksConnection)
			}()

			for {
				select {
				case <-socksConnection.udpAssociation.association:
					log.Printf("UDP association has been closed. %v", socksConnection)
					return
				default:
					if err := conn.SetDeadline(time.Now().Add(time.Duration(udpTimeout) * time.Second)); err != nil {
						return
					}

					buf := make([]byte, 65507) // 65507 is the maximum UDP payload size
					log.Printf("Waiting for a UDP data from %s", socksConnection.udpAssociation.destConn.RemoteAddr().String())
					n, err := socksConnection.udpAssociation.destConn.Read(buf)
					if err != nil {
						log.Printf("Failed to read UDP data from: %s: %v %v",
							socksConnection.udpAssociation.destConn.RemoteAddr().String(), err, socksConnection)
						return
					}

					log.Printf("A UDP data received from %s: %v", socksConnection.udpAssociation.destConn.RemoteAddr().String(), buf[:n])

					datagramSentToClient := newDatagram(datagram.dst, buf[:n])

					if _, err := socksConnection.server.udpConn.WriteToUDP(
						datagramSentToClient.bytes(),
						socksConnection.udpAssociation.clientAddr); err != nil {
						log.Printf("Failed to write UDP data to %s: %v %v", socksConnection.udpAssociation.clientAddr, err, socksConnection)
						return
					}
				}
			}
		}()
	}

	// Send the datagram from the client to the destination server

	select {
	case <-socksConnection.udpAssociation.association:
		log.Printf("UDP association has been closed. %v", socksConnection)
		return
	default:
		_, err := socksConnection.udpAssociation.destConn.Write(datagram.data)
		if err != nil {
			log.Printf("Failed to write UDP data to: %s, %v %v",
				socksConnection.udpAssociation.destConn.RemoteAddr().String(), err, socksConnection)
			return
		}
		log.Printf("A UDP data sent to %s: %v", socksConnection.udpAssociation.destConn.RemoteAddr().String(), datagram.data)
	}
}
