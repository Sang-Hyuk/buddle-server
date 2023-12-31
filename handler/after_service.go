package handler

import (
	"buddle-server/middleware"
	"buddle-server/model"
	"buddle-server/service"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"mime/multipart"
	"net/http"
	"os"
)

type AfterServiceHandler interface {
	Create(c echo.Context) error                      // A/S 신청
	FindAfterServiceInfo(c echo.Context) error        // A/S 신청정보 조회
	FindAfterServiceManagerInfo(c echo.Context) error // A/S 신청정보 조회 ( 관리자용 )
	DownloadFile(c echo.Context) error                // 첨부파일 다운로드
}

type afterServiceHandler struct {
	afterService service.AfterService
}

func NewAfterHandler(afterService service.AfterService) (AfterServiceHandler, error) {
	if afterService == nil {
		return nil, errors.New("after service is nil")
	}

	return &afterServiceHandler{
		afterService: afterService,
	}, nil
}

func (h afterServiceHandler) Create(c echo.Context) error {
	ctx, err := middleware.UpgradeContext(c)
	if err != nil {
		return errors.Wrap(err, "upgrade context")
	}

	afterService := new(model.AfterService)
	if err := ctx.Bind(afterService); err != nil {
		return c.JSON(http.StatusInternalServerError, model.Response{
			Message: "failed to bind request parameter",
		})
	}

	if err := afterService.Validate(); err != nil {
		return c.JSON(http.StatusInternalServerError, model.Response{
			Message: "request params were invalid",
		})
	}

	var src, src2, src3, src4, src5 multipart.File

	file1, err := c.FormFile("file1")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		return c.JSON(http.StatusInternalServerError, model.Response{
			Message: fmt.Sprintf("failed to get receipt param err : %+v", err),
		})
	}
	if file1 != nil {
		if src, err = file1.Open(); err != nil {
			return c.JSON(http.StatusInternalServerError, model.Response{
				Message: "failed to open upload receipt file",
			})
		}
		defer src.Close()
	}

	file2, err := c.FormFile("file2")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		return c.JSON(http.StatusInternalServerError, model.Response{
			Message: fmt.Sprintf("failed to get receipt param err : %+v", err),
		})
	}
	if file2 != nil {
		if src2, err = file2.Open(); err != nil {
			return c.JSON(http.StatusInternalServerError, model.Response{
				Message: "failed to open upload receipt file",
			})
		}
		defer src2.Close()
	}

	file3, err := c.FormFile("file3")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		return c.JSON(http.StatusInternalServerError, model.Response{
			Message: fmt.Sprintf("failed to get receipt param err : %+v", err),
		})
	}
	if file3 != nil {
		if src3, err = file3.Open(); err != nil && !errors.Is(err, http.ErrMissingFile) {
			return c.JSON(http.StatusInternalServerError, model.Response{
				Message: "failed to open upload receipt file",
			})
		}
		defer src3.Close()
	}

	file4, err := c.FormFile("file4")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		return c.JSON(http.StatusInternalServerError, model.Response{
			Message: fmt.Sprintf("failed to get receipt param err : %+v", err),
		})
	}
	if file4 != nil {
		if src4, err = file4.Open(); err != nil {
			return c.JSON(http.StatusInternalServerError, model.Response{
				Message: "failed to open upload receipt file",
			})
		}
		defer src4.Close()
	}

	file5, err := c.FormFile("file5")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		return c.JSON(http.StatusInternalServerError, model.Response{
			Message: fmt.Sprintf("failed to get receipt param err : %+v", err),
		})
	}
	if file5 != nil {
		if src5, err = file5.Open(); err != nil {
			return c.JSON(http.StatusInternalServerError, model.Response{
				Message: "failed to open upload receipt file",
			})
		}
		defer src5.Close()
	}

	resp, err := h.afterService.Create(ctx.GoContext(), afterService, src, src2, src3, src4, src5)
	if err != nil {
		return errors.Wrapf(err, "failed to auth product [ req = %+v ]", *afterService)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h afterServiceHandler) DownloadFile(c echo.Context) error {
	ctx, err := middleware.UpgradeContext(c)
	if err != nil {
		return errors.Wrap(err, "upgrade context")
	}

	var req struct {
		AfterServiceSeq int64 `query:"after_service_seq"`
		FileIdx         int64 `query:"file_idx"`
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

	afterServiceInfo, err := h.afterService.DownloadFile(ctx.GoContext(), req.AfterServiceSeq, req.FileIdx, tmpFile)
	if err != nil {
		logrus.Errorf("failed to download from s3 err:%+v", err)
		return c.JSON(http.StatusOK, model.Response{
			Success: false,
			Message: "파일 다운로드 실패",
		})
	}
	if afterServiceInfo == nil {
		return c.JSON(http.StatusOK, model.Response{
			Success: false,
			Message: "다운로드 받을 파일이 존재하지 않습니다",
		})
	}

	return c.Attachment(tmpFile.Name(), fmt.Sprintf("%s(%s) 첨부파일(%d)", afterServiceInfo.Name, afterServiceInfo.Phone, req.FileIdx))
}

func (h afterServiceHandler) FindAfterServiceInfo(c echo.Context) error {
	ctx, err := middleware.UpgradeContext(c)
	if err != nil {
		return errors.Wrap(err, "upgrade context")
	}

	req := new(model.AfterServiceRequest)
	if err := ctx.Bind(req); err != nil {
		return errors.Wrap(err, "failed to bind request parameter")
	}

	data, err := h.afterService.FindAfterServiceInfo(ctx.GoContext(), *req)
	if err != nil {
		return errors.Wrap(err, "failed to find after service info")
	}

	return c.JSON(http.StatusOK, data)
}

func (h afterServiceHandler) FindAfterServiceManagerInfo(c echo.Context) error {
	ctx, err := middleware.UpgradeContext(c)
	if err != nil {
		return errors.Wrap(err, "upgrade context")
	}

	req := new(model.AfterServiceRequest)
	if err := ctx.Bind(req); err != nil {
		return errors.Wrap(err, "failed to bind request parameter")
	}

	data, err := h.afterService.FindAfterServiceManagerInfo(ctx.GoContext(), *req)
	if err != nil {
		return errors.Wrap(err, "failed to find after service info")
	}

	return c.JSON(http.StatusOK, data)
}
