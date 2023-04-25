package rpcserver

type GeneralResponse struct {
	Code string      `json:"code" example:"OK"` // Code is OK for normal cases and ErrXXXX for errors.
	Msg  string      `json:"msg"`               // Msg is "" for normal cases and message for errors.
	Data interface{} `json:"data,omitempty"`    // Optional
}

type PagingOrderRequest struct {
	PagingParams
	OrderParams
}

type PagingOrderFilterRequest struct {
	PagingOrderRequest
	Filter string `json:"filter"` // Any search field supported. May not be supported.
}

type PagingResponse struct {
	GeneralResponse
	List  interface{} `json:"list,omitempty"` // List is always the result list
	Size  int         `json:"size"`           // Size is the result count in this response1
	Total int64       `json:"total"`          // Total is the total result in database
	Page  int         `json:"page"`           // Page is the given params in the request starts from 1
}

type DebugUAResponse struct {
	BrowserName string `json:"browser_name"`
	DeviceType  string `json:"device_type"`
	OsName      string `json:"os_name"`
	OsPlatform  string `json:"os_platform"`
	DbPlatform  string `json:"db_platform"`
}

type DebugResponse struct {
	Ip    string `json:"ip"`
	Value string `json:"value"`
}

type TimeRangeRequest struct {
	FromTime int64 `json:"from_time"`
	ToTime   int64 `json:"to_time"`
}

func DefaultPagingParams(params PagingParams) PagingParams {
	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Offset < 0 {
		params.Offset = 0
	}
	return params
}

func DefaultOrderParams(params OrderParams) OrderParams {
	return params
}
