package system

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	sysModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	sysReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	sysResp "github.com/flipped-aurora/gin-vue-admin/server/model/system/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/jordan-wright/email"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SenderEmailAccountService struct{}

func normalizeSecurity(security string) string {
	s := strings.ToLower(strings.TrimSpace(security))
	switch s {
	case "ssl", "tls", "none":
		return s
	default:
		return "ssl"
	}
}

func normalizeProvider(provider string) string {
	return strings.ToLower(strings.TrimSpace(provider))
}

func defaultDisplayName(emailAddr string) string {
	parts := strings.Split(emailAddr, "@")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return emailAddr
}

func validateEmail(emailAddr string) error {
	_, err := mail.ParseAddress(emailAddr)
	return err
}

func validatePort(port int) error {
	if port <= 0 || port > 65535 {
		return errors.New("invalid port")
	}
	return nil
}

func (s *SenderEmailAccountService) GetList(userID uint) (sysResp.SenderEmailAccountListResp, error) {
	var accounts []sysModel.SysSenderEmailAccount
	if err := global.GVA_DB.Where("sys_user_id = ?", userID).Order("is_default desc, id desc").Find(&accounts).Error; err != nil {
		return sysResp.SenderEmailAccountListResp{}, err
	}
	items := make([]sysResp.SenderEmailAccountItem, 0, len(accounts))
	for _, a := range accounts {
		status := a.Status
		if a.DailyLimit > 0 && a.TodaySent >= int(float64(a.DailyLimit)*0.8) && a.TodaySent < a.DailyLimit && status == "normal" {
			status = "quota_warning"
		}
		items = append(items, sysResp.SenderEmailAccountItem{
			ID:             a.ID,
			Email:          a.Email,
			DisplayName:    a.DisplayName,
			Provider:       a.Provider,
			SMTPHost:       a.SMTPHost,
			SMTPPort:       a.SMTPPort,
			Security:       a.Security,
			IsLoginAuth:    a.IsLoginAuth,
			IsDefault:      a.IsDefault,
			Status:         status,
			LastError:      a.LastError,
			LastCheckedAt:  a.LastCheckedAt,
			TodaySent:      a.TodaySent,
			DailyLimit:     a.DailyLimit,
			IntervalPolicy: a.IntervalPolicy,
		})
	}
	return sysResp.SenderEmailAccountListResp{List: items}, nil
}

func (s *SenderEmailAccountService) Create(userID uint, req sysReq.SenderEmailAccountCreateReq) (sysModel.SysSenderEmailAccount, error) {
	req.Email = strings.TrimSpace(req.Email)
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.Provider = normalizeProvider(req.Provider)
	req.SMTPHost = strings.TrimSpace(req.SMTPHost)
	req.Security = normalizeSecurity(req.Security)
	req.IntervalPolicy = strings.TrimSpace(req.IntervalPolicy)
	if req.IntervalPolicy == "" {
		req.IntervalPolicy = "1m"
	}
	if req.DailyLimit <= 0 {
		req.DailyLimit = 50
	}

	if err := validateEmail(req.Email); err != nil {
		return sysModel.SysSenderEmailAccount{}, errors.New("invalid email")
	}
	if err := validatePort(req.SMTPPort); err != nil {
		return sysModel.SysSenderEmailAccount{}, err
	}
	if req.SMTPHost == "" {
		return sysModel.SysSenderEmailAccount{}, errors.New("smtp host required")
	}
	if strings.TrimSpace(req.Secret) == "" {
		return sysModel.SysSenderEmailAccount{}, errors.New("secret required")
	}
	if req.DisplayName == "" {
		req.DisplayName = defaultDisplayName(req.Email)
	}

	if err := s.checkSMTPAuth(req.Email, req.Secret, req.SMTPHost, req.SMTPPort, req.Security, req.IsLoginAuth); err != nil {
		return sysModel.SysSenderEmailAccount{}, err
	}

	secretEnc, err := utils.EncryptStringAESGCM(req.Secret)
	if err != nil {
		return sysModel.SysSenderEmailAccount{}, err
	}

	account := sysModel.SysSenderEmailAccount{
		SysUserID:      userID,
		Email:          req.Email,
		DisplayName:    req.DisplayName,
		Provider:       req.Provider,
		SMTPHost:       req.SMTPHost,
		SMTPPort:       req.SMTPPort,
		Security:       req.Security,
		IsLoginAuth:    req.IsLoginAuth,
		SecretEnc:      secretEnc,
		Status:         "normal",
		LastError:      "",
		LastCheckedAt:  time.Now(),
		DailyLimit:     req.DailyLimit,
		IntervalPolicy: req.IntervalPolicy,
	}

	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var existing sysModel.SysSenderEmailAccount
		if err := tx.Where("sys_user_id = ? AND email = ?", userID, req.Email).First(&existing).Error; err == nil {
			return errors.New("email already exists")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var count int64
		if err := tx.Model(&sysModel.SysSenderEmailAccount{}).Where("sys_user_id = ?", userID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			account.IsDefault = true
		}
		return tx.Create(&account).Error
	})
	if err != nil {
		return sysModel.SysSenderEmailAccount{}, err
	}
	return account, nil
}

