package repository

import (
	"buddle-server/internal/db"
	"buddle-server/model"
	"context"
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"time"
)

type AfterServiceRepository interface {
	Create(c context.Context, as *model.AfterService) error
	FindAfterServiceInfo(c context.Context, req model.AfterServiceRequest) ([]*model.AfterService, error)
	FindAfterServiceManagerInfo(c context.Context, req model.AfterServiceRequest) ([]*model.AfterService, error)
	GetAfterServiceBySeq(c context.Context, afterServiceSeq int64) (*model.AfterService, error)
}

type afterServiceRepository struct {
}

func NewAfterServiceRepository() AfterServiceRepository {
	return &afterServiceRepository{}
}

func (r afterServiceRepository) Create(c context.Context, as *model.AfterService) error {
	switch {
	case c == nil:
		return fmt.Errorf("context is nil")
	case as == nil:
		return fmt.Errorf("after service info is nil")
	}

	conn, err := db.ConnFromContext(c, db.WriteDBKey)
	if err != nil {
		return errors.Wrap(err, "failed to get db connection")
	}

	if as.RegDate.IsZero() {
		as.RegDate = time.Now()
	}

	as.Modified = as.RegDate

	return conn.Debug().Create(as).Error
}

func (r afterServiceRepository) FindAfterServiceInfo(c context.Context, req model.AfterServiceRequest) ([]*model.AfterService, error) {
	if c == nil {
		return nil, errors.New("nil context")
	}

	conn, err := db.ConnFromContext(c, db.WriteDBKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get db connection")
	}

	result := make([]*model.AfterService, 0)
	if err := conn.Where("name=?", req.Name).Where("phone=?", req.Phone).Order("regdate desc").Find(&result).Error; err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(err, "failed to get after service info")
	}

	return result, nil
}

func (r afterServiceRepository) FindAfterServiceManagerInfo(c context.Context, req model.AfterServiceRequest) ([]*model.AfterService, error) {
	if c == nil {
		return nil, errors.New("nil context")
	}

	conn, err := db.ConnFromContext(c, db.WriteDBKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get db connection")
	}

	result := make([]*model.AfterService, 0)

	tx := conn.Order("regdate desc")

	if req.Name != "" {
		tx = tx.Where("name=?", req.Name)
	}

	if req.Phone != "" {
		tx = tx.Where("phone=?", req.Phone)
	}

	if err := tx.Find(&result).Error; err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(err, "failed to get manager after service info")
	}

	return result, nil
}

func (r afterServiceRepository) GetAfterServiceBySeq(c context.Context, afterServiceSeq int64) (*model.AfterService, error) {
	switch {
	case c == nil:
		return nil, errors.New("nil context")
	case afterServiceSeq == 0:
		return nil, errors.New("after service sequence is required")
	}

	conn, err := db.ConnFromContext(c, db.WriteDBKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get db connection")
	}

	result := new(model.AfterService)
	if err := conn.First(&result, afterServiceSeq).Error; err != nil {
		return nil, errors.Wrap(err, "failed to get after service by sequence")
	}

	return result, nil
}
