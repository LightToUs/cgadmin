package request

import "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type EmailTemplateCreateReq struct {
	FolderID *uint  `json:"folderId"`
	Name     string `json:"name"`
	Subject  string `json:"subject"`
	BodyHTML string `json:"bodyHtml"`
	BodyText string `json:"bodyText"`
	Status   string `json:"status"`
}

type EmailTemplateUpdateReq struct {
	ID       uint   `json:"id"`
	FolderID *uint  `json:"folderId"`
	Name     string `json:"name"`
	Subject  string `json:"subject"`
	BodyHTML string `json:"bodyHtml"`
	BodyText string `json:"bodyText"`
	Status   string `json:"status"`
}

type EmailTemplateDeleteReq struct {
	ID uint `json:"id"`
}

type EmailTemplateCopyReq struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type EmailTemplateBatchStatusReq struct {
	IDs    []uint `json:"ids"`
	Status string `json:"status"`
}

type EmailTemplateMoveReq struct {
	ID       uint  `json:"id"`
	FolderID *uint `json:"folderId"`
}

type EmailTemplateSearchReq struct {
	request.PageInfo
	Status   string `json:"status" form:"status"`
	FolderID *uint  `json:"folderId" form:"folderId"`
	SortBy   string `json:"sortBy" form:"sortBy"`
	SortDesc bool   `json:"sortDesc" form:"sortDesc"`
}

type EmailTemplatePreviewReq struct {
	Subject  string            `json:"subject"`
	BodyHTML string            `json:"bodyHtml"`
	Vars     map[string]string `json:"vars"`
}

type EmailTemplateTestSendReq struct {
	TemplateID      uint              `json:"templateId"`
	SenderAccountID uint              `json:"senderAccountId"`
	ToEmails        []string          `json:"toEmails"`
	Vars            map[string]string `json:"vars"`
}

type EmailTemplateImportReq struct {
	FolderID *uint  `form:"folderId"`
	Mode     string `form:"mode"`
}

type EmailTemplateExportReq struct {
	IDs    []uint `json:"ids"`
	Format string `json:"format"`
}
