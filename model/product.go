package model

import (
	"errors"
	"time"

	jsoniter "github.com/json-iterator/go"
)

type ProductType int

const (
	ProductTypePowderMilkMaker      ProductType = iota // 버들맘마 분유제조기 플러스
	ProductTypePowderMilkMakerSmart                    // 버들맘마 분유제조기 스마트
	ProductTypeBabyBottleWasher                        // 버들아이 젖병세척기
	ProductTypeSmartChopper                            // 버들 스마트 차퍼
)

func (t ProductType) String() string {
	switch t {
	case ProductTypePowderMilkMaker:
		return "버들맘마 분유제조기 플러스"
	case ProductTypePowderMilkMakerSmart:
		return "버들맘마 분유제조기 스마트"
	case ProductTypeBabyBottleWasher:
		return "버들아이 젖병세척기"
	case ProductTypeSmartChopper:
		return "버들 스마트 차퍼"
	}

	return ""
}

type Product struct {
	ProductSeq  int64       `json:"product_seq,omitempty" gorm:"Column:product_seq;PRIMARY_KEY"`
	SerialNo    string      `json:"serial_no,omitempty" gorm:"Column:serial_no"`
	ProductType ProductType `json:"product_type,omitempty" gorm:"Column:product_type"`
	RegDate     time.Time   `json:"reg_date,omitempty" gorm:"Column:regdate"`
	Modified    time.Time   `json:"modified,omitempty" gorm:"Column:modified"`
}

func (p Product) TableName() string {
	return "product"
}

type MarketType int

const (
	MarketTypeNaver MarketType = iota
	MarketTypeCoupang
	MarketTypeOffline
	MarketTypeEtc
)

func (t MarketType) String() string {
	switch t {
	case MarketTypeNaver:
		return "네이버"
	case MarketTypeCoupang:
		return "쿠팡"
	case MarketTypeOffline:
		return "오프라인 매장"
	default:
		return "기타"
	}
}

type ProductAuthStatus int

const (
	ProductAuthStatusNone   ProductAuthStatus = iota // 미인증
	ProductAuthStatusCancel                          // 인증 취소
	ProductAuthStatusOK                              // 인증 완료
)

type ProductRegist struct {
	ProductRegistSeq  int64             `form:"product_regist_seq" json:"product_regist_seq,omitempty" gorm:"Column:product_regist_seq;PRIMARY_KEY"`
	ProductSeq        int64             `form:"product_seq" json:"product_seq,omitempty" gorm:"Column:product_seq"`
	Name              string            `form:"name" json:"name,omitempty" gorm:"Column:name"`
	Phone             string            `form:"phone" json:"phone,omitempty" gorm:"Column:phone"`
	Addr              string            `form:"addr" json:"addr,omitempty" gorm:"Column:addr"`
	AddrDetail        string            `form:"addr_detail" json:"addr_detail,omitempty" gorm:"Column:addr_detail"`
	SerialNo          string            `form:"serial_no" json:"serial_no" gorm:"-"`
	ProductType       ProductType       `form:"product_type" json:"product_type" gorm:"-"`
	MarketType        MarketType        `form:"market_type" json:"market_type,omitempty" gorm:"Column:market_type"`
	Status            ProductAuthStatus `form:"status" json:"status,omitempty" gorm:"Column:status"`
	PurchaseDate      time.Time         `form:"purchase_date" json:"purchase_date" gorm:"Column:purchase_date"`
	ReceiptS3Location string            `form:"receipt_s3_location" json:"receipt_s3_location,omitempty" gorm:"Column:receipt_s3_location"`
	Regdate           time.Time         `form:"regdate" json:"regdate" gorm:"Column:regdate"`
	Modified          time.Time         `form:"modified" json:"modified" gorm:"Column:modified"`
}

func (pr ProductRegist) ToUpdateMap() map[string]interface{} {
	return map[string]interface{}{
		"name":          pr.Name,
		"phone":         pr.Phone,
		"addr":          pr.Addr,
		"addr_detail":   pr.AddrDetail,
		"purchase_date": pr.PurchaseDate,
		"regdate":       pr.Regdate,
		"modified":      time.Now(),
	}
}

func (pr ProductRegist) TableName() string {
	return "product_regist"
}

func (pr ProductRegist) Validate() error {
	switch {
	case pr.Name == "":
		return errors.New("name is required")
	case pr.Phone == "":
		return errors.New("phone is required")
	case pr.Addr == "":
		return errors.New("addr is required")
	case pr.AddrDetail == "":
		return errors.New("addrDetail is required")
	case pr.PurchaseDate.IsZero():
		return errors.New("purchaseDate is required")
	case pr.SerialNo == "":
		return errors.New("serial number is required")
	}

	return nil
}

