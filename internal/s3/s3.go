package s3

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
)

var (
	ErrNilContext      = errors.New("nil context.Context")
	ErrNilS3           = errors.New("nil S3")
	ErrEmptyBucketName = errors.New("empty BucketName")
	ErrEmptyKey        = errors.New("empty key")
	ErrEmptyPrefix     = errors.New("empty prefix")
	ErrEmptyFilename   = errors.New("empty filename")
	ErrInvalidTTL      = errors.New("invalid ttl")
	ErrObjectNotFound  = errors.New("object not found")
	ErrNilWriter       = errors.New("nil Open")
)

type S3 struct {
	srv        *awss3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	bucketName string
	region     string
}

var DefaultDownloaderOption = func(uploader *s3manager.Downloader) {
	uploader.PartSize = 10 * 1024 * 1024
}

func New(c Config) (*S3, error) {
	sess, err := c.createSession()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create session")
	}

	conf := aws.NewConfig().WithRegion(c.Region).WithEndpoint(c.Endpoint).WithDisableSSL(c.DisableSSL).WithS3ForcePathStyle(c.ForcePathStyle)
	s := awss3.New(sess, conf)
	if _, err := s.HeadBucket(&awss3.HeadBucketInput{Bucket: aws.String(c.Bucket)}); err != nil {
		return nil, errors.Wrap(err, "failed to check bucket exists.")
	}

	return &S3{
		srv:        s,
		uploader:   s3manager.NewUploaderWithClient(s, func(uploader *s3manager.Uploader) { uploader.PartSize = 10 * 1024 * 1024 }),
		downloader: s3manager.NewDownloaderWithClient(s, DefaultDownloaderOption),
		bucketName: c.Bucket,
		region:     c.Region,
	}, nil
}

func (s *S3) S3() *awss3.S3 {
	return s.srv
}

func (s *S3) Downloader() *s3manager.Downloader {
	return s.downloader
}

func (s *S3) Uploader() *s3manager.Uploader {
	return s.uploader
}

func (s *S3) BucketName() string {
	return s.bucketName
}

func (s *S3) Region() string {
	return s.region
}

func (s *S3) IsExist(c context.Context, key string) error {
	switch {
	case c == nil:
		return ErrNilContext
	case len(key) == 0:
		return ErrEmptyKey
	}

	headInput := &awss3.HeadObjectInput{
		Bucket: aws.String(s.BucketName()),
		Key:    aws.String(key),
	}
	if _, err := s.srv.HeadObjectWithContext(c, headInput); err != nil {
		if err, ok := err.(awserr.RequestFailure); ok {
			switch err.StatusCode() {
			case http.StatusNotFound:
				return ErrObjectNotFound
			default:
				return errors.Wrap(err, "request failed: head object")
			}
		} else {
			return errors.Wrap(err, "head object")
		}
	}

	return nil
}

type UploadOption func(input *s3manager.UploadInput)

func SetTags(tags ...Tag) UploadOption {
	return func(input *s3manager.UploadInput) {
		q := make(url.Values)
		for _, t := range tags {
			q.Add(t.Key, t.Value)
		}
		input.Tagging = aws.String(q.Encode())
	}
}

func (s *S3) Upload(c context.Context, key string, file io.Reader, opts ...UploadOption) error {
	switch {
	case c == nil:
		return ErrNilContext
	case len(key) == 0:
		return ErrEmptyKey
	}

	return s.upload(c, key, file, opts...)
}

func (s *S3) upload(c context.Context, key string, body io.Reader, opts ...UploadOption) error {
	switch {
	case c == nil:
		return ErrNilContext
	case len(key) == 0:
		return ErrEmptyKey
	case body == nil:
		return errors.New("nil Body")
	}

	input := &s3manager.UploadInput{
		Body:   body,
		Bucket: aws.String(s.BucketName()),
		Key:    aws.String(key),
		ACL:    aws.String(awss3.BucketCannedACLPublicRead),
	}
	for _, opt := range opts {
		opt(input)
	}

	if _, err := s.uploader.UploadWithContext(c, input); err != nil {
		return errors.Wrap(err, "put object")
	}

	return nil
}

func (s *S3) Download(c context.Context, writer io.WriterAt, key string) (int64, error) {
	switch {
	case c == nil:
		return 0, ErrNilContext
	case writer == nil:
		return 0, ErrNilWriter
	case len(key) == 0:
		return 0, ErrEmptyKey
	}

	n, err := s.downloader.DownloadWithContext(c, writer, &awss3.GetObjectInput{Bucket: aws.String(s.BucketName()), Key: aws.String(key)})
	if err != nil {
		return 0, errors.Wrap(err, "download object")
	}

	return n, nil
}