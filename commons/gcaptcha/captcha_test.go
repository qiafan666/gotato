package gcaptcha

import (
	redisV8 "github.com/go-redis/redis/v8"
	"github.com/mojocn/base64Captcha"
	"testing"
)

func TestCaptcha(t *testing.T) {
	rdb := redisV8.NewClient(&redisV8.Options{
		Addr:     "127.0.0.1:16379",
		Username: "",
		Password: "test",
	})
	store := &RedisStore{
		Rdb: rdb,
	}

	// audioCaptcha := base64Captcha.NewCaptcha(captcha.AudioDrive, store)
	// t.Log(audioCaptcha.Generate())
	// t.Log(store.Verify("0xc7Dvuya7O1eE1Ha74A", "0556", true))

	// stringCaptcha := base64Captcha.NewCaptcha(captcha.StringDrive, store)
	// t.Log(stringCaptcha.Generate())
	// t.Log(store.Verify("CrCbcFVKvXXREZqT3sVu", "UzDAbN", true))

	// chineseCaptcha := base64Captcha.NewCaptcha(captcha.ChineseDrive, store)
	// t.Log(chineseCaptcha.Generate())
	// t.Log(store.Verify("gq78U68yhuM0nCoMBzgq", "不想要你在", true))

	// matchCaptcha := base64Captcha.NewCaptcha(captcha.MathDrive, store)
	// t.Log(matchCaptcha.Generate())
	// t.Log(store.Verify("1XH9XkovkgkTyEFH3qBq", "-12", true))

	/*type adminServer struct {
	    //server 端数据类型应该是这样
		CaptchaStore base64Captcha.Store
	}

	实例化adminServer

	server := adminServer{
		CaptchaStore: &RedisStore{
		Rdb: rdb,
	},
	*/
	digitCaptcha := base64Captcha.NewCaptcha(DigitDrive, store)
	id, b64s, _, err := digitCaptcha.Generate()
	if err != nil {
		t.Error(err)
	}
	t.Log(id, b64s)
	t.Log(store.Verify("Y7b24J7uTtynKjAmoyJn", "081091", true))

}