type ProductManageInfos []ProductManageInfo

type ProductManageInfo struct {
	ProductSeq           int64             `json:"product_seq,omitempty"`
	SerialNo             string            `json:"serial_no,omitempty"`
	ProductType          ProductType       `json:"product_type,omitempty"`
	ProductRegdate       time.Time         `json:"product_regdate"`
	ProductRegistSeq     int64             `json:"product_regist_seq,omitempty"`
	Name                 string            `json:"name,omitempty"`
	Phone                string            `json:"phone,omitempty"`
	Addr                 string            `json:"addr,omitempty"`
	AddrDetail           string            `json:"addr_detail,omitempty"`
	MarketType           MarketType        `json:"market_type,omitempty"`
	Status               ProductAuthStatus `json:"status,omitempty"`
	PurchaseDate         time.Time         `json:"purchase_date"`
	ProductRegistRegdate time.Time         `json:"product_regist_regdate"`
	Filename             string            `json:"filename,omitempty"`
}

func (p ProductManageInfos) MarshalJSON() ([]byte, error) {
	type Result struct {
		ProductSeq           int64             `json:"product_seq,omitempty"`
		SerialNo             string            `json:"serial_no,omitempty"`
		ProductType          ProductType       `json:"product_type,omitempty"`
		ProductRegdate       string            `json:"product_regdate"`
		ProductRegistSeq     int64             `json:"product_regist_seq,omitempty"`
		Name                 string            `json:"name,omitempty"`
		Phone                string            `json:"phone,omitempty"`
		Addr                 string            `json:"addr,omitempty"`
		AddrDetail           string            `json:"addr_detail,omitempty"`
		MarketType           MarketType        `json:"market_type,omitempty"`
		Status               ProductAuthStatus `json:"status,omitempty"`
		PurchaseDate         string            `json:"purchase_date"`
		ProductRegistRegdate string            `json:"product_regist_regdate"`
		Filename             string            `json:"filename,omitempty"`
	}

	results := make([]Result, 0)

	for _, info := range p {
		results = append(results, Result{
			ProductSeq:           info.ProductSeq,
			SerialNo:             info.SerialNo,
			ProductType:          info.ProductType,
			ProductRegdate:       info.ProductRegdate.Format("2006-01-02"),
			ProductRegistSeq:     info.ProductRegistSeq,
			Name:                 info.Name,
			Phone:                info.Phone,
			Addr:                 info.Addr,
			AddrDetail:           info.AddrDetail,
			MarketType:           info.MarketType,
			Status:               info.Status,
			PurchaseDate:         info.PurchaseDate.Format("2006-01-02"),
			ProductRegistRegdate: info.ProductRegistRegdate.Format("2006-01-02"),
			Filename:             info.Filename,
		})
	}

	return jsoniter.Marshal(results)
}

type ProductAuthInfo struct {
	Name         string      `json:"name,omitempty" gorm:"Column:name"`
	Phone        string      `json:"phone,omitempty" gorm:"Column:phone"`
	ProductType  ProductType `json:"product_type,omitempty" gorm:"Column:product_type"`
	MarketType   MarketType  `json:"market_type,omitempty" gorm:"Column:market_type"`
	PurchaseDate time.Time   `json:"purchase_date,omitempty" gorm:"Column:purchase_date"`
	SerialNo     string      `json:"serial_no,omitempty" gorm:"Column:serial_no"`
}

func (pa ProductAuthInfo) TableName() string {
	return "product_regist"
}

func (pa ProductAuthInfo) MarshalJSON() ([]byte, error) {
	result := struct {
		Name         string `json:"name,omitempty"`
		Phone        string `json:"phone,omitempty"`
		ProductType  string `json:"product_type,omitempty"`
		MarketType   string `json:"market_type,omitempty"`
		PurchaseDate string `json:"purchase_date,omitempty"`
		SerialNo     string `json:"serial_no,omitempty"`
	}{
		Name:         pa.Name,
		Phone:        pa.Phone,
		ProductType:  pa.ProductType.String(),
		MarketType:   pa.MarketType.String(),
		PurchaseDate: pa.PurchaseDate.Format("2006-01-02"),
		SerialNo:     pa.SerialNo,
	}

	return jsoniter.Marshal(result)
}
