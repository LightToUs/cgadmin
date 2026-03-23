package system

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

type SysContact struct {
	global.GVA_MODEL
	SysUserID uint `json:"sysUserId" gorm:"column:sys_user_id;index"`

	CompanyName string `json:"companyName" gorm:"column:company_name;size:200"`
	Website     string `json:"website" gorm:"column:website;size:255"`
	ContactName string `json:"contactName" gorm:"column:contact_name;size:120"`
	Title       string `json:"title" gorm:"column:title;size:120"`
	Email       string `json:"email" gorm:"column:email;size:190;index"`
	Phone       string `json:"phone" gorm:"column:phone;size:50"`
	Country     string `json:"country" gorm:"column:country;size:80"`

	Status string `json:"status" gorm:"column:status;size:30;default:uncontacted"`

	EmailVerifyStatus string     `json:"emailVerifyStatus" gorm:"column:email_verify_status;size:20;default:unverified"`
	EmailVerifiedAt   *time.Time `json:"emailVerifiedAt" gorm:"column:email_verified_at"`

	Tags string `json:"tags" gorm:"column:tags;type:text"`
}

func (SysContact) TableName() string {
	return "sys_contacts"
}

