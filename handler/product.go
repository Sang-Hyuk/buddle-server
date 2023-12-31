package handler

import (
	"buddle-server/middleware"
	"buddle-server/model"
	"buddle-server/service"
	"encoding/csv"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

type ProductHandler interface {
	CreateProduct(c echo.Context) error     // 제품시리얼 정보 등록 ( CSV )
	AuthProduct(c echo.Context) error       // 사용자 제품 인증 ( 정품 인증 )
	ModAuthProduct(c echo.Context) error    // 사용자 제품 인증 정보 변경
	CancelAuthProduct(c echo.Context) error // 사용자 제품 인증 취소 ( 정품 인증 취소 )
	GetAuthProduct(c echo.Context) error    // 사용자 제품 인증 정보 조회( 정품 인증 )
	DownloadReceipt(c echo.Context) error   // 영수증 이미지 다운로드
	FindProductList(c echo.Context) error   // 제품 정보 리스트 ( 인증 정보 포함 )
	UpdateProduct(c echo.Context) error     // 제품 정보 수정
	DeleteProduct(c echo.Context) error     // 제품 정보 수정
}

type productHandler struct {
	productService service.ProductService
}

func NewProductHandler(productService service.ProductService) (ProductHandler, error) {
	if productService == nil {
		return nil, errors.New("product service is nil")
	}

	return &productHandler{
		productService: productService,
	}, nil
}

func (h productHandler) CreateProduct(c echo.Context) error {
	ctx, err := middleware.UpgradeContext(c)
	if err != nil {
		return errors.Wrap(err, "upgrade context")
	}

	// Source
	file, err := c.FormFile("csv_file")
	if err != nil {
		return errors.Wrap(err, "failed to get form_file param")
	}
	src, err := file.Open()
	if err != nil {
		return errors.Wrap(err, "failed to open upload csv file")
	}
	defer src.Close()

	success, failure, err := h.productService.CreateProduct(ctx.GoContext(), csv.NewReader(src))
	if err != nil {
		return errors.Wrap(err, "failed to create product by CSV file")
	}

	return c.JSON(http.StatusOK, struct {
		Success int64 `json:"success"`
		Failure int64 `json:"failure"`
	}{
		Success: success,
		Failure: failure,
	})
}

func (h productHandler) AuthProduct(c echo.Context) error {
	ctx, err := middleware.UpgradeContext(c)
	if err != nil {
		return errors.Wrap(err, "upgrade context")
	}

	productRegist := new(model.ProductRegist)
	if err := ctx.Bind(productRegist); err != nil {
		return c.JSON(http.StatusInternalServerError, model.Response{
			Message: "failed to bind request parameter",
		})
	}

	if err := productRegist.Validate(); err != nil {
		return c.JSON(http.StatusInternalServerError, model.Response{
			Message: "request params were invalid",
		})
	}

	file, err := c.FormFile("receipt")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Response{
			Message: fmt.Sprintf("failed to get receipt param err : %+v", err),
		})
	}
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Response{
			Message: "failed to open upload receipt file",
		})
	}
	defer src.Close()

	resp, err := h.productService.AuthProduct(ctx.GoContext(), productRegist, src)
	if err != nil {
		return errors.Wrapf(err, "failed to auth product [ req = %+v ]", *productRegist)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h productHandler) ModAuthProduct(c echo.Context) error {
	ctx, err := middleware.UpgradeContext(c)
	if err != nil {
		return errors.Wrap(err, "upgrade context")
	}

	var productRegistSeq int64
	if err := echo.PathParamsBinder(ctx).Int64("product_regist_seq", &productRegistSeq).BindError(); err != nil {
		return errors.Wrap(err, "failed to bind product_regist_seq param")
	}

	if productRegistSeq <= 0 {
		return fmt.Errorf("invalid product_regist_seq param (%d)", productRegistSeq)
	}

	productRegist := new(model.ProductRegist)
	if err := ctx.Bind(productRegist); err != nil {
		return c.JSON(http.StatusInternalServerError, model.Response{
			Message: "failed to bind request parameter",
		})
	}
	productRegist.ProductRegistSeq = productRegistSeq

	if err := h.productService.ModAuthProduct(ctx.GoContext(), productRegist); err != nil {
		return errors.Wrapf(err, "failed to cancel product auth [ product_regist_seq = %d ]", productRegistSeq)
	}

	return c.JSON(http.StatusOK, model.Response{Success: true, Message: "수정 되었습니다."})
}

