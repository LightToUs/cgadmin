package system

import "github.com/flipped-aurora/gin-vue-admin/server/global"

type SysContactList struct {
	global.GVA_MODEL
	SysUserID uint   `json:"sysUserId" gorm:"column:sys_user_id;index"`
	ParentID  uint   `json:"parentId" gorm:"column:parent_id;index"`
	Name      string `json:"name" gorm:"column:name;size:120"`
	Type      string `json:"type" gorm:"column:type;size:20;default:custom"`
	Rule      string `json:"rule" gorm:"column:rule;type:longtext"`
	Sort      int    `json:"sort" gorm:"column:sort;default:0"`
}

func (SysContactList) TableName() string {
	return "sys_contact_lists"
}

