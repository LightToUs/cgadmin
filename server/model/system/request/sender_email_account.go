package request

type SenderEmailAccountCreateReq struct {
	Email          string `json:"email"`
	DisplayName    string `json:"displayName"`
	Provider       string `json:"provider"`
	SMTPHost       string `json:"smtpHost"`
	SMTPPort       int    `json:"smtpPort"`
	Security       string `json:"security"`
	IsLoginAuth    bool   `json:"isLoginAuth"`
	Secret         string `json:"secret"`
	DailyLimit     int    `json:"dailyLimit"`
	IntervalPolicy string `json:"intervalPolicy"`
}

type SenderEmailAccountUpdateReq struct {
	ID             uint   `json:"id"`
	Email          string `json:"email"`
	DisplayName    string `json:"displayName"`
	Provider       string `json:"provider"`
	SMTPHost       string `json:"smtpHost"`
	SMTPPort       int    `json:"smtpPort"`
	Security       string `json:"security"`
	IsLoginAuth    bool   `json:"isLoginAuth"`
	Secret         string `json:"secret"`
	DailyLimit     int    `json:"dailyLimit"`
	IntervalPolicy string `json:"intervalPolicy"`
}

type SenderEmailAccountTestReq struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	SMTPHost    string `json:"smtpHost"`
	SMTPPort    int    `json:"smtpPort"`
	Security    string `json:"security"`
	IsLoginAuth bool   `json:"isLoginAuth"`
	Secret      string `json:"secret"`
}

type SenderEmailAccountTestByIDReq struct {
	ID uint `json:"id"`
}

type SenderEmailAccountSetDefaultReq struct {
	ID uint `json:"id"`
}

type SenderEmailAccountDeleteReq struct {
	ID uint `json:"id"`
}

type SenderEmailAccountQuotaUpdateReq struct {
	ID             uint   `json:"id"`
	DailyLimit     int    `json:"dailyLimit"`
	IntervalPolicy string `json:"intervalPolicy"`
}
