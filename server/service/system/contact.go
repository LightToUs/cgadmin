package system

import (
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	sysModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	sysReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	sysResp "github.com/flipped-aurora/gin-vue-admin/server/model/system/response"
	"gorm.io/gorm"
)

type ContactService struct{}

var emailRe = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)

func normalizeContactStatus(s string) string {
	v := strings.ToLower(strings.TrimSpace(s))
	switch v {
	case "uncontacted", "contacted", "replied":
		return v
	default:
		return "uncontacted"
	}
}

func normalizeVerifyStatus(s string) string {
	v := strings.ToLower(strings.TrimSpace(s))
	switch v {
	case "unverified", "valid", "risk", "invalid":
		return v
	default:
		return "unverified"
	}
}

func normalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func isEmailValidFormat(s string) bool {
	return emailRe.MatchString(strings.TrimSpace(s))
}

func (s *ContactService) Create(userID uint, req sysReq.ContactCreateReq) (sysModel.SysContact, error) {
	company := strings.TrimSpace(req.CompanyName)
	email := normalizeEmail(req.Email)
	if company == "" {
		return sysModel.SysContact{}, errors.New("companyName required")
	}
	if email == "" || !isEmailValidFormat(email) {
		return sysModel.SysContact{}, errors.New("email invalid")
	}
	c := sysModel.SysContact{
		SysUserID:         userID,
		CompanyName:       company,
		Website:           strings.TrimSpace(req.Website),
		ContactName:       strings.TrimSpace(req.ContactName),
		Title:             strings.TrimSpace(req.Title),
		Email:             email,
		Phone:             strings.TrimSpace(req.Phone),
		Country:           strings.TrimSpace(req.Country),
		Status:            normalizeContactStatus(req.Status),
		EmailVerifyStatus: "unverified",
		Tags:              strings.TrimSpace(req.Tags),
	}
	if err := global.GVA_DB.Create(&c).Error; err != nil {
		return sysModel.SysContact{}, err
	}
	return c, nil
}

func (s *ContactService) UpsertByEmail(tx *gorm.DB, userID uint, in sysModel.SysContact) (sysModel.SysContact, bool, error) {
	email := normalizeEmail(in.Email)
	if email == "" || !isEmailValidFormat(email) {
		return sysModel.SysContact{}, false, errors.New("email invalid")
	}
	var existing sysModel.SysContact
	err := tx.Where("sys_user_id = ? AND email = ?", userID, email).First(&existing).Error
	if err == nil {
		update := map[string]any{
			"company_name": in.CompanyName,
			"website":      in.Website,
			"contact_name": in.ContactName,
			"title":        in.Title,
			"phone":        in.Phone,
			"country":      in.Country,
			"status":       normalizeContactStatus(in.Status),
			"tags":         in.Tags,
		}
		if err := tx.Model(&sysModel.SysContact{}).Where("id = ? AND sys_user_id = ?", existing.ID, userID).Updates(update).Error; err != nil {
			return sysModel.SysContact{}, false, err
		}
		existing.CompanyName = in.CompanyName
		existing.Website = in.Website
		existing.ContactName = in.ContactName
		existing.Title = in.Title
		existing.Phone = in.Phone
		existing.Country = in.Country
		existing.Status = normalizeContactStatus(in.Status)
		existing.Tags = in.Tags
		return existing, true, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return sysModel.SysContact{}, false, err
	}
	in.SysUserID = userID
	in.Email = email
	in.Status = normalizeContactStatus(in.Status)
	in.EmailVerifyStatus = normalizeVerifyStatus(in.EmailVerifyStatus)
	if err := tx.Create(&in).Error; err != nil {
		return sysModel.SysContact{}, false, err
	}
	return in, false, nil
}

