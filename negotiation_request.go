package mysocks

import (
	"errors"
	"fmt"
	"io"
	"log"
)

var ErrNegotiationMethodNotSupported = errors.New("the method is not supported")

type NegotiationRequest struct {
	ver      byte
	nmethods byte
	methods  []byte
}

func NewNegotiationRequestFrom(reader io.Reader) (*NegotiationRequest, error) {
	verBytes := make([]byte, 1)
	if _, err := io.ReadFull(reader, verBytes); err != nil {
		return nil, err
	}
	ver := verBytes[0]
	if ver != Ver {
		return nil, fmt.Errorf("the value of the VER field in the negotiation request is invalid: %d", ver)
	}
	nmethodsBytes := make([]byte, 1)
	if _, err := io.ReadFull(reader, nmethodsBytes); err != nil {
		return nil, err
	}
	nmethods := nmethodsBytes[0]
	if nmethods == 0 {
		return nil, fmt.Errorf("the value of the NMETHODS field in the negotiation request is invalid. : %d", nmethods)
	}
	methods := make([]byte, int(nmethods))
	if _, err := io.ReadFull(reader, methods); err != nil {
		return nil, err
	}

	log.Printf("A negotiation request has been received. VER: %#v NMETHODS: %#v METHOS: %#v\n", ver, nmethods, methods)

	var methodAggreed bool
	for _, method := range methods {
		if method == byte(SupportedMethod) {
			methodAggreed = true
		}
	}

	if !methodAggreed {
		return nil, ErrNegotiationMethodNotSupported
	}

	return &NegotiationRequest{
		ver:      ver,
		nmethods: nmethods,
		methods:  methods,
	}, nil
}
