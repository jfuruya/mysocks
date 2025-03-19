package mysocks

const (
	cmdConnect   byte = 0x01
	cmdBind      byte = 0x02
	cmdAssociate byte = 0x03
)

func supportedCmd(cmd byte) bool {
	return cmd == cmdConnect || cmd == cmdAssociate
}
