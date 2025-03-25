package smtp

import (
	"github.com/qiafan666/gotato/gconfig"
	"testing"
)

func TestSMTP(t *testing.T) {
	config := gconfig.SmtpConfig{}
	err := Sendmail(config,
		"domain1.com,domain2.com", "Gotato 账号验证", example, "tessssssssssssssssssssssssssssssst")
	if err != nil {
		return
	}
}
