package model

import (
	"errors"
	jsoniter "github.com/json-iterator/go"
	"time"
)

type AfterService struct {
	AfterServiceSeq int64       `form:"after_service_seq" json:"after_service_seq" gorm:"Column:after_service_seq;PRIMARY_KEY"`
	Name            string      `form:"name" json:"name" gorm:"Column:name"`
	Phone           string      `form:"phone" json:"phone" gorm:"Column:phone"`
	Email           string      `form:"email" json:"email" gorm:"Column:email"`
	Addr            string      `form:"addr" json:"addr" gorm:"Column:addr"`
	AddrDetail      string      `form:"addr_detail" json:"addr_detail" gorm:"Column:addr_detail"`
	ProductType     ProductType `form:"product_type" json:"product_type" gorm:"Column:product_type;default:0"`
	MarketType      MarketType  `form:"market_type" json:"market_type" gorm:"Column:market_type"`
	PurchaseDate    time.Time   `form:"purchase_date" json:"purchase_date" gorm:"Column:purchase_date"`
	File1S3Location string      `form:"file1_s3_location" json:"file1_s3_location" gorm:"Column:file1_s3_location"`
	File2S3Location string      `form:"file2_s3_location" json:"file2_s3_location" gorm:"Column:file2_s3_location"`
	File3S3Location string      `form:"file3_s3_location" json:"file3_s3_location" gorm:"Column:file3_s3_location"`
	File4S3Location string      `form:"file4_s3_location" json:"file4_s3_location" gorm:"Column:file4_s3_location"`
	File5S3Location string      `form:"file5_s3_location" json:"file5_s3_location" gorm:"Column:file5_s3_location"`
	Contents        string      `form:"contents" json:"contents" gorm:"Column:contents"`
	RegDate         time.Time   `form:"regdate" json:"regdate" gorm:"Column:regdate"`
	Modified        time.Time   `form:"modified" json:"modified" gorm:"Column:modified"`
}

func (a AfterService) TableName() string {
	return "after_service"
}

func (a AfterService) Validate() error {
	switch {
	case a.Name == "":
		return errors.New("name is required")
	case a.Phone == "":
		return errors.New("phone is required")
	case a.Addr == "":
		return errors.New("addr is required")
	case a.AddrDetail == "":
		return errors.New("addrDetail is required")
	case a.PurchaseDate.IsZero():
		return errors.New("purchaseDate is required")
	}

	return nil
}

func (a AfterService) MarshalJSON() ([]byte, error) {
	result := struct {
		AfterServiceSeq int64  `json:"after_service_seq"`
		Name            string `json:"name"`
		Phone           string `json:"phone"`
		Email           string `json:"email"`
		Addr            string `json:"addr"`
		AddrDetail      string `json:"addr_detail"`
		ProductType     string `json:"product_type"`
		MarketType      string `json:"market_type"`
		PurchaseDate    string `json:"purchase_date"`
		Contents        string `json:"contents"`
	}{
		AfterServiceSeq: a.AfterServiceSeq,
		Name:            a.Name,
		Phone:           a.Phone,
		Email:           a.Email,
		Addr:            a.Addr,
		AddrDetail:      a.AddrDetail,
		ProductType:     a.ProductType.String(),
		MarketType:      a.MarketType.String(),
		PurchaseDate:    a.PurchaseDate.Format("2006-01-02"),
		Contents:        a.Contents,
	}

	return jsoniter.Marshal(result)
}
