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

// producerAck 生产者确认模式,默认wait_for_all
// no_response
//wait_for_local
//wait_for_all

// Config 配置项
type Config struct {
	Username     string    `yaml:"username"`     // 用户名
	Password     string    `yaml:"password"`     // 密码
	ProducerAck  string    `yaml:"producerAck"`  // 生产者确认模式
	CompressType string    `yaml:"compressType"` // 压缩类型
	Addr         []string  `yaml:"addr"`         // kafka地址
	TLS          TLSConfig `yaml:"tls"`          // TLS配置
}
