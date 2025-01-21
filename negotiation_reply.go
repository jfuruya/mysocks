package mysocks

import (
	"io"
	"log"
)

type NegotiationReply struct {
	Ver    byte
	Method byte
}

func NewNegotiationReply(method byte) *NegotiationReply {
	return &NegotiationReply{
		Ver:    Ver,
		Method: method,
	}
}

func (r *NegotiationReply) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte{r.Ver, r.Method})
	if err != nil {
		return 0, err
	}

	log.Printf("Negotiation reply sent. VER: %#v METHOD: %#v\n", r.Ver, r.Method)

	return int64(n), nil
}
