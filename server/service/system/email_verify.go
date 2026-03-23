package system

import (
	"errors"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	sysModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	sysReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"gorm.io/gorm"
)

type EmailVerifyService struct{}

func (s *EmailVerifyService) Start(userID uint, req sysReq.EmailVerifyStartReq) (sysModel.SysEmailVerifyJob, error) {
	scope := strings.ToLower(strings.TrimSpace(req.ScopeType))
	if scope == "" {
		scope = "allUnverified"
	}
	method := strings.ToLower(strings.TrimSpace(req.Method))
	if method == "" {
		method = "basic"
	}
	job := sysModel.SysEmailVerifyJob{
		SysUserID: userID,
		ScopeType: scope,
		ListID:    req.ListID,
		Method:    method,
		Status:    "running",
		Progress:  0,
	}
	if err := global.GVA_DB.Create(&job).Error; err != nil {
		return sysModel.SysEmailVerifyJob{}, err
	}
	go s.run(job.ID, userID)
	return job, nil
}

func (s *EmailVerifyService) GetJob(userID uint, id uint) (sysModel.SysEmailVerifyJob, error) {
	var job sysModel.SysEmailVerifyJob
	err := global.GVA_DB.Where("id = ? AND sys_user_id = ?", id, userID).First(&job).Error
	return job, err
}

func (s *EmailVerifyService) List(userID uint, page int, pageSize int) ([]sysModel.SysEmailVerifyJob, int64, error) {
	limit := pageSize
	offset := pageSize * (page - 1)
	db := global.GVA_DB.Model(&sysModel.SysEmailVerifyJob{}).Where("sys_user_id = ?", userID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []sysModel.SysEmailVerifyJob
	if limit != 0 {
		db = db.Limit(limit).Offset(offset)
	}
	if err := db.Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *EmailVerifyService) run(jobID uint, userID uint) {
	contactSvc := &ContactService{}
	job, err := s.GetJob(userID, jobID)
	if err != nil {
		return
	}

	var contacts []sysModel.SysContact
	db := global.GVA_DB.Model(&sysModel.SysContact{}).Where("sys_user_id = ?", userID)
	if job.ScopeType == "allunverified" {
		db = db.Where("email_verify_status = ?", "unverified")
	}
	if job.ScopeType == "currentlist" && job.ListID != nil && *job.ListID != 0 {
		db = db.Joins("JOIN sys_contact_list_items li ON li.contact_id = sys_contacts.id AND li.sys_user_id = sys_contacts.sys_user_id").
			Where("li.contact_list_id = ?", *job.ListID)
	}
	if err := db.Find(&contacts).Error; err != nil {
		_ = s.fail(jobID, userID, err.Error())
		return
	}

	total := len(contacts)
	valid := 0
	risk := 0
	invalid := 0

	for i, c := range contacts {
		status, ok := contactSvc.VerifyEmailBasic(c.Email)
		if !ok {
			invalid++
		} else if status == "risk" {
			risk++
		} else {
			valid++
		}
		_ = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
			return contactSvc.UpdateVerifyStatus(tx, userID, c.ID, status)
		})

		if i%50 == 0 || i == total-1 {
			progress := int(float64(i+1) / float64(max(total, 1)) * 100)
			_ = global.GVA_DB.Model(&sysModel.SysEmailVerifyJob{}).Where("id = ? AND sys_user_id = ?", jobID, userID).Updates(map[string]any{
				"progress": progress,
				"total":    total,
				"valid":    valid,
				"risk":     risk,
				"invalid":  invalid,
			}).Error
		}
	}
	now := time.Now()
	_ = global.GVA_DB.Model(&sysModel.SysEmailVerifyJob{}).Where("id = ? AND sys_user_id = ?", jobID, userID).Updates(map[string]any{
		"status":     "finished",
		"progress":   100,
		"total":      total,
		"valid":      valid,
		"risk":       risk,
		"invalid":    invalid,
		"finished_at": &now,
	}).Error
}

func (s *EmailVerifyService) fail(jobID uint, userID uint, msg string) error {
	now := time.Now()
	return global.GVA_DB.Model(&sysModel.SysEmailVerifyJob{}).Where("id = ? AND sys_user_id = ?", jobID, userID).Updates(map[string]any{
		"status":        "failed",
		"error_message": msg,
		"finished_at":   &now,
	}).Error
}

func (s *EmailVerifyService) ValidateStart(req sysReq.EmailVerifyStartReq) error {
	scope := strings.ToLower(strings.TrimSpace(req.ScopeType))
	if scope != "" && scope != "currentlist" && scope != "allunverified" {
		return errors.New("invalid scopeType")
	}
	method := strings.ToLower(strings.TrimSpace(req.Method))
	if method != "" && method != "basic" && method != "deep" {
		return errors.New("invalid method")
	}
	return nil
}

