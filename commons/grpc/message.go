package grpc

type Message struct {
	Command   Command    `json:"command"`  //业务命令
	PkgType   PkgType    `json:"pkg_type"` //包类型
	ReqId     int64      `json:"req_id"`   //请求ID
	Seq       uint32     `json:"seq"`      //序列号
	Result    uint32     `json:"result"`   //错误码
	Body      []byte     `json:"body"`     //消息体
	Heartbeat *Heartbeat `json:"heartbeat,omitempty"`
}

type Heartbeat struct {
	Timeout uint32 `json:"timeout"` //心跳超时时间 ms
}
