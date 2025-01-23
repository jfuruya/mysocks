package mysocks

const (
	noAuthRequired byte = 0x00
	gssAPI         byte = 0x01
	usernamePasswd byte = 0x02
	ianaAssigned   byte = 0x03
	reserved       byte = 0x80
	noAcceptable   byte = 0xFF
)

var supportedMethod = noAuthRequired