func (s *SenderEmailAccountService) Update(userID uint, req sysReq.SenderEmailAccountUpdateReq) error {
	req.Email = strings.TrimSpace(req.Email)
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.Provider = normalizeProvider(req.Provider)
	req.SMTPHost = strings.TrimSpace(req.SMTPHost)
	req.Security = normalizeSecurity(req.Security)
	req.IntervalPolicy = strings.TrimSpace(req.IntervalPolicy)
	if req.IntervalPolicy == "" {
		req.IntervalPolicy = "1m"
	}
	if req.DailyLimit <= 0 {
		req.DailyLimit = 50
	}

	var existing sysModel.SysSenderEmailAccount
	if err := global.GVA_DB.Where("id = ? AND sys_user_id = ?", req.ID, userID).First(&existing).Error; err != nil {
		return err
	}
	if req.Email == "" {
		req.Email = existing.Email
	}
	if err := validateEmail(req.Email); err != nil {
		return errors.New("invalid email")
	}
	if req.DisplayName == "" {
		req.DisplayName = defaultDisplayName(req.Email)
	}
	if req.SMTPHost == "" {
		req.SMTPHost = existing.SMTPHost
	}
	if req.SMTPPort == 0 {
		req.SMTPPort = existing.SMTPPort
	}
	if err := validatePort(req.SMTPPort); err != nil {
		return err
	}
	secretPlain := strings.TrimSpace(req.Secret)
	if secretPlain == "" {
		var err error
		secretPlain, err = utils.DecryptStringAESGCM(existing.SecretEnc)
		if err != nil {
			return err
		}
	}

	if err := s.checkSMTPAuth(req.Email, secretPlain, req.SMTPHost, req.SMTPPort, req.Security, req.IsLoginAuth); err != nil {
		return err
	}

	updates := map[string]any{
		"email":           req.Email,
		"display_name":    req.DisplayName,
		"provider":        req.Provider,
		"smtp_host":       req.SMTPHost,
		"smtp_port":       req.SMTPPort,
		"security":        req.Security,
		"is_login_auth":   req.IsLoginAuth,
		"daily_limit":     req.DailyLimit,
		"interval_policy": req.IntervalPolicy,
		"status":          "normal",
		"last_error":      "",
		"last_checked_at": time.Now(),
	}
	if strings.TrimSpace(req.Secret) != "" {
		secretEnc, err := utils.EncryptStringAESGCM(req.Secret)
		if err != nil {
			return err
		}
		updates["secret_enc"] = secretEnc
	}
	return global.GVA_DB.Model(&sysModel.SysSenderEmailAccount{}).Where("id = ? AND sys_user_id = ?", req.ID, userID).Updates(updates).Error
}

