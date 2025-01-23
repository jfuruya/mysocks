package mysocks

import (
	"io"
	"log"
)

type negotiationReply struct {
	ver    byte
	method byte
}

func newNegotiationReply(method byte) *negotiationReply {
	return &negotiationReply{
		ver:    fiexedVer,
		method: method,
	}
}

func (r *negotiationReply) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte{r.ver, r.method})
	if err != nil {
		return 0, err
	}

	log.Printf("Negotiation reply sent. VER: %#v METHOD: %#v\n", r.ver, r.method)

	return int64(n), nil
}
