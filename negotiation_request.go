package mysocks

import (
	"errors"
	"fmt"
	"io"
)

var errNegotiationMethodNotSupported = errors.New("the method is not supported")

type negotiationRequest struct {
	ver             byte
	nmethods        byte
	methods         []byte
	socksConnection *socksConnection
}

func newNegotiationRequestFrom(socksConnection *socksConnection) (*negotiationRequest, error) {
	reader := *socksConnection.clientTCPConn

	verBytes := make([]byte, 1)
	if _, err := io.ReadFull(reader, verBytes); err != nil {
		return nil, err
	}
	ver := verBytes[0]
	if ver != fiexedVer {
		return nil, fmt.Errorf("the value of the VER field in the negotiation request is invalid: %d", ver)
	}

	nmethodsBytes := make([]byte, 1)
	if _, err := io.ReadFull(reader, nmethodsBytes); err != nil {
		return nil, err
	}
	nmethods := nmethodsBytes[0]
	if nmethods == 0 {
		return nil, fmt.Errorf("the value of the NMETHODS field in the negotiation request is invalid: %d", nmethods)
	}
	methods := make([]byte, int(nmethods))
	if _, err := io.ReadFull(reader, methods); err != nil {
		return nil, err
	}

	socksConnection.logWithLevel(logLevelInfo,
		fmt.Sprintf("A negotiation request has been received. VER: %#v NMETHODS: %#v METHOS: %#v", ver, nmethods, methods))

	methodToUse := methodToUseIn(methods)

	if methodToUse == noAcceptable {
		return nil, errNegotiationMethodNotSupported
	}

	return &negotiationRequest{
		ver:             ver,
		nmethods:        nmethods,
		methods:         methods,
		socksConnection: socksConnection,
	}, nil
}