func (h productHandler) CancelAuthProduct(c echo.Context) error {
	ctx, err := middleware.UpgradeContext(c)
	if err != nil {
		return errors.Wrap(err, "upgrade context")
	}

	var productRegistSeq int64
	if err := echo.PathParamsBinder(ctx).Int64("product_regist_seq", &productRegistSeq).BindError(); err != nil {
		return errors.Wrap(err, "failed to bind product_regist_seq param")
	}

	if productRegistSeq <= 0 {
		return fmt.Errorf("invalid product_regist_seq param (%d)", productRegistSeq)
	}

	if err := h.productService.CancelAuthProduct(ctx.GoContext(), productRegistSeq); err != nil {
		return errors.Wrapf(err, "failed to cancel product auth [ product_regist_seq = %d ]", productRegistSeq)
	}

	return c.JSON(http.StatusOK, model.Response{Success: true})
}

func (h productHandler) GetAuthProduct(c echo.Context) error {
	ctx, err := middleware.UpgradeContext(c)
	if err != nil {
		return errors.Wrap(err, "upgrade context")
	}

	req := new(model.ProductAuthRequest)
	if err := ctx.Bind(req); err != nil {
		return errors.Wrap(err, "failed to bind request parameter")
	}

	data, err := h.productService.GetAuthProductInfo(ctx.GoContext(), *req)
	if err != nil {
		return errors.Wrap(err, "failed to find product manage info")
	}

	return c.JSON(http.StatusOK, data)
}

func (h productHandler) FindProductList(c echo.Context) error {
	ctx, err := middleware.UpgradeContext(c)
	if err != nil {
		return errors.Wrap(err, "upgrade context")
	}

	req := new(model.ProductManageRequest)
	if err := ctx.Bind(req); err != nil {
		return errors.Wrap(err, "failed to bind request parameter")
	}

	data, err := h.productService.FindProductManageInfo(ctx.GoContext(), *req)
	if err != nil {
		return errors.Wrap(err, "failed to find product manage info")
	}

	return c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    data,
	})
}

func (h productHandler) DownloadReceipt(c echo.Context) error {
	ctx, err := middleware.UpgradeContext(c)
	if err != nil {
		return errors.Wrap(err, "upgrade context")
	}

	var req struct {
		ProductRegistSeq int64 `query:"product_regist_seq"`
	}

	if err := ctx.Bind(&req); err != nil {
		return errors.Wrap(err, "failed to bind request parameter")
	}

	tmpFile, err := os.CreateTemp("/root/.buddle", "*")
	if err != nil {
		logrus.Errorf("failed to create temp file err:%+v", err)
		return c.JSON(http.StatusOK, model.Response{
			Success: false,
			Message: "임시 파일 생성 실패",
		})
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			logrus.Errorf("An error occurred to close temp file: %+v", err)
		}
		if err := os.Remove(tmpFile.Name()); err != nil {
			logrus.Errorf("An error occurred to delete temp file: %+v", err)
		}
	}()

	productRegistInfo, err := h.productService.DownloadReceipt(ctx.GoContext(), req.ProductRegistSeq, tmpFile)
	if err != nil {
		logrus.Errorf("failed to download from s3 err:%+v", err)
		return c.JSON(http.StatusOK, model.Response{
			Success: false,
			Message: "파일 다운로드 실패",
		})
	}

	return c.Attachment(tmpFile.Name(), fmt.Sprintf("%s(%s) 영수증", productRegistInfo.Name, productRegistInfo.Phone))
}

func (h productHandler) UpdateProduct(c echo.Context) error {
	panic("implement me")
}

func (h productHandler) DeleteProduct(c echo.Context) error {
	panic("implement me")
}
