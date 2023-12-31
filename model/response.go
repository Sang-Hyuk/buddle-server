package model

type ResponseErrorCode string

const (
	ResponseErrorCodeDuplProduct     ResponseErrorCode = "1000" // 이미 등록된 제품을 등록 하려는 경우
	ResponseErrorCodeProductNotExist ResponseErrorCode = "1001" // 제품이 존재하지 않음
	ResponseErrorCodeUserIDNotExist  ResponseErrorCode = "1002" // 회원 아이디가 존재하지 않음
	ResponseErrorCodeInvalidUserPwd  ResponseErrorCode = "1003" // 회원 비밀번호가 일치하지 않음

)

type Response struct {
	Success   bool              `json:"success"`
	Message   string            `json:"message,omitempty"`
	ErrorCode ResponseErrorCode `json:"error_code,omitempty"`
	Data      interface{}       `json:"data,omitempty"`
}

func SimpleSuccess() *Response {
	return &Response{
		Success: true,
		Message: "성공하였습니다.",
	}
}

func SimpleFail() *Response {
	return &Response{
		Success: true,
	}
}
