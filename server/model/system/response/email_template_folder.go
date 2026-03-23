package response

type EmailTemplateFolderNode struct {
	ID       uint                      `json:"id"`
	ParentID uint                      `json:"parentId"`
	Name     string                    `json:"name"`
	Color    string                    `json:"color"`
	Sort     int                       `json:"sort"`
	Children []EmailTemplateFolderNode `json:"children"`
}

