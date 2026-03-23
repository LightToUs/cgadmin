package request

type EmailTemplateFolderCreateReq struct {
	ParentID uint   `json:"parentId"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Sort     int    `json:"sort"`
}

type EmailTemplateFolderUpdateReq struct {
	ID       uint   `json:"id"`
	ParentID uint   `json:"parentId"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Sort     int    `json:"sort"`
}

type EmailTemplateFolderDeleteReq struct {
	ID uint `json:"id"`
}

