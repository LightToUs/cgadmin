package system

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	sysModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	sysResp "github.com/flipped-aurora/gin-vue-admin/server/model/system/response"
	"gorm.io/gorm"
)

type ContactListService struct{}

func (s *ContactListService) EnsureDefaults(userID uint) error {
	var existing sysModel.SysContactList
	err := global.GVA_DB.Where("sys_user_id = ? AND type = ? AND name = ?", userID, "system", "全部客户").First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	root := sysModel.SysContactList{
		SysUserID: userID,
		ParentID:  0,
		Name:      "全部客户",
		Type:      "system",
		Rule:      "",
		Sort:      0,
	}
	if err := global.GVA_DB.Create(&root).Error; err != nil {
		return err
	}
	smart := sysModel.SysContactList{
		SysUserID: userID,
		ParentID:  0,
		Name:      "智能列表",
		Type:      "group",
		Rule:      "",
		Sort:      1,
	}
	if err := global.GVA_DB.Create(&smart).Error; err != nil {
		return err
	}
	my := sysModel.SysContactList{
		SysUserID: userID,
		ParentID:  0,
		Name:      "我的列表",
		Type:      "group",
		Rule:      "",
		Sort:      2,
	}
	if err := global.GVA_DB.Create(&my).Error; err != nil {
		return err
	}
	return global.GVA_DB.Create(&[]sysModel.SysContactList{
		{SysUserID: userID, ParentID: smart.ID, Name: "本周新增", Type: "smart", Rule: `{"type":"newThisWeek"}`, Sort: 0},
		{SysUserID: userID, ParentID: smart.ID, Name: "已验证邮箱", Type: "smart", Rule: `{"type":"verified"}`, Sort: 1},
		{SysUserID: userID, ParentID: smart.ID, Name: "已回复", Type: "smart", Rule: `{"type":"replied"}`, Sort: 2},
	}).Error
}

func (s *ContactListService) GetTree(userID uint) ([]sysResp.ContactListNode, error) {
	if err := s.EnsureDefaults(userID); err != nil {
		return nil, err
	}
	var lists []sysModel.SysContactList
	if err := global.GVA_DB.Where("sys_user_id = ?", userID).Order("sort asc, id asc").Find(&lists).Error; err != nil {
		return nil, err
	}
	counts, err := s.countByList(userID, lists)
	if err != nil {
		return nil, err
	}
	nodes := make([]sysResp.ContactListNode, 0, len(lists))
	for _, l := range lists {
		nodes = append(nodes, sysResp.ContactListNode{
			ID:       l.ID,
			ParentID: l.ParentID,
			Name:     l.Name,
			Type:     l.Type,
			Count:    counts[l.ID],
		})
	}
	return buildContactListTree(nodes), nil
}

func (s *ContactListService) Create(userID uint, parentID uint, name string) (sysModel.SysContactList, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return sysModel.SysContactList{}, errors.New("name required")
	}
	if err := s.EnsureDefaults(userID); err != nil {
		return sysModel.SysContactList{}, err
	}
	if parentID != 0 {
		var parent sysModel.SysContactList
		if err := global.GVA_DB.Where("id = ? AND sys_user_id = ?", parentID, userID).First(&parent).Error; err != nil {
			return sysModel.SysContactList{}, err
		}
	}
	item := sysModel.SysContactList{
		SysUserID: userID,
		ParentID:  parentID,
		Name:      name,
		Type:      "custom",
		Rule:      "",
		Sort:      0,
	}
	if err := global.GVA_DB.Create(&item).Error; err != nil {
		return sysModel.SysContactList{}, err
	}
	return item, nil
}

func (s *ContactListService) Update(userID uint, id uint, name string, parentID uint) error {
	if id == 0 {
		return errors.New("id required")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("name required")
	}
	var existing sysModel.SysContactList
	if err := global.GVA_DB.Where("id = ? AND sys_user_id = ?", id, userID).First(&existing).Error; err != nil {
		return err
	}
	if existing.Type != "custom" {
		return errors.New("only custom list can be updated")
	}
	if parentID != 0 {
		var parent sysModel.SysContactList
		if err := global.GVA_DB.Where("id = ? AND sys_user_id = ?", parentID, userID).First(&parent).Error; err != nil {
			return err
		}
	}
	return global.GVA_DB.Model(&sysModel.SysContactList{}).Where("id = ? AND sys_user_id = ?", id, userID).Updates(map[string]any{
		"name":      name,
		"parent_id": parentID,
	}).Error
}

