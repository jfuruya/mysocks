package mysocks

import (
	"fmt"
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

func (socksConnection *socksConnection) logWithLevel(level int, message string) {
	fields := map[string]interface{}{
		"clientAddressOfTCPConnection": (*socksConnection.clientTCPConn).RemoteAddr().String(),
	}
	if socksConnection.udpAssociation != nil {
		fields["addressOfClientUDPSocket"] = socksConnection.udpAssociation.clientAddr.String()
	}

	logWithLevel(level, message, fields)
}

func (socksConnection *socksConnection) remoteIP() net.IP {
	return (*socksConnection.clientTCPConn).RemoteAddr().(*net.TCPAddr).IP
}

func (socksConnection *socksConnection) handle() {
	defer func() {
		(*socksConnection.clientTCPConn).Close()
		socksConnection.logWithLevel(logLevelInfo, "TCP connection has been closed.")
	}()

	negotiationRequest, err := newNegotiationRequestFrom(socksConnection)
	if err != nil {
		if err == errNegotiationMethodNotSupported {
			negotiationReply := newNegotiationReply(noAcceptable, socksConnection)
			if _, err := negotiationReply.WriteTo(*socksConnection.clientTCPConn); err != nil {
				socksConnection.logWithLevel(logLevelError, "Failed to write the negotiation reply.")
				return
			}
		}

		socksConnection.logWithLevel(logLevelError, "Failed to read the negotiation request.")
		return
	}

	negotiationReply := newNegotiationReply(methodToUseIn(negotiationRequest.methods), socksConnection)
	if _, err := negotiationReply.WriteTo(*socksConnection.clientTCPConn); err != nil {
		socksConnection.logWithLevel(logLevelError, "Failed to write the negotiation reply.")
		return
	}

	if methodNeedsUserPasswordAuth(negotiationReply.method) {
		userPasswordAuthRequest, err := newUserPasswordAuthRequestFrom(socksConnection)
		if err != nil {
			socksConnection.logWithLevel(logLevelError, "Failed to read the user password authentication request.")
			return
		}

		userName := userPasswordAuthRequest.usernameAsString()
		password := userPasswordAuthRequest.passwordAsString()

		authSuccess := authenticate(userName, password)

		userPasswordAuthReply := newUserPasswordAuthReply(authSuccess, socksConnection)
		if _, err := userPasswordAuthReply.WriteTo(*socksConnection.clientTCPConn); err != nil {
			socksConnection.logWithLevel(logLevelError, "Failed to write the authentication reply.")
			return
		}
	}

	request, err := newRequestFrom(socksConnection)
	if err != nil {
		if err == errRequestCmdNotSupported {
			reply := newErrorReply(repCmdNotSupported, atypIPv4, socksConnection)
			if _, err := reply.WriteTo(*socksConnection.clientTCPConn); err != nil {
				socksConnection.logWithLevel(logLevelError, "Failed to write the reply.")
				return
			}
		}

		if err == errRequestAtypNotSupported {
			reply := newErrorReply(repAddrNotSupported, atypIPv4, socksConnection)
			if _, err := reply.WriteTo(*socksConnection.clientTCPConn); err != nil {
				socksConnection.logWithLevel(logLevelError, "Failed to write the reply.")
				return
			}
		}

		socksConnection.logWithLevel(logLevelError, "Failed to read the request.")
		return
	}

	err = request.processCmd()
	if err != nil {
		if err == errRequestNotReacheble {
			reply := newErrorReply(repHostUnreach, atypIPv4, socksConnection)
			if _, err := reply.WriteTo(*socksConnection.clientTCPConn); err != nil {
				socksConnection.logWithLevel(logLevelError, "Failed to write the reply.")
				return
			}
		}
	}
}

func (socksConnection *socksConnection) handleUDP(datagram *datagram) {
	if socksConnection.udpAssociation.destConn == nil {
		conn, err := net.Dial("udp", datagram.destAddress())
		if err != nil {
			socksConnection.logWithLevel(logLevelError, fmt.Sprintf("Error: %v", err))
			return
		}

		socksConnection.logWithLevel(logLevelInfo,
			fmt.Sprintf("A UDP socket has been created to: %s, from: %s", datagram.destAddress(), conn.LocalAddr().String()))

		socksConnection.udpAssociation.destConn = conn.(*net.UDPConn)

		// Send the datagram from the destination server to the client

		go func() {
			defer func() {
				socksConnection.udpAssociation.destConn.Close()
				socksConnection.udpAssociation.destConn = nil
				socksConnection.logWithLevel(logLevelInfo, "UDP connection has been closed.")
			}()

			for {
				select {
				case <-socksConnection.udpAssociation.association:
					socksConnection.logWithLevel(logLevelInfo, "UDP association has been closed.")
					return
				default:
					if err := conn.SetDeadline(time.Now().Add(time.Duration(udpTimeout) * time.Second)); err != nil {
						return
					}

					buf := make([]byte, 65507) // 65507 is the maximum UDP payload size
					socksConnection.logWithLevel(logLevelInfo,
						fmt.Sprintf("Waiting for a UDP data from %s", socksConnection.udpAssociation.destConn.RemoteAddr().String()))
					n, err := socksConnection.udpAssociation.destConn.Read(buf)
					if err != nil {
						socksConnection.logWithLevel(logLevelError,
							fmt.Sprintf("Failed to read UDP data from '%s': %v",
								socksConnection.udpAssociation.destConn.RemoteAddr().String(), err))
						return
					}

					socksConnection.logWithLevel(logLevelInfo,
						fmt.Sprintf("A UDP data received from '%s': %v", socksConnection.udpAssociation.destConn.RemoteAddr().String(), buf[:n]))

					datagramSentToClient := newDatagram(datagram.dst, buf[:n])

					if _, err := socksConnection.server.udpConn.WriteToUDP(
						datagramSentToClient.bytes(),
						socksConnection.udpAssociation.clientAddr); err != nil {
						socksConnection.logWithLevel(logLevelError,
							fmt.Sprintf("Failed to write UDP data to '%s': %v", socksConnection.udpAssociation.clientAddr, err))
						return
					}
				}
			}
		}()
	}

	// Send the datagram from the client to the destination server

	select {
	case <-socksConnection.udpAssociation.association:
		socksConnection.logWithLevel(logLevelInfo, "UDP association has been closed.")
		return
	default:
		_, err := socksConnection.udpAssociation.destConn.Write(datagram.data)
		if err != nil {
			socksConnection.logWithLevel(logLevelError,
				fmt.Sprintf("Failed to write UDP data to '%s': %v",
					socksConnection.udpAssociation.destConn.RemoteAddr().String(), err))
			return
		}
		socksConnection.logWithLevel(logLevelInfo,
			fmt.Sprintf("A UDP data sent to '%s': %v", socksConnection.udpAssociation.destConn.RemoteAddr().String(), datagram.data))
	}
}
