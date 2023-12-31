package repository

import (
	"buddle-server/internal/db"
	"buddle-server/model"
	"context"
	"database/sql"
	"fmt"
	"github.com/pkg/errors" //nolint:goimports
	"time"
)

type ProductRepository interface {
	Create(c context.Context, product *model.Product) error
	CreateProductRegist(c context.Context, productRegist *model.ProductRegist) error
	CancelProductAuth(c context.Context, productRegistSeq int64) error
	ModProductAuth(c context.Context, productRegist *model.ProductRegist) error
	GetProductBySerial(c context.Context, serial string, productType model.ProductType) (*model.Product, error)
	GetProductRegistByProductSeq(c context.Context, productSeq int64) (*model.ProductRegist, error)
	GetProductRegistBySeq(c context.Context, productRegistSeq int64) (*model.ProductRegist, error)
	FindProductManageInfo(c context.Context, req model.ProductManageRequest) (model.ProductManageInfos, error)
	GetProductAuthInfo(c context.Context, req model.ProductAuthRequest) (*model.ProductAuthInfo, error)
	UpdateProduct(c context.Context) error
}

type productRepository struct{}

func NewProductRepository() ProductRepository {
	return &productRepository{}
}

func (r productRepository) Create(c context.Context, product *model.Product) error {
	switch {
	case c == nil:
		return fmt.Errorf("context is nil")
	case product == nil:
		return fmt.Errorf("product is nil")
	}

	conn, err := db.ConnFromContext(c, db.WriteDBKey)
	if err != nil {
		return errors.Wrap(err, "failed to get db connection")
	}

	if product.RegDate.IsZero() {
		product.RegDate = time.Now()
	}

	product.Modified = product.RegDate

	return conn.Create(product).Error
}

func (r productRepository) UpdateProduct(c context.Context) error {
	panic("implement me")
}

func (r productRepository) CreateProductRegist(c context.Context, productRegist *model.ProductRegist) error {
	switch {
	case c == nil:
		return errors.New("nil context")
	case productRegist == nil:
		return errors.New("product regist is nil")
	case productRegist.ProductSeq == 0:
		return errors.New("product sequence is invalid")
	}

	conn, err := db.ConnFromContext(c, db.WriteDBKey)
	if err != nil {
		return errors.Wrap(err, "failed to get db connection")
	}

	if productRegist.Regdate.IsZero() {
		productRegist.Regdate = time.Now()
	}

	productRegist.Modified = productRegist.Regdate
	productRegist.Status = model.ProductAuthStatusOK

	return conn.Create(productRegist).Error
}

func (r productRepository) CancelProductAuth(c context.Context, productRegistSeq int64) error {
	switch {
	case c == nil:
		return errors.New("nil context")
	case productRegistSeq == 0:
		return errors.New("product sequence is invalid")
	}

	conn, err := db.ConnFromContext(c, db.WriteDBKey)
	if err != nil {
		return errors.Wrap(err, "failed to get db connection")
	}

	if err := conn.Exec("UPDATE product_regist SET status = @ProductAuthStatusCancel, modified = @modified WHERE product_regist_seq = @productRegistSeq",
		sql.Named("ProductAuthStatusCancel", model.ProductAuthStatusCancel),
		sql.Named("modified", time.Now()),
		sql.Named("productRegistSeq", productRegistSeq)).Error; err != nil {
		return errors.Wrap(err, "failed to cancel product auth")
	}

	return nil
}

func (r productRepository) ModProductAuth(c context.Context, productRegist *model.ProductRegist) error {
	switch {
	case c == nil:
		return errors.New("nil context")
	case productRegist == nil:
		return errors.New("product register is nil")
	case productRegist.ProductRegistSeq == 0:
		return errors.New("product sequence is invalid")
	}

	conn, err := db.ConnFromContext(c, db.WriteDBKey)
	if err != nil {
		return errors.Wrap(err, "failed to get db connection")
	}

	if err = conn.Model(productRegist).Updates(productRegist.ToUpdateMap()).Error; err != nil {
		return errors.Wrap(err, "failed to modified product auth")
	}

	return nil
}

func (r productRepository) GetProductBySerial(c context.Context, serial string, productType model.ProductType) (*model.Product, error) {
	switch {
	case c == nil:
		return nil, errors.New("nil context")
	case serial == "":
		return nil, errors.New("serial number is required")
	}

	conn, err := db.ConnFromContext(c, db.WriteDBKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get db connection")
	}

	product := new(model.Product)
	if err := conn.Where("serial_no = ?", serial).Where("product_type = ?", productType).Take(&product).Error; err != nil {
		return nil, errors.Wrap(err, "failed to get product by serial no")
	}

	return product, nil
}