func (s *ContactService) List(userID uint, req sysReq.ContactListReq) ([]sysResp.ContactItem, int64, error) {
	limit := req.PageSize
	offset := req.PageSize * (req.Page - 1)
	db := global.GVA_DB.Model(&sysModel.SysContact{}).Where("sys_user_id = ?", userID)

	if strings.TrimSpace(req.Keyword) != "" {
		like := "%" + strings.TrimSpace(req.Keyword) + "%"
		db = db.Where("company_name LIKE ? OR contact_name LIKE ? OR email LIKE ?", like, like, like)
	}
	if strings.TrimSpace(req.Country) != "" {
		db = db.Where("country = ?", strings.TrimSpace(req.Country))
	}
	if strings.TrimSpace(req.Status) != "" {
		db = db.Where("status = ?", normalizeContactStatus(req.Status))
	}
	if strings.TrimSpace(req.Verified) != "" {
		db = db.Where("email_verify_status = ?", normalizeVerifyStatus(req.Verified))
	}

	if strings.TrimSpace(req.ListID) != "" && req.ListID != "all" {
		if id, err := strconv.ParseUint(strings.TrimSpace(req.ListID), 10, 64); err == nil && id > 0 {
			var list sysModel.SysContactList
			if err := global.GVA_DB.Where("id = ? AND sys_user_id = ?", uint(id), userID).First(&list).Error; err == nil {
				switch list.Type {
				case "custom":
					db = db.Joins("JOIN sys_contact_list_items li ON li.contact_id = sys_contacts.id AND li.sys_user_id = sys_contacts.sys_user_id").
						Where("li.contact_list_id = ?", list.ID)
				case "smart":
					var rule struct {
						Type string `json:"type"`
					}
					_ = json.Unmarshal([]byte(list.Rule), &rule)
					switch rule.Type {
					case "verified":
						db = db.Where("email_verify_status in ?", []string{"valid", "risk"})
					case "replied":
						db = db.Where("status = ?", "replied")
					case "newThisWeek":
						cutoff := time.Now().Add(-7 * 24 * time.Hour)
						db = db.Where("created_at >= ?", cutoff)
					}
				}
			}
		}
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := "updated_at desc"
	if strings.TrimSpace(req.SortBy) != "" {
		col := strings.ToLower(strings.TrimSpace(req.SortBy))
		allowed := map[string]string{
			"updatedat": "updated_at",
			"createdat": "created_at",
			"company":   "company_name",
			"country":   "country",
			"email":     "email",
		}
		if c, ok := allowed[col]; ok {
			if req.SortDesc {
				order = c + " desc"
			} else {
				order = c + " asc"
			}
		}
	}

	var list []sysModel.SysContact
	if limit != 0 {
		db = db.Limit(limit).Offset(offset)
	}
	if err := db.Order(order).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	items := make([]sysResp.ContactItem, 0, len(list))
	for _, c := range list {
		items = append(items, sysResp.ContactItem{
			ID:                c.ID,
			CompanyName:       c.CompanyName,
			Website:           c.Website,
			ContactName:       c.ContactName,
			Title:             c.Title,
			Email:             c.Email,
			Phone:             c.Phone,
			Country:           c.Country,
			Status:            c.Status,
			EmailVerifyStatus: c.EmailVerifyStatus,
			UpdatedAt:         c.UpdatedAt,
		})
	}
	return items, total, nil
}

func (s *ContactService) VerifyEmailBasic(email string) (string, bool) {
	email = normalizeEmail(email)
	if email == "" {
		return "invalid", false
	}
	if !isEmailValidFormat(email) {
		return "invalid", false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "invalid", false
	}
	local := parts[0]
	if local == "" {
		return "invalid", false
	}
	roleLocal := map[string]bool{
		"info": true, "sales": true, "support": true, "admin": true, "contact": true, "service": true,
	}
	if roleLocal[local] {
		return "risk", true
	}
	return "valid", true
}

func (s *ContactService) UpdateVerifyStatus(tx *gorm.DB, userID uint, contactID uint, status string) error {
	status = normalizeVerifyStatus(status)
	var t *time.Time
	if status == "valid" || status == "risk" || status == "invalid" {
		now := time.Now()
		t = &now
	}
	return tx.Model(&sysModel.SysContact{}).Where("id = ? AND sys_user_id = ?", contactID, userID).Updates(map[string]any{
		"email_verify_status": status,
		"email_verified_at":   t,
	}).Error
}

func stableColumnOrder(cols []string) []string {
	out := append([]string{}, cols...)
	sort.Strings(out)
	return out
}
