package minio

import (
	"context"
	"errors"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/signer"
	"github.com/qiafan666/gotato/commons/gerr"
	"github.com/qiafan666/gotato/commons/gs3"
	"io"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

const (
	unsignedPayload = "UNSIGNED-PAYLOAD"
)

const (
	minPartSize int64 = 1024 * 1024 * 5        // 5MB
	maxPartSize int64 = 1024 * 1024 * 1024 * 5 // 5GB
	maxNumSize  int64 = 10000
)

const (
	maxImageWidth      = 1024
	maxImageHeight     = 1024
	maxImageSize       = 1024 * 1024 * 50
	imageThumbnailPath = "gotato/thumbnail"
)

const successCode = http.StatusOK

var _ gs3.Interface = (*Minio)(nil)

type Config struct {
	Bucket          string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	SignEndpoint    string
	PublicRead      bool
}

func NewMinio(ctx context.Context, cache Cache, conf Config) (*Minio, error) {
	u, err := url.Parse(conf.Endpoint)
	if err != nil {
		return nil, err
	}
	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(conf.AccessKeyID, conf.SecretAccessKey, conf.SessionToken),
		Secure: u.Scheme == "https",
	}
	client, err := minio.New(u.Host, opts)
	if err != nil {
		return nil, err
	}
	m := &Minio{
		conf:   conf,
		bucket: conf.Bucket,
		core:   &minio.Core{Client: client},
		lock:   &sync.Mutex{},
		init:   false,
		cache:  cache,
	}
	if conf.SignEndpoint == "" || conf.SignEndpoint == conf.Endpoint {
		m.opts = opts
		m.sign = m.core.Client
		m.prefix = u.Path
		u.Path = ""
		conf.Endpoint = u.String()
		m.signEndpoint = conf.Endpoint
	} else {
		su, err := url.Parse(conf.SignEndpoint)
		if err != nil {
			return nil, err
		}
		m.opts = &minio.Options{
			Creds:  credentials.NewStaticV4(conf.AccessKeyID, conf.SecretAccessKey, conf.SessionToken),
			Secure: su.Scheme == "https",
		}
		m.sign, err = minio.New(su.Host, m.opts)
		if err != nil {
			return nil, err
		}
		m.prefix = su.Path
		su.Path = ""
		conf.SignEndpoint = su.String()
		m.signEndpoint = conf.SignEndpoint
	}
	if err := m.initMinio(ctx); err != nil {
		return nil, err
	}
	return m, nil
}

type Minio struct {
	conf         Config
	bucket       string
	signEndpoint string
	location     string
	opts         *minio.Options
	core         *minio.Core
	sign         *minio.Client
	lock         sync.Locker
	init         bool
	prefix       string
	cache        Cache
}