func (s *SenderEmailAccountService) Delete(userID uint, id uint) error {
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var account sysModel.SysSenderEmailAccount
		if err := tx.Where("id = ? AND sys_user_id = ?", id, userID).First(&account).Error; err != nil {
			return err
		}
		if err := tx.Delete(&sysModel.SysSenderEmailAccount{}, "id = ? AND sys_user_id = ?", id, userID).Error; err != nil {
			return err
		}
		if account.IsDefault {
			var next sysModel.SysSenderEmailAccount
			if err := tx.Where("sys_user_id = ?", userID).Order("id desc").First(&next).Error; err == nil {
				return tx.Model(&sysModel.SysSenderEmailAccount{}).Where("id = ?", next.ID).Update("is_default", true).Error
			}
		}
		return nil
	})
}

func (s *SenderEmailAccountService) SetDefault(userID uint, id uint) error {
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var account sysModel.SysSenderEmailAccount
		if err := tx.Where("id = ? AND sys_user_id = ?", id, userID).First(&account).Error; err != nil {
			return err
		}
		if err := tx.Model(&sysModel.SysSenderEmailAccount{}).Where("sys_user_id = ?", userID).Update("is_default", false).Error; err != nil {
			return err
		}
		return tx.Model(&sysModel.SysSenderEmailAccount{}).Where("id = ? AND sys_user_id = ?", id, userID).Update("is_default", true).Error
	})
}

func (s *SenderEmailAccountService) UpdateQuota(userID uint, req sysReq.SenderEmailAccountQuotaUpdateReq) error {
	req.IntervalPolicy = strings.TrimSpace(req.IntervalPolicy)
	if req.IntervalPolicy == "" {
		req.IntervalPolicy = "1m"
	}
	if req.DailyLimit <= 0 {
		req.DailyLimit = 50
	}
	return global.GVA_DB.Model(&sysModel.SysSenderEmailAccount{}).Where("id = ? AND sys_user_id = ?", req.ID, userID).Updates(map[string]any{
		"daily_limit":     req.DailyLimit,
		"interval_policy": req.IntervalPolicy,
	}).Error
}

func (s *SenderEmailAccountService) TestConnection(req sysReq.SenderEmailAccountTestReq) error {
	req.Email = strings.TrimSpace(req.Email)
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.SMTPHost = strings.TrimSpace(req.SMTPHost)
	req.Security = normalizeSecurity(req.Security)

	if err := validateEmail(req.Email); err != nil {
		return errors.New("invalid email")
	}
	if err := validatePort(req.SMTPPort); err != nil {
		return err
	}
	if strings.TrimSpace(req.Secret) == "" {
		return errors.New("secret required")
	}
	if req.DisplayName == "" {
		req.DisplayName = defaultDisplayName(req.Email)
	}

	if err := s.checkSMTPAuth(req.Email, req.Secret, req.SMTPHost, req.SMTPPort, req.Security, req.IsLoginAuth); err != nil {
		return err
	}
	return s.sendTestEmail(req.Email, req.DisplayName, req.Secret, req.SMTPHost, req.SMTPPort, req.Security, req.IsLoginAuth)
}

func (s *SenderEmailAccountService) TestConnectionByID(userID uint, id uint) error {
	var account sysModel.SysSenderEmailAccount
	if err := global.GVA_DB.Where("id = ? AND sys_user_id = ?", id, userID).First(&account).Error; err != nil {
		return err
	}
	secret, err := utils.DecryptStringAESGCM(account.SecretEnc)
	if err != nil {
		return err
	}
	if err := s.checkSMTPAuth(account.Email, secret, account.SMTPHost, account.SMTPPort, normalizeSecurity(account.Security), account.IsLoginAuth); err != nil {
		_ = global.GVA_DB.Model(&sysModel.SysSenderEmailAccount{}).Where("id = ?", account.ID).Updates(map[string]any{
			"status":          "error",
			"last_error":      err.Error(),
			"last_checked_at": time.Now(),
		}).Error
		return err
	}
	if err := s.sendTestEmail(account.Email, account.DisplayName, secret, account.SMTPHost, account.SMTPPort, normalizeSecurity(account.Security), account.IsLoginAuth); err != nil {
		_ = global.GVA_DB.Model(&sysModel.SysSenderEmailAccount{}).Where("id = ?", account.ID).Updates(map[string]any{
			"status":          "error",
			"last_error":      err.Error(),
			"last_checked_at": time.Now(),
		}).Error
		return err
	}
	_ = global.GVA_DB.Model(&sysModel.SysSenderEmailAccount{}).Where("id = ?", account.ID).Updates(map[string]any{
		"status":          "normal",
		"last_error":      "",
		"last_checked_at": time.Now(),
	}).Error
	return nil
}

