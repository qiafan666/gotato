package oss

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/qiafan666/gotato/config"
)

type Oss struct {
	bucket *oss.Bucket
	client *oss.Client
	name   string //oss  name
}

func (slf *Oss) Bucket() *oss.Bucket {
	return slf.bucket
}
func (slf *Oss) Client() *oss.Client {
	return slf.client
}

func (slf *Oss) Name() string {
	return slf.name
}

func (slf *Oss) StartOss(config config.OssConfig) error {

	slf.name = config.Name

	client, err := oss.New(config.OssEndPoint, config.AccessKeyID, config.AccessKeySecret)
	if err != nil {
		panic(fmt.Sprintf("oss.New Error:%s", err))
	}
	slf.client = client
	bucket, err := client.Bucket(config.OssBucket)
	if err != nil {
		panic(fmt.Sprintf("client.Bucket Error:%s", err))
	}
	slf.bucket = bucket

	return nil
}
