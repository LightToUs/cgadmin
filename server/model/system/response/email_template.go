package response

import "time"

type EmailTemplateItem struct {
	ID        uint      `json:"id"`
	FolderID  *uint     `json:"folderId"`
	Name      string    `json:"name"`
	Subject   string    `json:"subject"`
	BodyHTML  string    `json:"bodyHtml"`
	Status    string    `json:"status"`
	UsageCount int      `json:"usageCount"`
	OpenRate  float64   `json:"openRate"`
	ReplyRate float64   `json:"replyRate"`
	UpdatedAt time.Time `json:"updatedAt"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
}

type EmailTemplatePreviewResp struct {
	Subject  string `json:"subject"`
	BodyHTML string `json:"bodyHtml"`
	BodyText string `json:"bodyText"`
}