func (s *SenderEmailAccountService) DailyResetAll() error {
	today := time.Now().Format("2006-01-02")
	return global.GVA_DB.Model(&sysModel.SysSenderEmailAccount{}).Where("today_sent_date <> ?", today).Updates(map[string]any{
		"today_sent":      0,
		"today_sent_date": today,
	}).Error
}

func (s *SenderEmailAccountService) ConnectivityCheckAll() error {
	var accounts []sysModel.SysSenderEmailAccount
	if err := global.GVA_DB.Find(&accounts).Error; err != nil {
		return err
	}
	for _, account := range accounts {
		secret, err := utils.DecryptStringAESGCM(account.SecretEnc)
		if err != nil {
			_ = global.GVA_DB.Model(&sysModel.SysSenderEmailAccount{}).Where("id = ?", account.ID).Updates(map[string]any{
				"status":          "error",
				"last_error":      err.Error(),
				"last_checked_at": time.Now(),
			}).Error
			continue
		}
		err = s.checkSMTPAuth(account.Email, secret, account.SMTPHost, account.SMTPPort, normalizeSecurity(account.Security), account.IsLoginAuth)
		if err != nil {
			_ = global.GVA_DB.Model(&sysModel.SysSenderEmailAccount{}).Where("id = ?", account.ID).Updates(map[string]any{
				"status":          "error",
				"last_error":      err.Error(),
				"last_checked_at": time.Now(),
			}).Error
			continue
		}
		_ = global.GVA_DB.Model(&sysModel.SysSenderEmailAccount{}).Where("id = ?", account.ID).Updates(map[string]any{
			"status":          "normal",
			"last_error":      "",
			"last_checked_at": time.Now(),
		}).Error
	}
	return nil
}

func (s *SenderEmailAccountService) IncrementTodaySent(accountID uint, delta int) error {
	if delta <= 0 {
		return nil
	}
	today := time.Now().Format("2006-01-02")
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var account sysModel.SysSenderEmailAccount
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", accountID).First(&account).Error; err != nil {
			return err
		}
		updates := map[string]any{}
		if account.TodaySentDate != today {
			updates["today_sent_date"] = today
			updates["today_sent"] = delta
		} else {
			updates["today_sent"] = account.TodaySent + delta
		}
		return tx.Model(&sysModel.SysSenderEmailAccount{}).Where("id = ?", accountID).Updates(updates).Error
	})
}

type loginAuth struct {
	username, password string
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			prompt := strings.ToLower(string(fromServer))
			if strings.Contains(prompt, "username") || strings.Contains(prompt, "user") {
				return []byte(a.username), nil
			}
			if strings.Contains(prompt, "password") || strings.Contains(prompt, "pass") {
				return []byte(a.password), nil
			}
		}
	}
	return nil, nil
}

func (s *SenderEmailAccountService) checkSMTPAuth(emailAddr, secret, host string, port int, security string, isLoginAuth bool) error {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	var c *smtp.Client
	var err error

	switch security {
	case "ssl":
		conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 10 * time.Second}, "tcp", addr, &tls.Config{ServerName: host})
		if err != nil {
			return err
		}
		c, err = smtp.NewClient(conn, host)
		if err != nil {
			_ = conn.Close()
			return err
		}
	case "tls", "none":
		c, err = smtp.Dial(addr)
		if err != nil {
			return err
		}
		_ = c.Hello("localhost")
		if security == "tls" {
			if ok, _ := c.Extension("STARTTLS"); !ok {
				_ = c.Quit()
				return errors.New("server does not support STARTTLS")
			}
			if err := c.StartTLS(&tls.Config{ServerName: host}); err != nil {
				_ = c.Quit()
				return err
			}
		}
	default:
		return errors.New("invalid security")
	}
	defer func() {
		_ = c.Quit()
	}()

	var auth smtp.Auth
	if isLoginAuth {
		auth = &loginAuth{username: emailAddr, password: secret}
	} else {
		auth = smtp.PlainAuth("", emailAddr, secret, host)
	}
	if ok, _ := c.Extension("AUTH"); !ok {
		return errors.New("server does not support AUTH")
	}
	return c.Auth(auth)
}

