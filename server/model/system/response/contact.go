package response

import "time"

type ContactItem struct {
	ID uint `json:"id"`
	CompanyName string `json:"companyName"`
	Website string `json:"website"`
	ContactName string `json:"contactName"`
	Title string `json:"title"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	Country string `json:"country"`
	Status string `json:"status"`
	EmailVerifyStatus string `json:"emailVerifyStatus"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ContactListNode struct {
	ID uint `json:"id"`
	ParentID uint `json:"parentId"`
	Name string `json:"name"`
	Type string `json:"type"`
	Count int64 `json:"count"`
	Children []ContactListNode `json:"children,omitempty"`
}

