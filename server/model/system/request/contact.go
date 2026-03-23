package request

import "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type ContactCreateReq struct {
	CompanyName string `json:"companyName"`
	Website     string `json:"website"`
	ContactName string `json:"contactName"`
	Title       string `json:"title"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Country     string `json:"country"`
	Tags        string `json:"tags"`
	Status      string `json:"status"`
}

type ContactUpdateReq struct {
	ID          uint   `json:"id"`
	CompanyName string `json:"companyName"`
	Website     string `json:"website"`
	ContactName string `json:"contactName"`
	Title       string `json:"title"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Country     string `json:"country"`
	Tags        string `json:"tags"`
	Status      string `json:"status"`
}

type ContactDeleteReq struct {
	ID uint `json:"id"`
}

type ContactListReq struct {
	request.PageInfo
	ListID   string `json:"listId" form:"listId"`
	Keyword  string `json:"keyword" form:"keyword"`
	Country  string `json:"country" form:"country"`
	Status   string `json:"status" form:"status"`
	Verified string `json:"verified" form:"verified"`
	SortBy   string `json:"sortBy" form:"sortBy"`
	SortDesc bool   `json:"sortDesc" form:"sortDesc"`
}

