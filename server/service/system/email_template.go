package system

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"mime/multipart"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	commonReq "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	sysModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	sysReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	sysResp "github.com/flipped-aurora/gin-vue-admin/server/model/system/response"
	"gorm.io/gorm"
)

type EmailTemplateService struct{}

var placeholderRe = regexp.MustCompile(`\{\{\s*([^}]+?)\s*\}\}`)
var htmlTagRe = regexp.MustCompile(`<[^>]+>`)

func normalizeTemplateStatus(s string) string {
	v := strings.ToLower(strings.TrimSpace(s))
	switch v {
	case "enabled", "disabled":
		return v
	default:
		return "enabled"
	}
}

func renderTemplate(text string, vars map[string]string) string {
	if text == "" {
		return ""
	}
	if len(vars) == 0 {
		return text
	}
	return placeholderRe.ReplaceAllStringFunc(text, func(m string) string {
		sub := placeholderRe.FindStringSubmatch(m)
		if len(sub) != 2 {
			return m
		}
		key := strings.TrimSpace(sub[1])
		if key == "" {
			return m
		}
		if v, ok := vars[key]; ok {
			return v
		}
		return m
	})
}

func htmlToText(s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	noTags := htmlTagRe.ReplaceAllString(s, " ")
	noTags = html.UnescapeString(noTags)
	noTags = strings.ReplaceAll(noTags, "\u00a0", " ")
	noTags = strings.Join(strings.Fields(noTags), " ")
	return noTags
}

func systemVars(user sysModel.SysUser) map[string]string {
	now := time.Now()
	return map[string]string{
		"当前日期": now.Format("2006-01-02"),
		"当前时间": now.Format("15:04:05"),
		"用户姓名": strings.TrimSpace(user.NickName),
		"用户邮箱": strings.TrimSpace(user.Email),
		"用户公司": "",
		"退订链接": "{{退订链接}}",
	}
}

