package mysocks

import (
	"io"
	"log"
)

type negotiationReply struct {
	ver             byte
	method          byte
	socksConnection *socksConnection
}

func newNegotiationReply(method byte, socksConnection *socksConnection) *negotiationReply {
	return &negotiationReply{
		ver:             fiexedVer,
		method:          method,
		socksConnection: socksConnection,
	}
}

func (r *negotiationReply) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte{r.ver, r.method})
	if err != nil {
		return 0, err
	}

	log.Printf("Negotiation reply sent. VER: %#v METHOD: %#v %v\n", r.ver, r.method, r.socksConnection)

	return int64(n), nil
}
