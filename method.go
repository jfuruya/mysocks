package mysocks

const (
	noAuthRequired byte = 0x00
	gssAPI         byte = 0x01
	usernamePasswd byte = 0x02
	ianaAssigned   byte = 0x03
	reserved       byte = 0x80
	noAcceptable   byte = 0xFF
)

func methodToUseIn(methods []byte) byte {
	if methodExists(methods, usernamePasswd) {
		return usernamePasswd
	}
	if methodExists(methods, noAuthRequired) {
		return noAuthRequired
	}
	return noAcceptable
}

func methodExists(methods []byte, targetMethod byte) bool {
	for _, method := range methods {
		if method == targetMethod {
			return true
		}
	}
	return false
}

func methodNeedsUserPasswordAuth(method byte) bool {
	return method == usernamePasswd
}
