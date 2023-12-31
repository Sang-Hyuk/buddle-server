package service

import (
	"buddle-server/internal/s3"
	"buddle-server/model"
	"buddle-server/repository"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"io"
	"os"
	"time"
)

type AfterService interface {
	Create(c context.Context, as *model.AfterService, file1 io.Reader, file2 io.Reader, file3 io.Reader, file4 io.Reader, file5 io.Reader) (*model.Response, error)
	FindAfterServiceInfo(c context.Context, req model.AfterServiceRequest) (*model.Response, error)
	FindAfterServiceManagerInfo(c context.Context, req model.AfterServiceRequest) (*model.Response, error)
	DownloadFile(c context.Context, afterServiceSeq, fileIdx int64, file *os.File) (*model.AfterService, error)
}

type afterService struct {
	repo       repository.Repository
	fileBucket *s3.S3
}

func NewAfterService(repo repository.Repository, fileBucket *s3.S3) (AfterService, error) {
	if repo == nil {
		return nil, errors.New("repository is nil")
	}
	return &afterService{repo: repo, fileBucket: fileBucket}, nil
}

func (s afterService) Create(c context.Context, as *model.AfterService, file1 io.Reader, file2 io.Reader, file3 io.Reader, file4 io.Reader, file5 io.Reader) (*model.Response, error) {
	switch {
	case c == nil:
		return model.SimpleFail(), errors.New("nil context")
	case as == nil:
		return model.SimpleFail(), errors.New("nil request params")
	case s.fileBucket == nil:
		return model.SimpleFail(), errors.New("s3 file bucket is nil")
	}

	if file1 != nil {
		// s3 업로드
		s3location := fmt.Sprintf("after-service/%s/%s/file1", time.Now().Format("2006-01-02"), as.Phone)
		if err := s.fileBucket.Upload(c, s3location, file1); err != nil {
			logrus.Errorf("failed to upload s3 object: objectKey=%s, bucketName=%s err:%+v", s3location, s.fileBucket.BucketName(), err)
		}
		as.File1S3Location = s3location
	}

	if file2 != nil {
		// s3 업로드
		s3location := fmt.Sprintf("after-service/%s/%s/file2", time.Now().Format("2006-01-02"), as.Phone)
		if err := s.fileBucket.Upload(c, s3location, file2); err != nil {
			logrus.Errorf("failed to upload s3 object: objectKey=%s, bucketName=%s err:%+v", s3location, s.fileBucket.BucketName(), err)
		}
		as.File2S3Location = s3location
	}

	if file3 != nil {
		// s3 업로드
		s3location := fmt.Sprintf("after-service/%s/%s/file3", time.Now().Format("2006-01-02"), as.Phone)
		if err := s.fileBucket.Upload(c, s3location, file3); err != nil {
			logrus.Errorf("failed to upload s3 object: objectKey=%s, bucketName=%s err:%+v", s3location, s.fileBucket.BucketName(), err)
		}
		as.File3S3Location = s3location
	}

	if file4 != nil {
		// s3 업로드
		s3location := fmt.Sprintf("after-service/%s/%s/file4", time.Now().Format("2006-01-02"), as.Phone)
		if err := s.fileBucket.Upload(c, s3location, file4); err != nil {
			logrus.Errorf("failed to upload s3 object: objectKey=%s, bucketName=%s err:%+v", s3location, s.fileBucket.BucketName(), err)
		}
		as.File4S3Location = s3location
	}

	if file5 != nil {
		// s3 업로드
		s3location := fmt.Sprintf("after-service/%s/%s/file5", time.Now().Format("2006-01-02"), as.Phone)
		if err := s.fileBucket.Upload(c, s3location, file5); err != nil {
			logrus.Errorf("failed to upload s3 object: objectKey=%s, bucketName=%s err:%+v", s3location, s.fileBucket.BucketName(), err)
		}
		as.File5S3Location = s3location
	}

	if err := s.repo.AfterService().Create(c, as); err != nil {
		return nil, errors.Wrap(err, "failed to create product regist")
	}

	return model.SimpleSuccess(), nil
}

func (s afterService) DownloadFile(c context.Context, afterServiceSeq, fileIdx int64, file *os.File) (*model.AfterService, error) {
	switch {
	case c == nil:
		return nil, errors.New("nil context")
	case afterServiceSeq == 0:
		return nil, errors.New("invalid sequence")
	case file == nil:
		return nil, errors.New("file is nil")
	case s.fileBucket == nil:
		return nil, errors.New("s3 file bucket is nil")
	}

	asInfo, err := s.repo.AfterService().GetAfterServiceBySeq(c, afterServiceSeq)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get after service info by seq(%d)", afterServiceSeq)
	}

	var s3Location string
	switch fileIdx {
	case 1:
		s3Location = asInfo.File1S3Location
	case 2:
		s3Location = asInfo.File2S3Location
	case 3:
		s3Location = asInfo.File3S3Location
	case 4:
		s3Location = asInfo.File4S3Location
	case 5:
		s3Location = asInfo.File5S3Location
	default:
		return nil, errors.Wrap(err, "failed to download file")
	}

	if s3Location == "" {
		return nil, nil // 다운로드 받을 파일이 없음.
	}

	if _, err := s.fileBucket.Download(c, file, s3Location); err != nil {
		return nil, errors.Wrapf(err, "failed to download file [ s3 location : %+v ]", s3Location)
	}

	return asInfo, nil
}

func (s afterService) FindAfterServiceInfo(c context.Context, req model.AfterServiceRequest) (*model.Response, error) {
	if c == nil {
		return nil, errors.New("nil context")
	}

	data, err := s.repo.AfterService().FindAfterServiceInfo(c, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.Response{
				Success: false,
				Message: "일치하는 데이터가 없습니다.",
			}, nil
		}
		return nil, errors.Wrap(err, "failed to get after service info")
	}

	return &model.Response{
		Success: true,
		Message: "성공하였습니다.",
		Data:    data,
	}, nil
}

func (s afterService) FindAfterServiceManagerInfo(c context.Context, req model.AfterServiceRequest) (*model.Response, error) {
	if c == nil {
		return nil, errors.New("nil context")
	}

	data, err := s.repo.AfterService().FindAfterServiceManagerInfo(c, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.Response{
				Success: false,
				Message: "일치하는 데이터가 없습니다.",
			}, nil
		}
		return nil, errors.Wrap(err, "failed to get after service manager info")
	}

	return &model.Response{
		Success: true,
		Message: "성공하였습니다.",
		Data:    data,
	}, nil
}
