package system

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

type SysEmailTemplate struct {
	global.GVA_MODEL
	SysUserID uint  `json:"sysUserId" gorm:"column:sys_user_id;index"`
	FolderID  *uint `json:"folderId" gorm:"column:folder_id;index"`

	Name      string `json:"name" gorm:"column:name;size:120"`
	Subject   string `json:"subject" gorm:"column:subject;size:255"`
	BodyHTML  string `json:"bodyHtml" gorm:"column:body_html;type:longtext"`
	BodyText  string `json:"bodyText" gorm:"column:body_text;type:longtext"`
	Status    string `json:"status" gorm:"column:status;size:20;default:enabled"`
	UsageCount int   `json:"usageCount" gorm:"column:usage_count;default:0"`
	OpenRate  float64 `json:"openRate" gorm:"column:open_rate;default:0"`
	ReplyRate float64 `json:"replyRate" gorm:"column:reply_rate;default:0"`
	LastUsedAt *time.Time `json:"lastUsedAt" gorm:"column:last_used_at"`
}

func (SysEmailTemplate) TableName() string {
	return "sys_email_templates"
}