func (s *SenderEmailAccountService) sendTestEmail(emailAddr, displayName, secret, host string, port int, security string, isLoginAuth bool) error {
	var auth smtp.Auth
	if isLoginAuth {
		auth = &loginAuth{username: emailAddr, password: secret}
	} else {
		auth = smtp.PlainAuth("", emailAddr, secret, host)
	}

	e := email.NewEmail()
	if strings.TrimSpace(displayName) != "" {
		e.From = fmt.Sprintf("%s <%s>", displayName, emailAddr)
	} else {
		e.From = emailAddr
	}
	e.To = []string{emailAddr}
	e.Subject = "连接测试"
	e.HTML = []byte("连接成功，测试邮件已发送")
	hostAddr := fmt.Sprintf("%s:%d", host, port)
	if security == "ssl" {
		return e.SendWithTLS(hostAddr, auth, &tls.Config{ServerName: host})
	}
	return e.Send(hostAddr, auth)
}

func (s *SenderEmailAccountService) SendEmail(userID uint, accountID uint, to []string, subject string, bodyHTML string, bodyText string) error {
	if len(to) == 0 {
		return errors.New("to required")
	}
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return errors.New("subject required")
	}
	if strings.TrimSpace(bodyHTML) == "" && strings.TrimSpace(bodyText) == "" {
		return errors.New("body required")
	}
	for _, addr := range to {
		if err := validateEmail(strings.TrimSpace(addr)); err != nil {
			return errors.New("invalid recipient")
		}
	}

	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var account sysModel.SysSenderEmailAccount
		if accountID == 0 {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("sys_user_id = ? AND is_default = ?", userID, true).First(&account).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND sys_user_id = ?", accountID, userID).First(&account).Error; err != nil {
				return err
			}
		}

		if account.DailyLimit > 0 && account.TodaySent+len(to) > account.DailyLimit {
			return errors.New("daily limit reached")
		}

		secret, err := utils.DecryptStringAESGCM(account.SecretEnc)
		if err != nil {
			return err
		}

		var auth smtp.Auth
		if account.IsLoginAuth {
			auth = &loginAuth{username: account.Email, password: secret}
		} else {
			auth = smtp.PlainAuth("", account.Email, secret, account.SMTPHost)
		}

		e := email.NewEmail()
		if strings.TrimSpace(account.DisplayName) != "" {
			e.From = fmt.Sprintf("%s <%s>", account.DisplayName, account.Email)
		} else {
			e.From = account.Email
		}
		e.To = to
		e.Subject = subject
		if strings.TrimSpace(bodyHTML) != "" {
			e.HTML = []byte(bodyHTML)
		}
		if strings.TrimSpace(bodyText) != "" {
			e.Text = []byte(bodyText)
		}

		hostAddr := fmt.Sprintf("%s:%d", account.SMTPHost, account.SMTPPort)
		security := normalizeSecurity(account.Security)
		if security == "ssl" {
			if err := e.SendWithTLS(hostAddr, auth, &tls.Config{ServerName: account.SMTPHost}); err != nil {
				return err
			}
		} else {
			if err := e.Send(hostAddr, auth); err != nil {
				return err
			}
		}

		return tx.Model(&sysModel.SysSenderEmailAccount{}).Where("id = ?", account.ID).Update("today_sent", gorm.Expr("today_sent + ?", len(to))).Error
	})
}
