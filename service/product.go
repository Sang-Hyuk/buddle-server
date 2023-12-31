package service

import (
	"buddle-server/internal/s3"
	"buddle-server/model"
	"buddle-server/repository"
	"context"
	"encoding/csv"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type ProductService interface {
	CreateProduct(c context.Context, csvReader *csv.Reader) (int64, int64, error)
	AuthProduct(c context.Context, productRegist *model.ProductRegist, file io.Reader) (*model.Response, error)
	ModAuthProduct(c context.Context, productRegist *model.ProductRegist) error
	CancelAuthProduct(c context.Context, productRegistSeq int64) error
	GetAuthProductInfo(c context.Context, req model.ProductAuthRequest) (*model.Response, error)
	DownloadReceipt(c context.Context, productRegistSeq int64, file *os.File) (*model.ProductRegist, error)
	FindProductManageInfo(c context.Context, req model.ProductManageRequest) (model.ProductManageInfos, error)
}

type productService struct {
	repo       repository.Repository
	fileBucket *s3.S3
}

func NewProductService(repo repository.Repository, fileBucket *s3.S3) (ProductService, error) {
	if repo == nil {
		return nil, errors.New("repository is nil")
	}

	return &productService{repo: repo, fileBucket: fileBucket}, nil
}

func (s productService) CreateProduct(c context.Context, csvReader *csv.Reader) (success int64, failure int64, err error) {
	switch {
	case c == nil:
		return 0, 0, errors.New("nil context")
	case csvReader == nil:
		return 0, 0, errors.New("csvReader is nil")
	}

	rows, err := csvReader.ReadAll()
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to read upload csv file")
	}

	for i, row := range rows {
		for j := range row {
			// 모든공백제거
			rows[i][j] = strings.ReplaceAll(rows[i][j], " ", "")
		}

		productType, err := strconv.ParseInt(rows[i][1], 10, 64)
		if err != nil {
			logrus.Errorf("failed to parse csv data [ csv = %+v ]", rows)
			failure++
			continue
		}

		product := &model.Product{
			SerialNo:    rows[i][0],
			ProductType: model.ProductType(productType),
		}

		if err := s.repo.Product().Create(c, product); err != nil {
			logrus.Errorf("failed to create product info [ err = %+v ]", err)
			failure++
			continue
		}

		success++
	}

	return
}

func (s productService) CancelAuthProduct(c context.Context, productRegistSeq int64) error {
	if c == nil {
		return errors.New("nil context")
	}

	return s.repo.Product().CancelProductAuth(c, productRegistSeq)
}

func (s productService) AuthProduct(c context.Context, productRegist *model.ProductRegist, file io.Reader) (*model.Response, error) {
	switch {
	case c == nil:
		return model.SimpleFail(), errors.New("nil context")
	case productRegist == nil:
		return model.SimpleFail(), errors.New("nil request params")
	case file == nil:
		return model.SimpleFail(), errors.New("nil receipt file")
	}

	// 시리얼 번호 인증
	product, err := s.repo.Product().GetProductBySerial(c, productRegist.SerialNo, productRegist.ProductType)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.Response{
				Success:   false,
				Message:   "시리얼 번호가 잘못 되었습니다.",
				ErrorCode: model.ResponseErrorCodeProductNotExist,
			}, nil
		}
		return model.SimpleFail(), errors.Wrap(err, "failed to auth product")
	}

	// 기존 인증이 존재하는지 확인 ( 인증 완료 된 것 중에서만 찾음 )
	originProductRegist, err := s.repo.Product().GetProductRegistByProductSeq(c, product.ProductSeq)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Wrap(err, "failed to get product regist by product sequence")
	}

	if originProductRegist != nil && originProductRegist.ProductRegistSeq > 0 {
		return &model.Response{
			Success:   false,
			Message:   "이미 등록된 제품입니다.",
			ErrorCode: model.ResponseErrorCodeDuplProduct,
		}, nil
	}

	// s3 업로드
	s3location := fmt.Sprintf("%s/%d", time.Now().Format("2006-01-02"), product.ProductSeq)
	if s.fileBucket != nil {
		if err := s.fileBucket.Upload(c, s3location, file); err != nil {
			logrus.Errorf("failed to upload s3 object: objectKey=%s, bucketName=%s err:%+v", s3location, s.fileBucket.BucketName(), err)
		}
	}

	productRegist.ProductSeq = product.ProductSeq
	productRegist.ReceiptS3Location = s3location

	if err := s.repo.Product().CreateProductRegist(c, productRegist); err != nil {
		return nil, errors.Wrap(err, "failed to create product regist")
	}

	return model.SimpleSuccess(), nil
}

func (s productService) ModAuthProduct(c context.Context, productRegist *model.ProductRegist) error {
	switch {
	case c == nil:
		return errors.New("nil context")
	case productRegist == nil:
		return errors.New("nil product regist")
	}

	return s.repo.Product().ModProductAuth(c, productRegist)
}

func (s productService) GetAuthProductInfo(c context.Context, req model.ProductAuthRequest) (*model.Response, error) {
	if c == nil {
		return nil, errors.New("nil context")
	}

	data, err := s.repo.Product().GetProductAuthInfo(c, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get product auth info")
	}
	if data.Name == "" {
		return &model.Response{
			Success: false,
			Message: "일치하는 데이터가 없습니다.",
		}, nil
	}

	return &model.Response{
		Success: true,
		Message: "성공하였습니다.",
		Data:    data,
	}, nil
}

func (s productService) DownloadReceipt(c context.Context, productRegistSeq int64, file *os.File) (*model.ProductRegist, error) {
	switch {
	case c == nil:
		return nil, errors.New("nil context")
	case productRegistSeq == 0:
		return nil, errors.New("invalid product regist sequence")
	case file == nil:
		return nil, errors.New("file is nil")
	}

	productRegistInfo, err := s.repo.Product().GetProductRegistBySeq(c, productRegistSeq)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get product regist by seq(%d)", productRegistSeq)
	}

	if s.fileBucket != nil {
		if _, err := s.fileBucket.Download(c, file, productRegistInfo.ReceiptS3Location); err != nil {
			return nil, errors.Wrapf(err, "failed to download file [ s3 location : %+v ]", productRegistInfo.ReceiptS3Location)
		}
	}

	return productRegistInfo, nil
}

func (s productService) FindProductManageInfo(c context.Context, req model.ProductManageRequest) (model.ProductManageInfos, error) {
	if c == nil {
		return nil, errors.New("nil context")
	}

	return s.repo.Product().FindProductManageInfo(c, req)
}
