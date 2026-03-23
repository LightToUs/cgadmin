package system

import "github.com/flipped-aurora/gin-vue-admin/server/global"

type SysEmailTemplateFolder struct {
	global.GVA_MODEL
	SysUserID uint   `json:"sysUserId" gorm:"column:sys_user_id;index"`
	ParentID  uint   `json:"parentId" gorm:"column:parent_id;index"`
	Name      string `json:"name" gorm:"column:name;size:100"`
	Color     string `json:"color" gorm:"column:color;size:20"`
	Sort      int    `json:"sort" gorm:"column:sort;default:0"`
}

func (SysEmailTemplateFolder) TableName() string {
	return "sys_email_template_folders"
}

