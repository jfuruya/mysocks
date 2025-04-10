package mysocks

import (
	"fmt"
	"io"
)

type userPasswordAuthRequest struct {
	ver             byte
	uname           []byte
	passwd          []byte
	socksConnection *socksConnection
}

func newUserPasswordAuthRequestFrom(socksConnection *socksConnection) (*userPasswordAuthRequest, error) {
	reader := *socksConnection.clientTCPConn

	verBytes := make([]byte, 1)
	if _, err := io.ReadFull(reader, verBytes); err != nil {
		return nil, err
	}
	ver := verBytes[0]
	if ver != fiexedUserPasswordAuthVer {
		return nil, fmt.Errorf("the value of the VER field in the user password authentication request is invalid: %d", ver)
	}

	ulenBytes := make([]byte, 1)
	if _, err := io.ReadFull(reader, ulenBytes); err != nil {
		return nil, err
	}
	ulen := ulenBytes[0]
	if ulen == 0 {
		return nil, fmt.Errorf("the value of the ULEN field in the user password authentication request is invalid: %d", ulen)
	}
	uname := make([]byte, int(ulen))
	if _, err := io.ReadFull(reader, uname); err != nil {
		return nil, err
	}

	plenBytes := make([]byte, 1)
	if _, err := io.ReadFull(reader, plenBytes); err != nil {
		return nil, err
	}
	plen := plenBytes[0]
	if plen == 0 {
		return nil, fmt.Errorf("the value of the PLEN field in the user password authentication request is invalid: %d", plen)
	}
	passwd := make([]byte, int(plen))
	if _, err := io.ReadFull(reader, passwd); err != nil {
		return nil, err
	}

	socksConnection.logWithLevel(logLevelInfo,
		fmt.Sprintf("A user password authentication request has been received. VER: %#v ULEN: %#v UNAME: %#v PLEN: %#v PASSWD: %#v", ver, ulen, uname, plen, passwd))

	return &userPasswordAuthRequest{
		ver:             ver,
		uname:           uname,
		passwd:          passwd,
		socksConnection: socksConnection,
	}, nil
}

func (userPasswordAuthRequest *userPasswordAuthRequest) usernameAsString() string {
	return string(userPasswordAuthRequest.uname)
}

func (userPasswordAuthRequest *userPasswordAuthRequest) passwordAsString() string {
	return string(userPasswordAuthRequest.passwd)
}