func (s *ContactListService) Delete(userID uint, id uint) error {
	if id == 0 {
		return errors.New("id required")
	}
	var existing sysModel.SysContactList
	if err := global.GVA_DB.Where("id = ? AND sys_user_id = ?", id, userID).First(&existing).Error; err != nil {
		return err
	}
	if existing.Type != "custom" {
		return errors.New("only custom list can be deleted")
	}
	var childCount int64
	if err := global.GVA_DB.Model(&sysModel.SysContactList{}).Where("sys_user_id = ? AND parent_id = ?", userID, id).Count(&childCount).Error; err != nil {
		return err
	}
	if childCount > 0 {
		return errors.New("has children")
	}
	var itemCount int64
	if err := global.GVA_DB.Model(&sysModel.SysContactListItem{}).Where("sys_user_id = ? AND contact_list_id = ?", userID, id).Count(&itemCount).Error; err != nil {
		return err
	}
	if itemCount > 0 {
		return errors.New("not empty")
	}
	return global.GVA_DB.Delete(&sysModel.SysContactList{}, "id = ? AND sys_user_id = ?", id, userID).Error
}

func (s *ContactListService) AddContactToList(tx *gorm.DB, userID uint, listID uint, contactID uint) error {
	if listID == 0 {
		return nil
	}
	var existing sysModel.SysContactListItem
	err := tx.Where("sys_user_id = ? AND contact_list_id = ? AND contact_id = ?", userID, listID, contactID).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return tx.Create(&sysModel.SysContactListItem{
		SysUserID:     userID,
		ContactListID: listID,
		ContactID:     contactID,
	}).Error
}

func buildContactListTree(nodes []sysResp.ContactListNode) []sysResp.ContactListNode {
	byID := map[uint]*sysResp.ContactListNode{}
	for i := range nodes {
		n := nodes[i]
		n.Children = nil
		nodes[i] = n
		byID[n.ID] = &nodes[i]
	}
	roots := make([]sysResp.ContactListNode, 0)
	for i := range nodes {
		n := nodes[i]
		if n.ParentID == 0 {
			roots = append(roots, n)
			continue
		}
		if p, ok := byID[n.ParentID]; ok {
			p.Children = append(p.Children, n)
		} else {
			roots = append(roots, n)
		}
	}
	sort.SliceStable(roots, func(i, j int) bool { return roots[i].ID < roots[j].ID })
	return roots
}

func (s *ContactListService) countByList(userID uint, lists []sysModel.SysContactList) (map[uint]int64, error) {
	out := map[uint]int64{}
	var total int64
	if err := global.GVA_DB.Model(&sysModel.SysContact{}).Where("sys_user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, err
	}
	for _, l := range lists {
		out[l.ID] = 0
		if l.Type == "system" && l.Name == "全部客户" {
			out[l.ID] = total
		}
	}
	var items []sysModel.SysContactListItem
	if err := global.GVA_DB.Where("sys_user_id = ?", userID).Find(&items).Error; err != nil {
		return nil, err
	}
	countMap := map[uint]int64{}
	for _, it := range items {
		countMap[it.ContactListID]++
	}
	for _, l := range lists {
		if l.Type == "custom" {
			out[l.ID] = countMap[l.ID]
		}
		if l.Type == "smart" {
			switch l.Rule {
			case `{"type":"verified"}`:
				var c int64
				if err := global.GVA_DB.Model(&sysModel.SysContact{}).Where("sys_user_id = ? AND email_verify_status in ?", userID, []string{"valid", "risk"}).Count(&c).Error; err != nil {
					return nil, err
				}
				out[l.ID] = c
			case `{"type":"replied"}`:
				var c int64
				if err := global.GVA_DB.Model(&sysModel.SysContact{}).Where("sys_user_id = ? AND status = ?", userID, "replied").Count(&c).Error; err != nil {
					return nil, err
				}
				out[l.ID] = c
			case `{"type":"newThisWeek"}`:
				var c int64
				cutoff := time.Now().Add(-7 * 24 * time.Hour)
				if err := global.GVA_DB.Model(&sysModel.SysContact{}).Where("sys_user_id = ? AND created_at >= ?", userID, cutoff).Count(&c).Error; err != nil {
					return nil, err
				}
				out[l.ID] = c
			}
		}
	}
	return out, nil
}
