package mysocks

const (
	CmdConnect   byte = 0x01
	CmdBind      byte = 0x02
	CmdAssociate byte = 0x03
)

var SupportedCmd = CmdConnect
