package mysocks

const (
	NoAuthRequired byte = 0x00
	GSSAPI         byte = 0x01
	UsernamePasswd byte = 0x02
	IANAAssigned   byte = 0x03
	Reserved       byte = 0x80
	NoAcceptable   byte = 0xFF
)

var SupportedMethod = NoAuthRequired