func (r productRepository) GetProductRegistByProductSeq(c context.Context, productSeq int64) (*model.ProductRegist, error) {
	switch {
	case c == nil:
		return nil, errors.New("nil context")
	case productSeq == 0:
		return nil, errors.New("product sequence is required")
	}

	conn, err := db.ConnFromContext(c, db.WriteDBKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get db connection")
	}

	productRegist := new(model.ProductRegist)
	if err := conn.Where("product_seq = ? AND status = ?", productSeq, model.ProductAuthStatusOK).Take(&productRegist).Error; err != nil {
		return nil, errors.Wrap(err, "failed to get product regist by product sequence")
	}

	return productRegist, nil
}

func (r productRepository) GetProductRegistBySeq(c context.Context, productRegistSeq int64) (*model.ProductRegist, error) {
	switch {
	case c == nil:
		return nil, errors.New("nil context")
	case productRegistSeq == 0:
		return nil, errors.New("product sequence is required")
	}

	conn, err := db.ConnFromContext(c, db.WriteDBKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get db connection")
	}

	productRegist := new(model.ProductRegist)
	if err := conn.First(&productRegist, productRegistSeq).Error; err != nil {
		return nil, errors.Wrap(err, "failed to get product regist by product sequence")
	}

	return productRegist, nil
}

func (r productRepository) FindProductManageInfo(c context.Context, req model.ProductManageRequest) (model.ProductManageInfos, error) {
	if c == nil {
		return nil, errors.New("nil context")
	}

	conn, err := db.ConnFromContext(c, db.WriteDBKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get db connection")
	}

	tx := conn.Table("product p").Select(
		[]string{
			"p.product_seq",
			"p.serial_no",
			"p.product_type",
			"p.regdate AS product_regdate",
			"pr.product_regist_seq",
			"pr.name",
			"pr.phone",
			"pr.addr",
			"pr.addr_detail",
			"pr.market_type",
			"pr.purchase_date",
			"pr.regdate AS product_regist_regdate",
			"pr.status",
			"CONCAT(pr.name, '(', pr.phone, ') 영수증') AS filename",
		},
	)

	switch req.AuthStatus {
	case model.ProductAuthStatusOK:
		tx.Joins("INNER JOIN product_regist pr ON p.product_seq = pr.product_seq AND pr.status = ?", model.ProductAuthStatusOK)
	case model.ProductAuthStatusCancel:
		tx.Joins("INNER JOIN product_regist pr ON p.product_seq = pr.product_seq AND pr.status = ?", model.ProductAuthStatusCancel)
	default:
		tx.Joins("LEFT JOIN product_regist pr ON p.product_seq = pr.product_seq")
	}

	if req.Name != "" {
		tx.Where("pr.name LIKE ?", req.Name+"%")
	}

	if req.Phone != "" {
		tx.Where("pr.phone LIKE ?", req.Phone+"%")
	}

	if req.SerialNo != "" {
		tx.Where("p.serial_no LIKE ?", req.SerialNo+"%")
	}

	if req.Limit > 0 {
		tx.Limit(req.Limit).Offset(req.Offset)
	}

	result := make(model.ProductManageInfos, 0)
	if err := tx.Order("pr.name").Scan(&result).Error; err != nil {
		return nil, errors.Wrap(err, "failed to execute find product manage info list query")
	}

	return result, nil
}

func (r productRepository) GetProductAuthInfo(c context.Context, req model.ProductAuthRequest) (*model.ProductAuthInfo, error) {
	if c == nil {
		return nil, errors.New("nil context")
	}

	conn, err := db.ConnFromContext(c, db.WriteDBKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get db connection")
	}

	tx := conn.Table("product p").
		Select(
			[]string{
				"pr.name",
				"pr.phone",
				"p.product_type",
				"pr.market_type",
				"pr.purchase_date",
				"p.serial_no",
			},
		).
		Joins("INNER JOIN product_regist pr ON p.product_seq = pr.product_seq").
		Where("pr.name = ?", req.Name).
		Where("pr.phone = ?", req.Phone).
		Order("pr.regdate desc").
		Limit(1)

	result := model.ProductAuthInfo{}
	if err := tx.Order("pr.name").Scan(&result).Error; err != nil {
		return nil, errors.Wrap(err, "failed to execute find product manage info list query")
	}

	return &result, nil
}