func mergeVars(a, b map[string]string) map[string]string {
	if len(a) == 0 && len(b) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

func (s *EmailTemplateService) Create(userID uint, req sysReq.EmailTemplateCreateReq) (sysModel.SysEmailTemplate, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Subject = strings.TrimSpace(req.Subject)
	if req.Name == "" {
		return sysModel.SysEmailTemplate{}, errors.New("name required")
	}
	if req.Subject == "" {
		return sysModel.SysEmailTemplate{}, errors.New("subject required")
	}
	if strings.TrimSpace(req.BodyHTML) == "" {
		return sysModel.SysEmailTemplate{}, errors.New("body required")
	}
	if err := ensureFolderExists(global.GVA_DB, userID, req.FolderID); err != nil {
		return sysModel.SysEmailTemplate{}, err
	}
	tpl := sysModel.SysEmailTemplate{
		SysUserID: userID,
		FolderID:  req.FolderID,
		Name:      req.Name,
		Subject:   req.Subject,
		BodyHTML:  req.BodyHTML,
		BodyText:  req.BodyText,
		Status:    normalizeTemplateStatus(req.Status),
	}
	if strings.TrimSpace(tpl.BodyText) == "" {
		tpl.BodyText = htmlToText(tpl.BodyHTML)
	}
	if err := global.GVA_DB.Create(&tpl).Error; err != nil {
		return sysModel.SysEmailTemplate{}, err
	}
	return tpl, nil
}

func (s *EmailTemplateService) Update(userID uint, req sysReq.EmailTemplateUpdateReq) error {
	if req.ID == 0 {
		return errors.New("id required")
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Subject = strings.TrimSpace(req.Subject)
	if req.Name == "" {
		return errors.New("name required")
	}
	if req.Subject == "" {
		return errors.New("subject required")
	}
	if strings.TrimSpace(req.BodyHTML) == "" {
		return errors.New("body required")
	}
	if err := ensureFolderExists(global.GVA_DB, userID, req.FolderID); err != nil {
		return err
	}
	bodyText := strings.TrimSpace(req.BodyText)
	if bodyText == "" {
		bodyText = htmlToText(req.BodyHTML)
	}
	return global.GVA_DB.Model(&sysModel.SysEmailTemplate{}).Where("id = ? AND sys_user_id = ?", req.ID, userID).Updates(map[string]any{
		"folder_id": req.FolderID,
		"name":      req.Name,
		"subject":   req.Subject,
		"body_html": req.BodyHTML,
		"body_text": bodyText,
		"status":    normalizeTemplateStatus(req.Status),
	}).Error
}

func (s *EmailTemplateService) Delete(userID uint, id uint) error {
	if id == 0 {
		return errors.New("id required")
	}
	return global.GVA_DB.Delete(&sysModel.SysEmailTemplate{}, "id = ? AND sys_user_id = ?", id, userID).Error
}

func (s *EmailTemplateService) DeleteByIDs(userID uint, ids commonReq.IdsReq) error {
	if len(ids.Ids) == 0 {
		return nil
	}
	return global.GVA_DB.Delete(&[]sysModel.SysEmailTemplate{}, "sys_user_id = ? AND id in ?", userID, ids.Ids).Error
}

func (s *EmailTemplateService) GetByID(userID uint, id uint) (sysModel.SysEmailTemplate, error) {
	var tpl sysModel.SysEmailTemplate
	err := global.GVA_DB.Where("id = ? AND sys_user_id = ?", id, userID).First(&tpl).Error
	return tpl, err
}

func (s *EmailTemplateService) Copy(userID uint, req sysReq.EmailTemplateCopyReq) (sysModel.SysEmailTemplate, error) {
	if req.ID == 0 {
		return sysModel.SysEmailTemplate{}, errors.New("id required")
	}
	var tpl sysModel.SysEmailTemplate
	if err := global.GVA_DB.Where("id = ? AND sys_user_id = ?", req.ID, userID).First(&tpl).Error; err != nil {
		return sysModel.SysEmailTemplate{}, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = tpl.Name + " - 副本"
	}
	newTpl := sysModel.SysEmailTemplate{
		SysUserID: userID,
		FolderID:  tpl.FolderID,
		Name:      name,
		Subject:   tpl.Subject,
		BodyHTML:  tpl.BodyHTML,
		BodyText:  tpl.BodyText,
		Status:    tpl.Status,
	}
	if err := global.GVA_DB.Create(&newTpl).Error; err != nil {
		return sysModel.SysEmailTemplate{}, err
	}
	return newTpl, nil
}

func (s *EmailTemplateService) BatchStatus(userID uint, req sysReq.EmailTemplateBatchStatusReq) error {
	if len(req.IDs) == 0 {
		return nil
	}
	status := normalizeTemplateStatus(req.Status)
	return global.GVA_DB.Model(&sysModel.SysEmailTemplate{}).Where("sys_user_id = ? AND id in ?", userID, req.IDs).Update("status", status).Error
}

func (s *EmailTemplateService) Move(userID uint, req sysReq.EmailTemplateMoveReq) error {
	if req.ID == 0 {
		return errors.New("id required")
	}
	if err := ensureFolderExists(global.GVA_DB, userID, req.FolderID); err != nil {
		return err
	}
	return global.GVA_DB.Model(&sysModel.SysEmailTemplate{}).Where("id = ? AND sys_user_id = ?", req.ID, userID).Update("folder_id", req.FolderID).Error
}

func (s *EmailTemplateService) List(userID uint, req sysReq.EmailTemplateSearchReq) ([]sysResp.EmailTemplateItem, int64, error) {
	limit := req.PageSize
	offset := req.PageSize * (req.Page - 1)
	db := global.GVA_DB.Model(&sysModel.SysEmailTemplate{}).Where("sys_user_id = ?", userID)
	if v := strings.TrimSpace(req.Keyword); v != "" {
		like := "%" + v + "%"
		db = db.Where("name LIKE ? OR subject LIKE ?", like, like)
	}
	if v := strings.TrimSpace(req.Status); v != "" {
		db = db.Where("status = ?", normalizeTemplateStatus(v))
	}
	if req.FolderID != nil {
		db = db.Where("folder_id = ?", *req.FolderID)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := "updated_at desc"
	if strings.TrimSpace(req.SortBy) != "" {
		col := strings.ToLower(strings.TrimSpace(req.SortBy))
		allowed := map[string]string{
			"updatedat":  "updated_at",
			"usagecount": "usage_count",
			"replyrate":  "reply_rate",
			"createdat":  "created_at",
		}
		if c, ok := allowed[col]; ok {
			if req.SortDesc {
				order = c + " desc"
			} else {
				order = c + " asc"
			}
		}
	}

	var list []sysModel.SysEmailTemplate
	if limit != 0 {
		db = db.Limit(limit).Offset(offset)
	}
	if err := db.Order(order).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	items := make([]sysResp.EmailTemplateItem, 0, len(list))
	for _, t := range list {
		items = append(items, sysResp.EmailTemplateItem{
			ID:         t.ID,
			FolderID:   t.FolderID,
			Name:       t.Name,
			Subject:    t.Subject,
			BodyHTML:   t.BodyHTML,
			Status:     t.Status,
			UsageCount: t.UsageCount,
			OpenRate:   t.OpenRate,
			ReplyRate:  t.ReplyRate,
			UpdatedAt:  t.UpdatedAt,
			LastUsedAt: t.LastUsedAt,
		})
	}
	return items, total, nil
}

func (s *EmailTemplateService) Preview(userID uint, req sysReq.EmailTemplatePreviewReq) (sysResp.EmailTemplatePreviewResp, error) {
	var user sysModel.SysUser
	_ = global.GVA_DB.Where("id = ?", userID).First(&user).Error
	vars := mergeVars(systemVars(user), req.Vars)
	subject := renderTemplate(req.Subject, vars)
	bodyHtml := renderTemplate(req.BodyHTML, vars)
	bodyText := htmlToText(bodyHtml)
	return sysResp.EmailTemplatePreviewResp{Subject: subject, BodyHTML: bodyHtml, BodyText: bodyText}, nil
}

func (s *EmailTemplateService) TestSend(userID uint, req sysReq.EmailTemplateTestSendReq) error {
	if req.TemplateID == 0 {
		return errors.New("templateId required")
	}
	if len(req.ToEmails) == 0 {
		return errors.New("toEmails required")
	}
	var tpl sysModel.SysEmailTemplate
	if err := global.GVA_DB.Where("id = ? AND sys_user_id = ?", req.TemplateID, userID).First(&tpl).Error; err != nil {
		return err
	}
	var user sysModel.SysUser
	_ = global.GVA_DB.Where("id = ?", userID).First(&user).Error
	vars := mergeVars(systemVars(user), req.Vars)
	subject := renderTemplate(tpl.Subject, vars)
	bodyHtml := renderTemplate(tpl.BodyHTML, vars)
	bodyText := tpl.BodyText
	if strings.TrimSpace(bodyText) == "" {
		bodyText = htmlToText(bodyHtml)
	} else {
		bodyText = renderTemplate(bodyText, vars)
	}

	senderSvc := SenderEmailAccountService{}
	if err := senderSvc.SendEmail(userID, req.SenderAccountID, req.ToEmails, subject, bodyHtml, bodyText); err != nil {
		return err
	}

	now := time.Now()
	return global.GVA_DB.Model(&sysModel.SysEmailTemplate{}).Where("id = ? AND sys_user_id = ?", tpl.ID, userID).Updates(map[string]any{
		"usage_count":  gorm.Expr("usage_count + ?", 1),
		"last_used_at": &now,
	}).Error
}

type exportedTemplate struct {
	Name     string `json:"name"`
	Subject  string `json:"subject"`
	BodyHTML string `json:"bodyHtml"`
	BodyText string `json:"bodyText"`
	Status   string `json:"status"`
	FolderID *uint  `json:"folderId"`
}

func (s *EmailTemplateService) Export(userID uint, req sysReq.EmailTemplateExportReq) ([]byte, string, string, error) {
	if len(req.IDs) == 0 {
		return nil, "", "", errors.New("ids required")
	}
	format := strings.ToLower(strings.TrimSpace(req.Format))
	if format == "" {
		format = "json"
	}
	if format != "json" && format != "html" {
		return nil, "", "", errors.New("invalid format")
	}

	var tpls []sysModel.SysEmailTemplate
	if err := global.GVA_DB.Where("sys_user_id = ? AND id in ?", userID, req.IDs).Find(&tpls).Error; err != nil {
		return nil, "", "", err
	}
	if len(tpls) == 0 {
		return nil, "", "", errors.New("not found")
	}

	if format == "html" {
		combined := make([]string, 0, len(tpls))
		for _, t := range tpls {
			combined = append(combined, t.BodyHTML)
		}
		content := strings.Join(combined, "\n")
		name := fmt.Sprintf("email_templates_%s.html", time.Now().Format("20060102150405"))
		return []byte(content), "text/html; charset=utf-8", name, nil
	}

	out := make([]exportedTemplate, 0, len(tpls))
	for _, t := range tpls {
		out = append(out, exportedTemplate{
			Name:     t.Name,
			Subject:  t.Subject,
			BodyHTML: t.BodyHTML,
			BodyText: t.BodyText,
			Status:   t.Status,
			FolderID: t.FolderID,
		})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, "", "", err
	}
	name := fmt.Sprintf("email_templates_%s.json", time.Now().Format("20060102150405"))
	return b, "application/json; charset=utf-8", name, nil
}

func (s *EmailTemplateService) Import(userID uint, fileHeader *multipart.FileHeader, req sysReq.EmailTemplateImportReq) (int, error) {
	if fileHeader == nil {
		return 0, errors.New("file required")
	}
	nameLower := strings.ToLower(strings.TrimSpace(fileHeader.Filename))
	format := ""
	if strings.HasSuffix(nameLower, ".json") {
		format = "json"
	} else if strings.HasSuffix(nameLower, ".html") || strings.HasSuffix(nameLower, ".htm") {
		format = "html"
	}
	if format == "" {
		return 0, errors.New("only .json/.html supported")
	}

	if err := ensureFolderExists(global.GVA_DB, userID, req.FolderID); err != nil {
		return 0, err
	}

	f, err := fileHeader.Open()
	if err != nil {
		return 0, err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return 0, err
	}

	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		mode = "create"
	}

	if format == "html" {
		name := strings.TrimSuffix(fileHeader.Filename, ".html")
		if name == "" {
			name = "导入模板"
		}
		_, err := s.Create(userID, sysReq.EmailTemplateCreateReq{
			FolderID: req.FolderID,
			Name:     name,
			Subject:  name,
			BodyHTML: string(data),
			BodyText: "",
			Status:   "enabled",
		})
		if err != nil {
			return 0, err
		}
		return 1, nil
	}

	var imported []exportedTemplate
	if err := json.Unmarshal(data, &imported); err != nil {
		return 0, err
	}
	if len(imported) == 0 {
		return 0, nil
	}

	created := 0
	for _, t := range imported {
		folderID := req.FolderID
		if folderID == nil {
			folderID = t.FolderID
		}
		name := strings.TrimSpace(t.Name)
		if name == "" {
			continue
		}
		if mode == "override" {
			var existing sysModel.SysEmailTemplate
			err := global.GVA_DB.Where("sys_user_id = ? AND name = ?", userID, name).First(&existing).Error
			if err == nil {
				_ = s.Update(userID, sysReq.EmailTemplateUpdateReq{
					ID:       existing.ID,
					FolderID: folderID,
					Name:     name,
					Subject:  strings.TrimSpace(t.Subject),
					BodyHTML: t.BodyHTML,
					BodyText: t.BodyText,
					Status:   normalizeTemplateStatus(t.Status),
				})
				created++
				continue
			}
		}
		_, err := s.Create(userID, sysReq.EmailTemplateCreateReq{
			FolderID: folderID,
			Name:     name,
			Subject:  strings.TrimSpace(t.Subject),
			BodyHTML: t.BodyHTML,
			BodyText: t.BodyText,
			Status:   normalizeTemplateStatus(t.Status),
		})
		if err == nil {
			created++
		}
	}
	return created, nil
}
