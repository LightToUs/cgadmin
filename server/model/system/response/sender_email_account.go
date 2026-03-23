package response

import "time"

type SenderEmailAccountItem struct {
	ID             uint      `json:"id"`
	Email          string    `json:"email"`
	DisplayName    string    `json:"displayName"`
	Provider       string    `json:"provider"`
	SMTPHost       string    `json:"smtpHost"`
	SMTPPort       int       `json:"smtpPort"`
	Security       string    `json:"security"`
	IsLoginAuth    bool      `json:"isLoginAuth"`
	IsDefault      bool      `json:"isDefault"`
	Status         string    `json:"status"`
	LastError      string    `json:"lastError"`
	LastCheckedAt  time.Time `json:"lastCheckedAt"`
	TodaySent      int       `json:"todaySent"`
	DailyLimit     int       `json:"dailyLimit"`
	IntervalPolicy string    `json:"intervalPolicy"`
}

type SenderEmailAccountListResp struct {
	List []SenderEmailAccountItem `json:"list"`
}
