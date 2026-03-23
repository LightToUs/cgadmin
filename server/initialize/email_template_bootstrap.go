package initialize

import (
	"errors"

	adapter "github.com/casbin/gorm-adapter/v3"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	sysModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func EnsureEmailTemplateBootstrap() {
	if global.GVA_DB == nil {
		return
	}
	if global.GVA_CONFIG.System.DisableAutoMigrate {
		return
	}

	parent, listMenu, editMenu, err := ensureEmailTemplateMenus()
	if err != nil {
		global.GVA_LOG.Error("ensure email template menus failed", zap.Error(err))
	} else {
		if err := ensureAuthorityMenu("888", []uint{parent.ID, listMenu.ID, editMenu.ID}); err != nil {
			global.GVA_LOG.Error("ensure email template authority menu failed", zap.Error(err))
		}
	}

	if err := ensureEmailTemplateCasbinPolicies("888"); err != nil {
		global.GVA_LOG.Error("ensure email template casbin policies failed", zap.Error(err))
	}
}

func ensureEmailTemplateMenus() (sysModel.SysBaseMenu, sysModel.SysBaseMenu, sysModel.SysBaseMenu, error) {
	var parent sysModel.SysBaseMenu
	err := global.GVA_DB.Where("name = ?", "contentCenter").First(&parent).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		parent = sysModel.SysBaseMenu{
			MenuLevel: 0,
			Hidden:    false,
			ParentId:  0,
			Path:      "contentCenter",
			Name:      "contentCenter",
			Component: "view/routerHolder.vue",
			Sort:      6,
			Meta:      sysModel.Meta{Title: "内容中心", Icon: "document"},
		}
		if err := global.GVA_DB.Create(&parent).Error; err != nil {
			return sysModel.SysBaseMenu{}, sysModel.SysBaseMenu{}, sysModel.SysBaseMenu{}, err
		}
	} else if err != nil {
		return sysModel.SysBaseMenu{}, sysModel.SysBaseMenu{}, sysModel.SysBaseMenu{}, err
	}

	var listMenu sysModel.SysBaseMenu
	err = global.GVA_DB.Where("name = ?", "emailTemplate").First(&listMenu).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		listMenu = sysModel.SysBaseMenu{
			MenuLevel: 1,
			Hidden:    false,
			ParentId:  parent.ID,
			Path:      "emailTemplate",
			Name:      "emailTemplate",
			Component: "view/contentCenter/emailTemplate/index.vue",
			Sort:      1,
			Meta:      sysModel.Meta{Title: "邮件模板", Icon: "message"},
		}
		if err := global.GVA_DB.Create(&listMenu).Error; err != nil {
			return sysModel.SysBaseMenu{}, sysModel.SysBaseMenu{}, sysModel.SysBaseMenu{}, err
		}
	} else if err != nil {
		return sysModel.SysBaseMenu{}, sysModel.SysBaseMenu{}, sysModel.SysBaseMenu{}, err
	} else if listMenu.ParentId == 0 || listMenu.ParentId != parent.ID {
		_ = global.GVA_DB.Model(&sysModel.SysBaseMenu{}).Where("id = ?", listMenu.ID).Update("parent_id", parent.ID).Error
	}

	var editMenu sysModel.SysBaseMenu
	err = global.GVA_DB.Where("name = ?", "emailTemplateEdit").First(&editMenu).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		editMenu = sysModel.SysBaseMenu{
			MenuLevel: 1,
			Hidden:    true,
			ParentId:  parent.ID,
			Path:      "emailTemplateEdit/:id",
			Name:      "emailTemplateEdit",
			Component: "view/contentCenter/emailTemplate/edit.vue",
			Sort:      0,
			Meta:      sysModel.Meta{Title: "编辑邮件模板", Icon: "edit"},
		}
		if err := global.GVA_DB.Create(&editMenu).Error; err != nil {
			return sysModel.SysBaseMenu{}, sysModel.SysBaseMenu{}, sysModel.SysBaseMenu{}, err
		}
	} else if err != nil {
		return sysModel.SysBaseMenu{}, sysModel.SysBaseMenu{}, sysModel.SysBaseMenu{}, err
	} else if editMenu.ParentId == 0 || editMenu.ParentId != parent.ID {
		_ = global.GVA_DB.Model(&sysModel.SysBaseMenu{}).Where("id = ?", editMenu.ID).Update("parent_id", parent.ID).Error
	}

	return parent, listMenu, editMenu, nil
}

func ensureEmailTemplateCasbinPolicies(authorityID string) error {
	policies := []adapter.CasbinRule{
		{Ptype: "p", V0: authorityID, V1: "/emailTemplate/list", V2: "GET"},
		{Ptype: "p", V0: authorityID, V1: "/emailTemplate/detail", V2: "GET"},
		{Ptype: "p", V0: authorityID, V1: "/emailTemplate/create", V2: "POST"},
		{Ptype: "p", V0: authorityID, V1: "/emailTemplate/update", V2: "PUT"},
		{Ptype: "p", V0: authorityID, V1: "/emailTemplate/delete", V2: "DELETE"},
		{Ptype: "p", V0: authorityID, V1: "/emailTemplate/deleteByIds", V2: "DELETE"},
		{Ptype: "p", V0: authorityID, V1: "/emailTemplate/copy", V2: "POST"},
		{Ptype: "p", V0: authorityID, V1: "/emailTemplate/batchStatus", V2: "POST"},
		{Ptype: "p", V0: authorityID, V1: "/emailTemplate/move", V2: "POST"},
		{Ptype: "p", V0: authorityID, V1: "/emailTemplate/preview", V2: "POST"},
		{Ptype: "p", V0: authorityID, V1: "/emailTemplate/testSend", V2: "POST"},
		{Ptype: "p", V0: authorityID, V1: "/emailTemplate/export", V2: "POST"},
		{Ptype: "p", V0: authorityID, V1: "/emailTemplate/import", V2: "POST"},

		{Ptype: "p", V0: authorityID, V1: "/emailTemplateFolder/tree", V2: "GET"},
		{Ptype: "p", V0: authorityID, V1: "/emailTemplateFolder/create", V2: "POST"},
		{Ptype: "p", V0: authorityID, V1: "/emailTemplateFolder/update", V2: "PUT"},
		{Ptype: "p", V0: authorityID, V1: "/emailTemplateFolder/delete", V2: "DELETE"},
	}
	for _, p := range policies {
		var existing adapter.CasbinRule
		err := global.GVA_DB.Where(adapter.CasbinRule{Ptype: p.Ptype, V0: p.V0, V1: p.V1, V2: p.V2}).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := global.GVA_DB.Create(&p).Error; err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}
