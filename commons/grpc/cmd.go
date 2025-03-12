package grpc

type Command uint32

const (
	CmdHeartbeat Command = 0
	CmdTestLogic Command = 1
)
