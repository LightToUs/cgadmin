package system

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

type SysSenderEmailAccount struct {
	global.GVA_MODEL
	SysUserID      uint      `json:"sysUserId" gorm:"index;not null;column:sys_user_id"`
	Email          string    `json:"email" gorm:"index;not null;size:255"`
	DisplayName    string    `json:"displayName" gorm:"size:255"`
	Provider       string    `json:"provider" gorm:"index;not null;size:32"`
	SMTPHost       string    `json:"smtpHost" gorm:"not null;size:255"`
	SMTPPort       int       `json:"smtpPort" gorm:"not null"`
	Security       string    `json:"security" gorm:"not null;size:16"`
	IsLoginAuth    bool      `json:"isLoginAuth" gorm:"not null;default:false"`
	SecretEnc      string    `json:"-" gorm:"type:text;not null;column:secret_enc"`
	IsDefault      bool      `json:"isDefault" gorm:"index;not null;default:false"`
	Status         string    `json:"status" gorm:"index;not null;default:'error';size:32"`
	LastError      string    `json:"lastError" gorm:"type:text"`
	LastCheckedAt  time.Time `json:"lastCheckedAt"`
	TodaySent      int       `json:"todaySent" gorm:"not null;default:0"`
	TodaySentDate  string    `json:"todaySentDate" gorm:"not null;default:'';size:10"`
	DailyLimit     int       `json:"dailyLimit" gorm:"not null;default:50"`
	IntervalPolicy string    `json:"intervalPolicy" gorm:"not null;default:'1m';size:16"`
}

func (SysSenderEmailAccount) TableName() string {
	return "sys_sender_email_accounts"
}