func (m *Minio) initMinio(ctx context.Context) error {
	if m.init {
		return nil
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.init {
		return nil
	}
	exists, err := m.core.Client.BucketExists(ctx, m.conf.Bucket)
	if err != nil {
		return fmt.Errorf("check bucket exists error: %w", err)
	}
	if !exists {
		if err = m.core.Client.MakeBucket(ctx, m.conf.Bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("make bucket error: %w", err)
		}
	}
	if m.conf.PublicRead {
		policy := fmt.Sprintf(
			`{"Version": "2012-10-17","Statement": [{"Action": ["gs3:GetObject","gs3:PutObject"],"Effect": "Allow","Principal": {"AWS": ["*"]},"Resource": ["arn:aws:gs3:::%s/*"],"Sid": ""}]}`,
			m.conf.Bucket,
		)
		if err = m.core.Client.SetBucketPolicy(ctx, m.conf.Bucket, policy); err != nil {
			return err
		}
	}
	m.location, err = m.core.Client.GetBucketLocation(ctx, m.conf.Bucket)
	if err != nil {
		return err
	}
	func() {
		if m.conf.SignEndpoint == "" || m.conf.SignEndpoint == m.conf.Endpoint {
			return
		}
		defer func() {
			if r := recover(); r != nil {
				m.sign = m.core.Client
			}
		}()
		blc := reflect.ValueOf(m.sign).Elem().FieldByName("bucketLocCache")
		vblc := reflect.New(reflect.PtrTo(blc.Type()))
		*(*unsafe.Pointer)(vblc.UnsafePointer()) = unsafe.Pointer(blc.UnsafeAddr())
		vblc.Elem().Elem().Interface().(interface{ Set(string, string) }).Set(m.conf.Bucket, m.location)
	}()
	m.init = true
	return nil
}

func (m *Minio) Engine() string {
	return "minio"
}

func (m *Minio) PartLimit() *gs3.PartLimit {
	return &gs3.PartLimit{
		MinPartSize: minPartSize,
		MaxPartSize: maxPartSize,
		MaxNumSize:  maxNumSize,
	}
}

func (m *Minio) InitiateMultipartUpload(ctx context.Context, name string) (*gs3.InitiateMultipartUploadResult, error) {
	if err := m.initMinio(ctx); err != nil {
		return nil, err
	}
	uploadID, err := m.core.NewMultipartUpload(ctx, m.bucket, name, minio.PutObjectOptions{})
	if err != nil {
		return nil, err
	}
	return &gs3.InitiateMultipartUploadResult{
		Bucket:   m.bucket,
		Key:      name,
		UploadID: uploadID,
	}, nil
}

func (m *Minio) CompleteMultipartUpload(ctx context.Context, uploadID string, name string, parts []gs3.Part) (*gs3.CompleteMultipartUploadResult, error) {
	if err := m.initMinio(ctx); err != nil {
		return nil, err
	}
	minioParts := make([]minio.CompletePart, len(parts))
	for i, part := range parts {
		minioParts[i] = minio.CompletePart{
			PartNumber: part.PartNumber,
			ETag:       strings.ToLower(part.ETag),
		}
	}
	upload, err := m.core.CompleteMultipartUpload(ctx, m.bucket, name, uploadID, minioParts, minio.PutObjectOptions{})
	if err != nil {
		return nil, err
	}
	m.delObjectImageInfoKey(ctx, name, upload.Size)
	return &gs3.CompleteMultipartUploadResult{
		Location: upload.Location,
		Bucket:   upload.Bucket,
		Key:      upload.Key,
		ETag:     strings.ToLower(upload.ETag),
	}, nil
}

func (m *Minio) PartSize(ctx context.Context, size int64) (int64, error) {
	if size <= 0 {
		return 0, errors.New("size must be greater than 0")
	}
	if size > maxPartSize*maxNumSize {
		return 0, fmt.Errorf("MINIO size must be less than the maximum allowed limit")
	}
	if size <= minPartSize*maxNumSize {
		return minPartSize, nil
	}
	partSize := size / maxNumSize
	if size%maxNumSize != 0 {
		partSize++
	}
	return partSize, nil
}

func (m *Minio) AuthSign(ctx context.Context, uploadID string, name string, expire time.Duration, partNumbers []int) (*gs3.AuthSignResult, error) {
	if err := m.initMinio(ctx); err != nil {
		return nil, err
	}
	creds, err := m.opts.Creds.Get()
	if err != nil {
		return nil, err
	}
	result := gs3.AuthSignResult{
		URL:   m.signEndpoint + "/" + m.bucket + "/" + name,
		Query: url.Values{"uploadId": {uploadID}},
		Parts: make([]gs3.SignPart, len(partNumbers)),
	}
	for i, partNumber := range partNumbers {
		rawURL := result.URL + "?partNumber=" + strconv.Itoa(partNumber) + "&uploadId=" + uploadID
		request, err := http.NewRequestWithContext(ctx, http.MethodPut, rawURL, nil)
		if err != nil {
			return nil, err
		}
		request.Header.Set("X-Amz-Content-Sha256", unsignedPayload)
		request = signer.SignV4Trailer(*request, creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken, m.location, nil)
		result.Parts[i] = gs3.SignPart{
			PartNumber: partNumber,
			Query:      url.Values{"partNumber": {strconv.Itoa(partNumber)}},
			Header:     request.Header,
		}
	}
	if m.prefix != "" {
		result.URL = m.signEndpoint + m.prefix + "/" + m.bucket + "/" + name
	}
	return &result, nil
}

func (m *Minio) PresignedPutObject(ctx context.Context, name string, expire time.Duration) (string, error) {
	if err := m.initMinio(ctx); err != nil {
		return "", err
	}
	rawURL, err := m.sign.PresignedPutObject(ctx, m.bucket, name, expire)
	if err != nil {
		return "", err
	}
	if m.prefix != "" {
		rawURL.Path = path.Join(m.prefix, rawURL.Path)
	}
	return rawURL.String(), nil
}

func (m *Minio) DeleteObject(ctx context.Context, name string) error {
	if err := m.initMinio(ctx); err != nil {
		return err
	}
	return m.core.Client.RemoveObject(ctx, m.bucket, name, minio.RemoveObjectOptions{})
}

func (m *Minio) StatObject(ctx context.Context, name string) (*gs3.ObjectInfo, error) {
	if err := m.initMinio(ctx); err != nil {
		return nil, err
	}
	info, err := m.core.Client.StatObject(ctx, m.bucket, name, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}
	return &gs3.ObjectInfo{
		ETag:         strings.ToLower(info.ETag),
		Key:          info.Key,
		Size:         info.Size,
		LastModified: info.LastModified,
	}, nil
}

func (m *Minio) CopyObject(ctx context.Context, src string, dst string) (*gs3.CopyObjectInfo, error) {
	if err := m.initMinio(ctx); err != nil {
		return nil, err
	}
	result, err := m.core.Client.CopyObject(ctx, minio.CopyDestOptions{
		Bucket: m.bucket,
		Object: dst,
	}, minio.CopySrcOptions{
		Bucket: m.bucket,
		Object: src,
	})
	if err != nil {
		return nil, err
	}
	return &gs3.CopyObjectInfo{
		Key:  dst,
		ETag: strings.ToLower(result.ETag),
	}, nil
}

func (m *Minio) IsNotFound(err error) bool {
	switch e := gerr.Unwrap(err).(type) {
	case minio.ErrorResponse:
		return e.StatusCode == http.StatusNotFound || e.Code == "NoSuchKey"
	case *minio.ErrorResponse:
		return e.StatusCode == http.StatusNotFound || e.Code == "NoSuchKey"
	default:
		return false
	}
}

func (m *Minio) AbortMultipartUpload(ctx context.Context, uploadID string, name string) error {
	if err := m.initMinio(ctx); err != nil {
		return err
	}
	return m.core.AbortMultipartUpload(ctx, m.bucket, name, uploadID)
}

func (m *Minio) ListUploadedParts(ctx context.Context, uploadID string, name string, partNumberMarker int, maxParts int) (*gs3.ListUploadedPartsResult, error) {
	if err := m.initMinio(ctx); err != nil {
		return nil, err
	}
	result, err := m.core.ListObjectParts(ctx, m.bucket, name, uploadID, partNumberMarker, maxParts)
	if err != nil {
		return nil, err
	}
	res := &gs3.ListUploadedPartsResult{
		Key:                  result.Key,
		UploadID:             result.UploadID,
		MaxParts:             result.MaxParts,
		NextPartNumberMarker: result.NextPartNumberMarker,
		UploadedParts:        make([]gs3.UploadedPart, len(result.ObjectParts)),
	}
	for i, part := range result.ObjectParts {
		res.UploadedParts[i] = gs3.UploadedPart{
			PartNumber:   part.PartNumber,
			LastModified: part.LastModified,
			ETag:         part.ETag,
			Size:         part.Size,
		}
	}
	return res, nil
}

func (m *Minio) PresignedGetObject(ctx context.Context, name string, expire time.Duration, query url.Values) (string, error) {
	if expire <= 0 {
		expire = time.Hour * 24 * 365 * 99 // 99 years
	} else if expire < time.Second {
		expire = time.Second
	}
	var (
		rawURL *url.URL
		err    error
	)
	if m.conf.PublicRead {
		rawURL, err = makeTargetURL(m.sign, m.bucket, name, m.location, false, query)
	} else {
		rawURL, err = m.sign.PresignedGetObject(ctx, m.bucket, name, expire, query)
	}
	if err != nil {
		return "", err
	}
	if m.prefix != "" {
		rawURL.Path = path.Join(m.prefix, rawURL.Path)
	}
	return rawURL.String(), nil
}

func (m *Minio) AccessURL(ctx context.Context, name string, expire time.Duration, opt *gs3.AccessURLOption) (string, error) {
	if err := m.initMinio(ctx); err != nil {
		return "", err
	}
	reqParams := make(url.Values)
	if opt != nil {
		if opt.ContentType != "" {
			reqParams.Set("response-content-type", opt.ContentType)
		}
		if opt.Filename != "" {
			reqParams.Set("response-content-disposition", `attachment; filename=`+strconv.Quote(opt.Filename))
		}
	}
	if opt.Image == nil || (opt.Image.Width < 0 && opt.Image.Height < 0 && opt.Image.Format == "") || (opt.Image.Width > maxImageWidth || opt.Image.Height > maxImageHeight) {
		return m.PresignedGetObject(ctx, name, expire, reqParams)
	}
	return m.getImageThumbnailURL(ctx, name, expire, (opt.Image))
}

func (m *Minio) getObjectData(ctx context.Context, name string, limit int64) ([]byte, error) {
	object, err := m.core.Client.GetObject(ctx, m.bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer object.Close()
	if limit < 0 {
		return io.ReadAll(object)
	}
	return io.ReadAll(io.LimitReader(object, limit))
}

func (m *Minio) FormData(ctx context.Context, name string, size int64, contentType string, duration time.Duration) (*gs3.FormData, error) {
	if err := m.initMinio(ctx); err != nil {
		return nil, err
	}
	policy := minio.NewPostPolicy()
	if err := policy.SetKey(name); err != nil {
		return nil, err
	}
	expires := time.Now().Add(duration)
	if err := policy.SetExpires(expires); err != nil {
		return nil, err
	}
	if size > 0 {
		if err := policy.SetContentLengthRange(0, size); err != nil {
			return nil, err
		}
	}
	if err := policy.SetSuccessStatusAction(strconv.Itoa(successCode)); err != nil {
		return nil, err
	}
	if contentType != "" {
		if err := policy.SetContentType(contentType); err != nil {
			return nil, err
		}
	}
	if err := policy.SetBucket(m.bucket); err != nil {
		return nil, err
	}
	u, fd, err := m.core.PresignedPostPolicy(ctx, policy)
	if err != nil {
		return nil, err
	}
	sign, err := url.Parse(m.signEndpoint)
	if err != nil {
		return nil, err
	}
	u.Scheme = sign.Scheme
	u.Host = sign.Host
	return &gs3.FormData{
		URL:          u.String(),
		File:         "file",
		Header:       nil,
		FormData:     fd,
		Expires:      expires,
		SuccessCodes: []int{successCode},
	}, nil
}
