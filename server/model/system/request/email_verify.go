package request

import "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type EmailVerifyStartReq struct {
	ScopeType string `json:"scopeType"`
	ListID    *uint  `json:"listId"`
	Method    string `json:"method"`
}

type EmailVerifyJobGetReq struct {
	ID uint `json:"id" form:"id"`
}

type EmailVerifyHistoryReq struct {
	request.PageInfo
}

