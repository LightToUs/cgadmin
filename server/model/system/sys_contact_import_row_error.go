package system

import "github.com/flipped-aurora/gin-vue-admin/server/global"

type SysContactImportRowError struct {
	global.GVA_MODEL
	SysUserID uint `json:"sysUserId" gorm:"column:sys_user_id;index"`
	JobID     uint `json:"jobId" gorm:"column:job_id;index"`
	RowIndex  int  `json:"rowIndex" gorm:"column:row_index"`

	RawJSON    string `json:"rawJson" gorm:"column:raw_json;type:longtext"`
	ErrorsJSON string `json:"errorsJson" gorm:"column:errors_json;type:longtext"`
}

func (SysContactImportRowError) TableName() string {
	return "sys_contact_import_row_errors"
}

