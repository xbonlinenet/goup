package xrpc

const (
	XbErrorCodeSuccess = 0
	XbErrorCodeGeneralFailed = 1
)

// general xiaobu api response
type XbApiResponse struct {
	Code    int                        `json:"code" desc:"0: success others: fail"`
	Message string                     `json:"message" desc:"message need PostInfos" `
	Data    interface{} 				`json:"data"`
}

func BuildSuccessXbResp(data interface{}) *XbApiResponse {
	return &XbApiResponse{
		Code: XbErrorCodeSuccess,
		Message: "Success",
		Data: data,
	}
}

func BuildGeneralFailedXbResp() *XbApiResponse {
	return &XbApiResponse{
		Code: XbErrorCodeGeneralFailed,
		Message: "Failed",
		Data: nil,
	}
}

func BuildFailedXbResp(failedMsg string) *XbApiResponse {
	return &XbApiResponse{
		Code: XbErrorCodeGeneralFailed,
		Message: failedMsg,
		Data: nil,
	}
}

func BuildCustomErrorXbResp(errorCode int, errorMsg string) *XbApiResponse {
	return &XbApiResponse{
		Code: errorCode,
		Message: errorMsg,
		Data: nil,
	}
}