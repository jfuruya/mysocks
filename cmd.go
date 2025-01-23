package mysocks

const (
	cmdConnect   byte = 0x01
	cmdBind      byte = 0x02
	cmdAssociate byte = 0x03
)

var supportedCmd = cmdConnect
