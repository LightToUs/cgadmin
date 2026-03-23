package initialize

import (
	"errors"

	adapter "github.com/casbin/gorm-adapter/v3"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	sysModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func EnsureContactBootstrap() {
	if global.GVA_DB == nil {
		return
	}
	if global.GVA_CONFIG.System.DisableAutoMigrate {
		return
	}

	menus, err := ensureContactMenus()
	if err != nil {
		global.GVA_LOG.Error("ensure contact menus failed", zap.Error(err))
	} else {
		if err := ensureAuthorityMenu("888", menus); err != nil {
			global.GVA_LOG.Error("ensure contact authority menu failed", zap.Error(err))
		}
	}

	if err := ensureContactCasbinPolicies("888"); err != nil {
		global.GVA_LOG.Error("ensure contact casbin policies failed", zap.Error(err))
	}
}

func ensureContactMenus() ([]uint, error) {
	parent, err := ensureMenu(sysModel.SysBaseMenu{
		MenuLevel: 0, Hidden: false, ParentId: 0,
		Path: "contact", Name: "contact", Component: "view/routerHolder.vue", Sort: 6,
		Meta: sysModel.Meta{Title: "联系人", Icon: "user"},
	})
	if err != nil {
		return nil, err
	}
	tools, err := ensureMenu(sysModel.SysBaseMenu{
		MenuLevel: 1, Hidden: false, ParentId: parent.ID,
		Path: "tools", Name: "contactTools", Component: "view/routerHolder.vue", Sort: 5,
		Meta: sysModel.Meta{Title: "工具", Icon: "tools"},
	})
	if err != nil {
		return nil, err
	}

	importMenu, err := ensureMenu(sysModel.SysBaseMenu{
		MenuLevel: 1, Hidden: false, ParentId: parent.ID,
		Path: "import", Name: "contactImport", Component: "view/contact/import/index.vue", Sort: 1,
		Meta: sysModel.Meta{Title: "导入客户", Icon: "upload"},
	})
	if err != nil {
		return nil, err
	}
	historyMenu, err := ensureMenu(sysModel.SysBaseMenu{
		MenuLevel: 1, Hidden: false, ParentId: parent.ID,
		Path: "importHistory", Name: "contactImportHistory", Component: "view/contact/import/history.vue", Sort: 2,
		Meta: sysModel.Meta{Title: "导入历史", Icon: "list"},
	})
	if err != nil {
		return nil, err
	}
	manualAddMenu, err := ensureMenu(sysModel.SysBaseMenu{
		MenuLevel: 1, Hidden: false, ParentId: parent.ID,
		Path: "manualAdd", Name: "contactManualAdd", Component: "view/contact/manualAdd/index.vue", Sort: 3,
		Meta: sysModel.Meta{Title: "手动添加", Icon: "plus"},
	})
	if err != nil {
		return nil, err
	}
	listMenu, err := ensureMenu(sysModel.SysBaseMenu{
		MenuLevel: 1, Hidden: false, ParentId: parent.ID,
		Path: "myList", Name: "contactMyList", Component: "view/contact/lists/index.vue", Sort: 4,
		Meta: sysModel.Meta{Title: "我的列表", Icon: "folder"},
	})
	if err != nil {
		return nil, err
	}
	verifyMenu, err := ensureMenu(sysModel.SysBaseMenu{
		MenuLevel: 2, Hidden: false, ParentId: tools.ID,
		Path: "emailVerify", Name: "contactEmailVerify", Component: "view/contact/tools/emailVerify/index.vue", Sort: 1,
		Meta: sysModel.Meta{Title: "邮箱验证", Icon: "message"},
	})
	if err != nil {
		return nil, err
	}

	return []uint{parent.ID, tools.ID, importMenu.ID, historyMenu.ID, manualAddMenu.ID, listMenu.ID, verifyMenu.ID}, nil
}

func ensureMenu(in sysModel.SysBaseMenu) (sysModel.SysBaseMenu, error) {
	var m sysModel.SysBaseMenu
	err := global.GVA_DB.Where("name = ?", in.Name).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := global.GVA_DB.Create(&in).Error; err != nil {
			return sysModel.SysBaseMenu{}, err
		}
		return in, nil
	}
	if err != nil {
		return sysModel.SysBaseMenu{}, err
	}
	if in.ParentId != 0 && (m.ParentId == 0 || m.ParentId != in.ParentId) {
		_ = global.GVA_DB.Model(&sysModel.SysBaseMenu{}).Where("id = ?", m.ID).Update("parent_id", in.ParentId).Error
	}
	return m, nil
}

func ensureContactCasbinPolicies(authorityID string) error {
	policies := []adapter.CasbinRule{
		{Ptype: "p", V0: authorityID, V1: "/contact/list", V2: "GET"},
		{Ptype: "p", V0: authorityID, V1: "/contact/create", V2: "POST"},

		{Ptype: "p", V0: authorityID, V1: "/contactList/tree", V2: "GET"},
		{Ptype: "p", V0: authorityID, V1: "/contactList/create", V2: "POST"},
		{Ptype: "p", V0: authorityID, V1: "/contactList/update", V2: "PUT"},
		{Ptype: "p", V0: authorityID, V1: "/contactList/delete", V2: "DELETE"},

		{Ptype: "p", V0: authorityID, V1: "/contactImport/upload", V2: "POST"},
		{Ptype: "p", V0: authorityID, V1: "/contactImport/googleSheet", V2: "POST"},
		{Ptype: "p", V0: authorityID, V1: "/contactImport/suggestMapping", V2: "POST"},
		{Ptype: "p", V0: authorityID, V1: "/contactImport/validate", V2: "POST"},
		{Ptype: "p", V0: authorityID, V1: "/contactImport/start", V2: "POST"},
		{Ptype: "p", V0: authorityID, V1: "/contactImport/job", V2: "GET"},
		{Ptype: "p", V0: authorityID, V1: "/contactImport/history", V2: "GET"},
		{Ptype: "p", V0: authorityID, V1: "/contactImport/errors", V2: "GET"},
		{Ptype: "p", V0: authorityID, V1: "/contactImport/delete", V2: "POST"},
		{Ptype: "p", V0: authorityID, V1: "/contactImport/exportFailed", V2: "POST"},
		{Ptype: "p", V0: authorityID, V1: "/contactImport/downloadByToken", V2: "GET"},
		{Ptype: "p", V0: authorityID, V1: "/contactImport/template", V2: "GET"},

		{Ptype: "p", V0: authorityID, V1: "/emailVerify/start", V2: "POST"},
		{Ptype: "p", V0: authorityID, V1: "/emailVerify/job", V2: "GET"},
		{Ptype: "p", V0: authorityID, V1: "/emailVerify/history", V2: "GET"},
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

