package gcaptcha

import (
	"context"
	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
	"time"
)

var AudioDrive = &base64Captcha.DriverAudio{
	Length:   4,
	Language: "zh",
}

var StringDrive = &base64Captcha.DriverString{
	Height:          80,
	Width:           240,
	NoiseCount:      0,
	ShowLineOptions: 0,
	Length:          6,
	Source:          base64Captcha.TxtAlphabet,
	BgColor:         nil,
	Fonts:           nil,
}

var ChineseDrive = &base64Captcha.DriverChinese{
	Height:          80,
	Width:           240,
	NoiseCount:      0,
	ShowLineOptions: 0,
	Length:          2,
	Source:          "设想,你在,处理,消费者,的音,频输,出音,频可,能无,论什,么都,没有,任何,输出,或者,它可,能是,单声道,立体声,或是,环绕立,体声的,,不想要,的值",
	BgColor:         nil,
	Fonts:           []string{"wqy-microhei.ttc"},
}

var MathDrive = &base64Captcha.DriverMath{
	Height:          80,
	Width:           240,
	NoiseCount:      0,
	ShowLineOptions: 0,
	BgColor:         nil,
	Fonts:           nil,
}

var DigitDrive = &base64Captcha.DriverDigit{
	Height:   80,
	Width:    240,
	Length:   6,
	MaxSkew:  0.6,
	DotCount: 8,
}

type RedisStore struct {
	Rdb redis.UniversalClient
}

func (s *RedisStore) Set(id string, value string) error {
	key := "captcha:" + id
	if err := s.Rdb.Set(context.Background(), key, value, time.Minute*10).Err(); err != nil {
		return err
	}
	return nil
}

func (s *RedisStore) Get(id string, clear bool) string {
	key := "captcha:" + id
	val, err := s.Rdb.Get(context.Background(), key).Result()
	if err != nil {
		return ""
	}
	if clear {
		if err := s.Rdb.Del(context.Background(), key).Err(); err != nil {
			return ""
		}
	}
	return val
}

func (s *RedisStore) Verify(id string, answer string, clear bool) bool {
	val := s.Get(id, clear)
	return val == answer
}
