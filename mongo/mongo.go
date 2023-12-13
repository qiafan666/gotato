package mongo

import (
	"errors"
	"fmt"
	"github.com/qiafan666/gotato/config"
	"strings"
)

type Mongo struct {
	Url  string
	DB   string
	c    *DialContext
	name string
}

func (slf *Mongo) Name() string {
	return slf.name
}

func (slf *Mongo) StartMongo(config config.MongoConfig) (err error) {
	if slf.c != nil {
		return errors.New("mongo already opened")
	}
	slf.name = config.Name
	slf.Url = config.Url

	slf.DB = config.Url[strings.LastIndex(config.Url, "/")+1:]
	if len(slf.DB) <= 0 {
		panic(errors.New("url format error"))
	}

	if len(config.Tls) > 0 {
		slf.c, err = dialTLS(config.Url, 10, config.Tls)
	} else {
		slf.c, err = dial(config.Url, 10)
	}
	if err != nil {
		panic(fmt.Sprintf("mongo connetc error %s", err.Error()))
	}

	return nil
}
