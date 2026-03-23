package request

import "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type ContactImportUploadReq struct {
	Source string `form:"source"`
}

type ContactImportSuggestReq struct {
	JobID uint `json:"jobId"`
}

type ContactImportValidateReq struct {
	JobID   uint              `json:"jobId"`
	Mapping map[string]string `json:"mapping"`
	Options ContactImportOptions `json:"options"`
}

type ContactImportStartReq struct {
	JobID   uint              `json:"jobId"`
	Mapping map[string]string `json:"mapping"`
	Options ContactImportOptions `json:"options"`
}

type ContactImportOptions struct {
	OnInvalid string `json:"onInvalid"`
	OnDuplicate string `json:"onDuplicate"`
	ListID *uint `json:"listId"`
}

type ContactImportJobListReq struct {
	request.PageInfo
}

type ContactImportJobGetReq struct {
	ID uint `json:"id" form:"id"`
}

type ContactImportJobDeleteReq struct {
	ID uint `json:"id"`
}

type ContactImportErrorsReq struct {
	JobID uint `json:"jobId" form:"jobId"`
	Type  string `json:"type" form:"type"`
	request.PageInfo
}

type ContactImportExportFailedReq struct {
	JobID uint `json:"jobId"`
}

type ContactImportGoogleSheetReq struct {
	URL string `json:"url"`
}

