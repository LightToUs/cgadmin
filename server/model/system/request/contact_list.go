package request

type ContactListCreateReq struct {
	ParentID uint   `json:"parentId"`
	Name     string `json:"name"`
}

type ContactListUpdateReq struct {
	ID       uint   `json:"id"`
	ParentID uint   `json:"parentId"`
	Name     string `json:"name"`
}

type ContactListDeleteReq struct {
	ID uint `json:"id"`
}

