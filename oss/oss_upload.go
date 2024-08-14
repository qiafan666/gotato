package oss

import (
	"context"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io"
	"io/ioutil"
	"time"
)

type Client interface {
	HealthCheck(ctx context.Context) (health bool, err error)
	UploadAndSignUrl(ctx context.Context, fileReader io.Reader, objectName string, expiredInSec int64) (string, error)
	DeleteByObjectName(ctx context.Context, objectName string) error
	UploadByReader(ctx context.Context, fileReader io.Reader, fileName string) (err error)
	DownloadFile(ctx context.Context, fileName string) (data []byte, err error)
	IsFileExist(ctx context.Context, fileName string) (isExist bool, err error)
	GetFileURL(ctx context.Context, fileName string, expireTime time.Duration) (url string, err error)
	MemoryParameter(ctx context.Context) (memoryParameters MemoryParameter, err error)
	GetClient(ctx context.Context) (client *oss.Client, err error)
	GetBucket(ctx context.Context) (bucket *oss.Bucket, err error)
}

type ClientImp struct {
	ossBucket       string
	accessKeyID     string
	accessKeySecret string
	ossEndPoint     string
}

func (slf *ClientImp) HealthCheck(ctx context.Context) (health bool, err error) {
	_, err = oss.New(slf.ossEndPoint, slf.accessKeyID, slf.accessKeySecret)
	if err != nil {
		return false, err
	}
	health = true
	return
}
func (slf *ClientImp) GetFileURL(ctx context.Context, fileName string, expireTime time.Duration) (url string, err error) {
	client, err := oss.New(slf.ossEndPoint, slf.accessKeyID, slf.accessKeySecret)
	if err != nil {
		return "", err
	}
	bucket, err := client.Bucket(slf.ossBucket)
	if err != nil {
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
		return false, err
	}
	bucket, err := client.Bucket(slf.ossBucket)
	if err != nil {
		return false, err
	}
	return bucket.IsObjectExist(fileName)
}

func (slf *ClientImp) UploadAndSignUrl(ctx context.Context, fileReader io.Reader, objectName string, expiredInSec int64) (string, error) {
	// 创建OSSClient实例。
	client, err := oss.New(slf.ossEndPoint, slf.accessKeyID, slf.accessKeySecret)
	if err != nil {
		return "", err
	}
	bucket, err := client.Bucket(slf.ossBucket)
	if err != nil {
		return "", err
	}
	err = bucket.PutObject(objectName, fileReader)
	if err != nil {
		return "", err
	}
	//oss.Process("image/format,png")
	signedURL, err := bucket.SignURL(objectName, oss.HTTPGet, expiredInSec)
	if err != nil {
		err = bucket.DeleteObject(objectName)
		if err != nil {
			return "", err
		}
		return "", err
	}
	return signedURL, nil
}

func (slf *ClientImp) DeleteByObjectName(ctx context.Context, objectName string) error {
	client, err := oss.New(slf.ossEndPoint, slf.accessKeyID, slf.accessKeySecret)
	if err != nil {
		return err
	}
	bucket, err := client.Bucket(slf.ossBucket)
	if err != nil {
		return err
	}
	err = bucket.DeleteObject(objectName)
	if err != nil {
		return err
	}
	return nil
}

func (slf *ClientImp) UploadByReader(ctx context.Context, fileReader io.Reader, fileName string) (err error) {

	client, err := oss.New(slf.ossEndPoint, slf.accessKeyID, slf.accessKeySecret)
	if err != nil {
		return err
	}

	bucket, err := client.Bucket(slf.ossBucket)
	if err != nil {
		return err
	}

	err = bucket.PutObject(fileName, fileReader)
	if err != nil {
		return err
	}
	return nil
}

func (slf *ClientImp) DownloadFile(ctx context.Context, fileName string) (data []byte, err error) {

	client, err := oss.New(slf.ossEndPoint, slf.accessKeyID, slf.accessKeySecret)
	if err != nil {
		return nil, err
	}

	bucket, err := client.Bucket(slf.ossBucket)
	if err != nil {
		return nil, err
	}

	body, err := bucket.GetObject(fileName)
	if err != nil {
		return nil, err
	}
	// 数据读取完成后，获取的流必须关闭，否则会造成连接泄漏，导致请求无连接可用，程序无法正常工作。
	defer body.Close()

	data, err = ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return data, nil
}
func (slf *ClientImp) MemoryParameter(ctx context.Context) (memoryParameters MemoryParameter, err error) {

	client, err := oss.New(slf.ossEndPoint, slf.accessKeyID, slf.accessKeySecret)
	if err != nil {
		return MemoryParameter{}, err
	}

	stat, err := client.GetBucketStat(slf.ossBucket)
	if err != nil {
		return MemoryParameter{}, err
	}
	memoryParameters.Storage = stat.Storage
	memoryParameters.ObjectCount = stat.ObjectCount
	memoryParameters.MultipartUploadCount = stat.MultipartUploadCount
	memoryParameters.LiveChannelCount = stat.LiveChannelCount
	memoryParameters.LastModifiedTime = stat.LastModifiedTime
	memoryParameters.StandardStorage = stat.StandardStorage
	memoryParameters.StandardObjectCount = stat.StandardObjectCount
	memoryParameters.InfrequentAccessStorage = stat.InfrequentAccessStorage
	memoryParameters.InfrequentAccessRealStorage = stat.InfrequentAccessRealStorage
	memoryParameters.InfrequentAccessObjectCount = stat.InfrequentAccessObjectCount
	memoryParameters.ArchiveStorage = stat.ArchiveStorage
	memoryParameters.ArchiveRealStorage = stat.ArchiveRealStorage
	memoryParameters.ArchiveObjectCount = stat.ArchiveObjectCount
	memoryParameters.ColdArchiveStorage = stat.ColdArchiveStorage
	memoryParameters.ColdArchiveRealStorage = stat.ColdArchiveRealStorage
	memoryParameters.ColdArchiveObjectCount = stat.ColdArchiveObjectCount
	return
}

func (slf *ClientImp) GetClient(ctx context.Context) (client *oss.Client, err error) {

	client, err = oss.New(slf.ossEndPoint, slf.accessKeyID, slf.accessKeySecret)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (slf *ClientImp) GetBucket(ctx context.Context) (bucket *oss.Bucket, err error) {

	client, err := oss.New(slf.ossEndPoint, slf.accessKeyID, slf.accessKeySecret)
	if err != nil {
		return nil, err
	}
	bucket, err = client.Bucket(slf.ossBucket)
	if err != nil {
		return nil, err
	}
	return bucket, nil
}
