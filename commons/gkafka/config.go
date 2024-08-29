package gkafka

// TLSConfig 配置TLS连接
type TLSConfig struct {
	EnableTLS          bool   `yaml:"enableTLS"`
	CACrt              string `yaml:"caCrt"`
	ClientCrt          string `yaml:"clientCrt"`
	ClientKey          string `yaml:"clientKey"`
	ClientKeyPwd       string `yaml:"clientKeyPwd"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
}

// RequiredAcks 生产者确认模式
// NoResponse doesn't send any response, the TCP ACK is all you get.
//NoResponse RequiredAcks = 0
// WaitForLocal waits for only the local commit to succeed before responding.
//WaitForLocal RequiredAcks = 1
// WaitForAll waits for all in-sync replicas to commit before responding.
// The minimum number of in-sync replicas is configured on the broker via
// the `min.insync.replicas` configuration key.
//WaitForAll RequiredAcks = -1

// Config 配置项
type Config struct {
	Username     string    `yaml:"username"`     // 用户名
	Password     string    `yaml:"password"`     // 密码
	ProducerAck  string    `yaml:"producerAck"`  // 生产者确认模式
	CompressType string    `yaml:"compressType"` // 压缩类型
	Addr         []string  `yaml:"addr"`         // kafka地址
	TLS          TLSConfig `yaml:"tls"`          // TLS配置
}
