package system

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

type SysEmailVerifyJob struct {
	global.GVA_MODEL
	SysUserID uint `json:"sysUserId" gorm:"column:sys_user_id;index"`

	ScopeType string `json:"scopeType" gorm:"column:scope_type;size:30"`
	ListID    *uint  `json:"listId" gorm:"column:list_id;index"`
	Method    string `json:"method" gorm:"column:method;size:20;default:basic"`

	Status   string `json:"status" gorm:"column:status;size:20;default:running"`
	Progress int    `json:"progress" gorm:"column:progress;default:0"`

	Total  int `json:"total" gorm:"column:total;default:0"`
	Valid  int `json:"valid" gorm:"column:valid;default:0"`
	Risk   int `json:"risk" gorm:"column:risk;default:0"`
	Invalid int `json:"invalid" gorm:"column:invalid;default:0"`

	ErrorMessage string `json:"errorMessage" gorm:"column:error_message;type:text"`
	FinishedAt   *time.Time `json:"finishedAt" gorm:"column:finished_at"`
}

func (SysEmailVerifyJob) TableName() string {
	return "sys_email_verify_jobs"
}

