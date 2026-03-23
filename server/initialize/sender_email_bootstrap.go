package initialize

import (
	"errors"
	"strconv"

	adapter "github.com/casbin/gorm-adapter/v3"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	sysModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func EnsureSenderEmailBootstrap() {
	if global.GVA_DB == nil {
		return
	}
	if global.GVA_CONFIG.System.DisableAutoMigrate {
		return
	}

	sendConfigMenu, emailAccountMenu, err := ensureSenderEmailMenus()
	if err != nil {
		global.GVA_LOG.Error("ensure sender email menus failed", zap.Error(err))
	} else {
		if err := ensureAuthorityMenu("888", []uint{sendConfigMenu.ID, emailAccountMenu.ID}); err != nil {
			global.GVA_LOG.Error("ensure sender email authority menu failed", zap.Error(err))
		}
	}

	if err := ensureSenderEmailCasbinPolicies("888"); err != nil {
		global.GVA_LOG.Error("ensure sender email casbin policies failed", zap.Error(err))
	}
}

func ensureSenderEmailMenus() (sysModel.SysBaseMenu, sysModel.SysBaseMenu, error) {
	var parent sysModel.SysBaseMenu
	err := global.GVA_DB.Where("name = ?", "sendConfig").First(&parent).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		parent = sysModel.SysBaseMenu{
			MenuLevel: 0,
			Hidden:    false,
			ParentId:  0,
			Path:      "sendConfig",
			Name:      "sendConfig",
			Component: "view/routerHolder.vue",
			Sort:      6,
			Meta:      sysModel.Meta{Title: "发件配置", Icon: "message"},
		}
		if err := global.GVA_DB.Create(&parent).Error; err != nil {
			return sysModel.SysBaseMenu{}, sysModel.SysBaseMenu{}, err
		}
	} else if err != nil {
		return sysModel.SysBaseMenu{}, sysModel.SysBaseMenu{}, err
	}

	var child sysModel.SysBaseMenu
	err = global.GVA_DB.Where("name = ?", "sendEmailAccount").First(&child).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		child = sysModel.SysBaseMenu{
			MenuLevel: 1,
			Hidden:    false,
			ParentId:  parent.ID,
			Path:      "emailAccount",
			Name:      "sendEmailAccount",
			Component: "view/sendConfig/emailAccount/index.vue",
			Sort:      1,
			Meta:      sysModel.Meta{Title: "邮箱账号", Icon: "message"},
		}
		if err := global.GVA_DB.Create(&child).Error; err != nil {
			return sysModel.SysBaseMenu{}, sysModel.SysBaseMenu{}, err
		}
	} else if err != nil {
		return sysModel.SysBaseMenu{}, sysModel.SysBaseMenu{}, err
	} else if child.ParentId == 0 || child.ParentId != parent.ID {
		_ = global.GVA_DB.Model(&sysModel.SysBaseMenu{}).Where("id = ?", child.ID).Update("parent_id", parent.ID).Error
	}

	return parent, child, nil
}

func ensureAuthorityMenu(authorityID string, menuIDs []uint) error {
	for _, mid := range menuIDs {
		record := sysModel.SysAuthorityMenu{
			MenuId:      strconv.FormatUint(uint64(mid), 10),
			AuthorityId: authorityID,
		}
		var existing sysModel.SysAuthorityMenu
		err := global.GVA_DB.Where(record).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := global.GVA_DB.Create(&record).Error; err != nil {
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

func ensureSenderEmailCasbinPolicies(authorityID string) error {
	policies := []adapter.CasbinRule{
		{Ptype: "p", V0: authorityID, V1: "/senderEmailAccount/list", V2: "GET"},
		{Ptype: "p", V0: authorityID, V1: "/senderEmailAccount/create", V2: "POST"},
		{Ptype: "p", V0: authorityID, V1: "/senderEmailAccount/update", V2: "PUT"},
		{Ptype: "p", V0: authorityID, V1: "/senderEmailAccount/updateQuota", V2: "PUT"},
		{Ptype: "p", V0: authorityID, V1: "/senderEmailAccount/delete", V2: "DELETE"},
		{Ptype: "p", V0: authorityID, V1: "/senderEmailAccount/setDefault", V2: "POST"},
		{Ptype: "p", V0: authorityID, V1: "/senderEmailAccount/test", V2: "POST"},
		{Ptype: "p", V0: authorityID, V1: "/senderEmailAccount/testById", V2: "POST"},
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
