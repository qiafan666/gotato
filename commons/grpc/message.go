package grpc

type Message struct {
	Command   Command
	PkgType   PkgType
	ReqId     uint64
	Sequence  uint32
	Result    uint32
	Body      []byte
	Heartbeat *Heartbeat
}

type Heartbeat struct {
	Timeout uint32
}
