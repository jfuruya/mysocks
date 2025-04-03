package mysocks

import (
	"fmt"
	"io"
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

func (negotiationReply *negotiationReply) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte{negotiationReply.ver, negotiationReply.method})
	if err != nil {
		return 0, err
	}

	negotiationReply.socksConnection.logWithLevel(logLevelInfo,
		fmt.Sprintf("Negotiation reply sent. VER: %#v METHOD: %#v", negotiationReply.ver, negotiationReply.method))

	return int64(n), nil
}
