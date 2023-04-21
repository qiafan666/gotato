package oss

import (
	"context"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	slog "github.com/qiafan666/quickweb/commons/log"
	"io"
	"io/ioutil"
	"time"
)

type Client interface {
	UploadAndSignUrl(ctx context.Context, fileReader io.Reader, objectName string, expiredInSec int64) (string, error)
	DeleteByObjectName(ctx context.Context, objectName string)
	UploadByReader(ctx context.Context, fileReader io.Reader, fileName string) (err error)
	DownloadFile(ctx context.Context, fileName string) (data []byte, err error)
	IsFileExist(ctx context.Context, fileName string) (isExist bool, err error)
	GetFileURL(ctx context.Context, fileName string, expireTime time.Duration) (url string, err error)
}

type ClientImp struct {
	ossBucket       string
	accessKeyID     string
	accessKeySecret string
	ossEndPoint     string
}

func (slf *ClientImp) GetFileURL(ctx context.Context, fileName string, expireTime time.Duration) (url string, err error) {
	client, err := oss.New(slf.ossEndPoint, slf.accessKeyID, slf.accessKeySecret)
	if err != nil {
		slog.Slog.ErrorF(ctx, "ClientImp IsFileExist Error:%s", err)
		return "", err
	}
	bucket, err := client.Bucket(slf.ossBucket)
	if err != nil {
		slog.Slog.ErrorF(ctx, "ClientImp IsFileExist  Error:%s", err)
		return "", err
	}
	url, err = bucket.SignURL(fileName, oss.HTTPGet, int64(expireTime))
	if err != nil {
		return "", err
	}
	return

}

func ClientInstance(ossBucket, accessKeyID, accessKeySecret, ossEndPoint string) Client {
	return &ClientImp{
		ossBucket:       ossBucket,
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
		ossEndPoint:     ossEndPoint,
	}
}

func (slf *ClientImp) IsFileExist(ctx context.Context, fileName string) (isExist bool, err error) {
	client, err := oss.New(slf.ossEndPoint, slf.accessKeyID, slf.accessKeySecret)
	if err != nil {
		slog.Slog.ErrorF(ctx, "ClientImp IsFileExist Error:%s", err)
		return false, err
	}
	bucket, err := client.Bucket(slf.ossBucket)
	if err != nil {
		slog.Slog.ErrorF(ctx, "ClientImp IsFileExist  Error:%s", err)
		return false, err
	}
	return bucket.IsObjectExist(fileName)
}

func (slf *ClientImp) UploadAndSignUrl(ctx context.Context, fileReader io.Reader, objectName string, expiredInSec int64) (string, error) {
	// 创建OSSClient实例。
	client, err := oss.New(slf.ossEndPoint, slf.accessKeyID, slf.accessKeySecret)
	if err != nil {
		slog.Slog.ErrorF(ctx, "Error:%s", err)
		return "", err
	}
	bucket, err := client.Bucket(slf.ossBucket)
	if err != nil {
		slog.Slog.ErrorF(ctx, "Error:%s", err)
		return "", err
	}
	err = bucket.PutObject(objectName, fileReader)
	if err != nil {
		slog.Slog.ErrorF(ctx, "Error:%s", err)
		return "", err
	}
	//oss.Process("image/format,png")
	signedURL, err := bucket.SignURL(objectName, oss.HTTPGet, expiredInSec)
	if err != nil {
		bucket.DeleteObject(objectName)
		return "", err
	}
	return signedURL, nil
}

func (slf *ClientImp) DeleteByObjectName(ctx context.Context, objectName string) {
	client, err := oss.New(slf.ossEndPoint, slf.accessKeyID, slf.accessKeySecret)
	if err != nil {
		slog.Slog.ErrorF(ctx, "Error:%s", err)
		return
	}
	bucket, err := client.Bucket(slf.ossBucket)
	if err != nil {
		slog.Slog.ErrorF(ctx, "Error:%s", err)
		return
	}
	err = bucket.DeleteObject(objectName)
	if err != nil {
		slog.Slog.ErrorF(ctx, "Error:%s", err)
	}
}

func (slf *ClientImp) UploadByReader(ctx context.Context, fileReader io.Reader, fileName string) (err error) {

	client, err := oss.New(slf.ossEndPoint, slf.accessKeyID, slf.accessKeySecret)
	if err != nil {
		slog.Slog.ErrorF(ctx, "Error:%s", err)
		return
	}

	bucket, err := client.Bucket(slf.ossBucket)
	if err != nil {
		slog.Slog.ErrorF(ctx, "Error:%s", err)
		return
	}

	err = bucket.PutObject(fileName, fileReader)
	if err != nil {
		slog.Slog.ErrorF(ctx, "Error:%s", err)
		return
	}
	return
}

func (slf *ClientImp) DownloadFile(ctx context.Context, fileName string) (data []byte, err error) {

	client, err := oss.New(slf.ossEndPoint, slf.accessKeyID, slf.accessKeySecret)
	if err != nil {
		slog.Slog.ErrorF(ctx, "Error:%s", err)
		return
	}

	bucket, err := client.Bucket(slf.ossBucket)
	if err != nil {
		slog.Slog.ErrorF(ctx, "Error:%s", err)
		return
	}

	body, err := bucket.GetObject(fileName)
	if err != nil {
		slog.Slog.ErrorF(ctx, "Error:%s", err)
		return
	}
	// 数据读取完成后，获取的流必须关闭，否则会造成连接泄漏，导致请求无连接可用，程序无法正常工作。
	defer body.Close()

	data, err = ioutil.ReadAll(body)
	if err != nil {
		slog.Slog.ErrorF(ctx, "Error:%s", err)
		return
	}
	return
}
