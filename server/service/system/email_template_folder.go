package system

import (
	"errors"
	"sort"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	sysModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	sysReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	sysResp "github.com/flipped-aurora/gin-vue-admin/server/model/system/response"
	"gorm.io/gorm"
)

type EmailTemplateFolderService struct{}

func (s *EmailTemplateFolderService) Create(userID uint, req sysReq.EmailTemplateFolderCreateReq) (sysModel.SysEmailTemplateFolder, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return sysModel.SysEmailTemplateFolder{}, errors.New("name required")
	}
	f := sysModel.SysEmailTemplateFolder{
		SysUserID: userID,
		ParentID:  req.ParentID,
		Name:      req.Name,
		Color:     strings.TrimSpace(req.Color),
		Sort:      req.Sort,
	}
	if f.ParentID != 0 {
		var parent sysModel.SysEmailTemplateFolder
		if err := global.GVA_DB.Where("id = ? AND sys_user_id = ?", f.ParentID, userID).First(&parent).Error; err != nil {
			return sysModel.SysEmailTemplateFolder{}, err
		}
	}
	if err := global.GVA_DB.Create(&f).Error; err != nil {
		return sysModel.SysEmailTemplateFolder{}, err
	}
	return f, nil
}

func (s *EmailTemplateFolderService) Update(userID uint, req sysReq.EmailTemplateFolderUpdateReq) error {
	req.Name = strings.TrimSpace(req.Name)
	if req.ID == 0 {
		return errors.New("id required")
	}
	if req.Name == "" {
		return errors.New("name required")
	}
	if req.ParentID != 0 && req.ParentID == req.ID {
		return errors.New("invalid parent")
	}
	if req.ParentID != 0 {
		var parent sysModel.SysEmailTemplateFolder
		if err := global.GVA_DB.Where("id = ? AND sys_user_id = ?", req.ParentID, userID).First(&parent).Error; err != nil {
			return err
		}
	}
	return global.GVA_DB.Model(&sysModel.SysEmailTemplateFolder{}).Where("id = ? AND sys_user_id = ?", req.ID, userID).Updates(map[string]any{
		"parent_id": req.ParentID,
		"name":      req.Name,
		"color":     strings.TrimSpace(req.Color),
		"sort":      req.Sort,
	}).Error
}

func (s *EmailTemplateFolderService) Delete(userID uint, id uint) error {
	if id == 0 {
		return errors.New("id required")
	}
	var childCount int64
	if err := global.GVA_DB.Model(&sysModel.SysEmailTemplateFolder{}).Where("sys_user_id = ? AND parent_id = ?", userID, id).Count(&childCount).Error; err != nil {
		return err
	}
	if childCount > 0 {
		return errors.New("folder has children")
	}
	var templateCount int64
	if err := global.GVA_DB.Model(&sysModel.SysEmailTemplate{}).Where("sys_user_id = ? AND folder_id = ?", userID, id).Count(&templateCount).Error; err != nil {
		return err
	}
	if templateCount > 0 {
		return errors.New("folder has templates")
	}
	return global.GVA_DB.Delete(&sysModel.SysEmailTemplateFolder{}, "id = ? AND sys_user_id = ?", id, userID).Error
}

func (s *EmailTemplateFolderService) GetTree(userID uint) ([]sysResp.EmailTemplateFolderNode, error) {
	var folders []sysModel.SysEmailTemplateFolder
	if err := global.GVA_DB.Where("sys_user_id = ?", userID).Order("sort asc, id asc").Find(&folders).Error; err != nil {
		return nil, err
	}

	type folderNode struct {
		ID       uint
		ParentID uint
		Name     string
		Color    string
		Sort     int
		Children []*folderNode
	}

	nodeByID := make(map[uint]*folderNode, len(folders))
	for _, f := range folders {
		nodeByID[f.ID] = &folderNode{
			ID:       f.ID,
			ParentID: f.ParentID,
			Name:     f.Name,
			Color:    f.Color,
			Sort:     f.Sort,
			Children: nil,
		}
	}

	var roots []*folderNode
	for _, f := range folders {
		n := nodeByID[f.ID]
		if f.ParentID == 0 {
			roots = append(roots, n)
			continue
		}
		parent := nodeByID[f.ParentID]
		if parent == nil {
			roots = append(roots, n)
			continue
		}
		parent.Children = append(parent.Children, n)
	}

	var sortNodes func(nodes []*folderNode)
	sortNodes = func(nodes []*folderNode) {
		sort.SliceStable(nodes, func(i, j int) bool {
			if nodes[i].Sort != nodes[j].Sort {
				return nodes[i].Sort < nodes[j].Sort
			}
			return nodes[i].ID < nodes[j].ID
		})
		for i := range nodes {
			if len(nodes[i].Children) > 0 {
				sortNodes(nodes[i].Children)
			}
		}
	}

	sortNodes(roots)

	var toResp func(n *folderNode) sysResp.EmailTemplateFolderNode
	toResp = func(n *folderNode) sysResp.EmailTemplateFolderNode {
		out := sysResp.EmailTemplateFolderNode{
			ID:       n.ID,
			ParentID: n.ParentID,
			Name:     n.Name,
			Color:    n.Color,
			Sort:     n.Sort,
			Children: nil,
		}
		if len(n.Children) > 0 {
			out.Children = make([]sysResp.EmailTemplateFolderNode, 0, len(n.Children))
			for _, c := range n.Children {
				out.Children = append(out.Children, toResp(c))
			}
		}
		return out
	}

	resp := make([]sysResp.EmailTemplateFolderNode, 0, len(roots))
	for _, n := range roots {
		resp = append(resp, toResp(n))
	}
	return resp, nil
}

func (s *EmailTemplateFolderService) GetByID(userID uint, id uint) (sysModel.SysEmailTemplateFolder, error) {
	var f sysModel.SysEmailTemplateFolder
	err := global.GVA_DB.Where("id = ? AND sys_user_id = ?", id, userID).First(&f).Error
	return f, err
}

func ensureFolderExists(db *gorm.DB, userID uint, folderID *uint) error {
	if folderID == nil {
		return nil
	}
	if *folderID == 0 {
		return errors.New("invalid folder")
	}
	var f sysModel.SysEmailTemplateFolder
	return db.Where("id = ? AND sys_user_id = ?", *folderID, userID).First(&f).Error
}
