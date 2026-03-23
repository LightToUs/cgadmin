package system

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

type SysContactImportJob struct {
	global.GVA_MODEL
	SysUserID uint `json:"sysUserId" gorm:"column:sys_user_id;index"`

	Filename string `json:"filename" gorm:"column:filename;size:255"`
	FilePath string `json:"filePath" gorm:"column:file_path;size:500"`
	FileType string `json:"fileType" gorm:"column:file_type;size:20"`

	Status   string `json:"status" gorm:"column:status;size:20;default:uploaded"`
	Progress int    `json:"progress" gorm:"column:progress;default:0"`

	ColumnsJSON string `json:"columnsJson" gorm:"column:columns_json;type:longtext"`
	SampleJSON  string `json:"sampleJson" gorm:"column:sample_json;type:longtext"`

	MappingJSON string `json:"mappingJson" gorm:"column:mapping_json;type:longtext"`
	OptionsJSON string `json:"optionsJson" gorm:"column:options_json;type:longtext"`

	Total                int `json:"total" gorm:"column:total;default:0"`
	ValidCount           int `json:"validCount" gorm:"column:valid_count;default:0"`
	InvalidCount         int `json:"invalidCount" gorm:"column:invalid_count;default:0"`
	DuplicateCount       int `json:"duplicateCount" gorm:"column:duplicate_count;default:0"`
	MissingRequiredCount int `json:"missingRequiredCount" gorm:"column:missing_required_count;default:0"`

	CreatedCount int `json:"createdCount" gorm:"column:created_count;default:0"`
	UpdatedCount int `json:"updatedCount" gorm:"column:updated_count;default:0"`
	FailedCount  int `json:"failedCount" gorm:"column:failed_count;default:0"`

	ErrorFilePath string `json:"errorFilePath" gorm:"column:error_file_path;size:500"`
	ErrorMessage  string `json:"errorMessage" gorm:"column:error_message;type:text"`

	FinishedAt *time.Time `json:"finishedAt" gorm:"column:finished_at"`
}

func (SysContactImportJob) TableName() string {
	return "sys_contact_import_jobs"
}
