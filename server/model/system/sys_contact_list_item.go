package system

import "github.com/flipped-aurora/gin-vue-admin/server/global"

type SysContactListItem struct {
	global.GVA_MODEL
	SysUserID    uint `json:"sysUserId" gorm:"column:sys_user_id;index"`
	ContactListID uint `json:"contactListId" gorm:"column:contact_list_id;index"`
	ContactID     uint `json:"contactId" gorm:"column:contact_id;index"`
}

func (SysContactListItem) TableName() string {
	return "sys_contact_list_items"
}

