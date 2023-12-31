package model

type ProductManageRequest struct {
	Limit      int               `json:"limit,omitempty" query:"limit"`
	Offset     int               `json:"offset,omitempty" query:"offset"`
	AuthStatus ProductAuthStatus `json:"status" query:"status"`
	SerialNo   string            `json:"serial_no,omitempty" query:"serial_no"`
	Name       string            `json:"name,omitempty" query:"name"`
	Phone      string            `json:"phone,omitempty" query:"phone"`
}

func (r ProductManageRequest) Validate() error {
	return nil
}

type ProductAuthRequest struct {
	Name  string `json:"name,omitempty" query:"name"`
	Phone string `json:"phone,omitempty" query:"phone"`
}

type AfterServiceRequest struct {
	Name  string `json:"name,omitempty" query:"name"`
	Phone string `json:"phone,omitempty" query:"phone"`
}
