package mysocks

import (
	"fmt"
	"io"
)

type userPasswordAuthReply struct {
	ver             byte
	status          byte
	socksConnection *socksConnection
}

const (
	userPasswordAuthReplyStatusSuccess byte = 0x00
	userPasswordAuthReplyStatusFailure byte = 0x01
)

func newUserPasswordAuthReply(authSuccess bool, socksConnection *socksConnection) *userPasswordAuthReply {
	status := statusFor(authSuccess)

	return &userPasswordAuthReply{
		ver:             fiexedUserPasswordAuthVer,
		status:          status,
		socksConnection: socksConnection,
	}
}

func (userPasswordAuthReply *userPasswordAuthReply) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte{userPasswordAuthReply.ver, userPasswordAuthReply.status})
	if err != nil {
		return 0, err
	}

	userPasswordAuthReply.socksConnection.logWithLevel(logLevelInfo,
		fmt.Sprintf("Uer password authentication reply sent. VER: %#v STATUS: %#v", userPasswordAuthReply.ver, userPasswordAuthReply.status))

	return int64(n), nil
}

func statusFor(authSuccess bool) byte {
	if authSuccess {
		return userPasswordAuthReplyStatusSuccess
	}
	return userPasswordAuthReplyStatusFailure
}
