package grpc

type PkgType uint16

const (
	PkgTypeRequest PkgType = 0 + iota
	PkgTypeReply
	PkgTypePush
)
