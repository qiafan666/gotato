package gcaptcha

import (
	"context"
	"github.com/mojocn/base64Captcha"
	"github.com/qiafan666/gotato/commons/gredis"
	"testing"
)

func TestCaptcha(t *testing.T) {
	client, err := gredis.NewRedisClient(context.Background(), &gredis.Config{
		ClusterMode: false,
		Address:     []string{"127.0.0.1:6379"},
		Username:    "",
		Password:    "",
		DB:          0,
		MaxRetry:    3,
		PoolSize:    10,
	})
	if err != nil {
		return
	}
	store := &RedisStore{
		Rdb: client.GetRedis(),
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
