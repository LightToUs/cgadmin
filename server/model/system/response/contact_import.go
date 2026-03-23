package response

import "time"

type ContactImportJobSummary struct {
	ID       uint      `json:"id"`
	Filename string    `json:"filename"`
	FileType string    `json:"fileType"`
	Status   string    `json:"status"`
	Progress int       `json:"progress"`
	Total    int       `json:"total"`
	ValidCount          int `json:"validCount"`
	InvalidCount        int `json:"invalidCount"`
	DuplicateCount      int `json:"duplicateCount"`
	MissingRequiredCount int `json:"missingRequiredCount"`
	CreatedCount int `json:"createdCount"`
	UpdatedCount int `json:"updatedCount"`
	FailedCount  int `json:"failedCount"`
	ErrorMessage string `json:"errorMessage"`
	CreatedAt time.Time `json:"createdAt"`
	FinishedAt *time.Time `json:"finishedAt"`
}

type ContactImportUploadResp struct {
	JobID   uint     `json:"jobId"`
	Columns []string `json:"columns"`
	Sample  []map[string]string `json:"sample"`
}

type ContactImportSuggestResp struct {
	Columns []string          `json:"columns"`
	Suggest map[string]string `json:"suggest"`
}

type ContactImportValidateResp struct {
	Total int `json:"total"`
	ValidCount int `json:"validCount"`
	InvalidCount int `json:"invalidCount"`
	DuplicateCount int `json:"duplicateCount"`
	MissingRequiredCount int `json:"missingRequiredCount"`
}

type ContactImportExportFailedResp struct {
	Token string `json:"token"`
}

