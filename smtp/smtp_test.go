package smtp

import "testing"

func TestSMTP(t *testing.T) {
	err := Sendmail("smtp", "domain1.com,domain2.com", "Gotato 账号验证", example, "tessssssssssssssssssssssssssssssst")
	if err != nil {
		return
	}
}
