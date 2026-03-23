package system

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	sysModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	sysReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	sysResp "github.com/flipped-aurora/gin-vue-admin/server/model/system/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ContactImportApi struct{}

var (
	contactImportDownloadCache      = make(map[string]any)
	contactImportDownloadExpiration = make(map[string]time.Time)
	contactImportDownloadMutex      sync.RWMutex
)

func init() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			contactImportDownloadMutex.Lock()
			now := time.Now()
			for token, exp := range contactImportDownloadExpiration {
				if now.After(exp) {
					delete(contactImportDownloadCache, token)
					delete(contactImportDownloadExpiration, token)
				}
			}
			contactImportDownloadMutex.Unlock()
		}
	}()
}

func (a *ContactImportApi) UploadContactImport(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if file.Size > 10<<20 {
		response.FailWithMessage("文件最大10MB", c)
		return
	}
	f, err := file.Open()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	out, err := contactImportService.Upload(userID, file.Filename, data)
	if err != nil {
		global.GVA_LOG.Error("上传失败!", zap.Error(err))
		response.FailWithMessage("上传失败:"+err.Error(), c)
		return
	}
	response.OkWithData(out, c)
}

func (a *ContactImportApi) UploadGoogleSheet(c *gin.Context) {
	var req sysReq.ContactImportGoogleSheetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	out, err := contactImportService.UploadFromGoogleSheet(userID, req.URL)
	if err != nil {
		global.GVA_LOG.Error("导入失败!", zap.Error(err))
		response.FailWithMessage("导入失败:"+err.Error(), c)
		return
	}
	response.OkWithData(out, c)
}

func (a *ContactImportApi) SuggestMapping(c *gin.Context) {
	var req sysReq.ContactImportSuggestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	out, err := contactImportService.SuggestMapping(userID, req.JobID)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithData(out, c)
}

func (a *ContactImportApi) ValidateImport(c *gin.Context) {
	var req sysReq.ContactImportValidateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	out, err := contactImportService.Validate(userID, req)
	if err != nil {
		global.GVA_LOG.Error("验证失败!", zap.Error(err))
		response.FailWithMessage("验证失败:"+err.Error(), c)
		return
	}
	response.OkWithData(out, c)
}

func (a *ContactImportApi) StartImport(c *gin.Context) {
	var req sysReq.ContactImportStartReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	if err := contactImportService.Start(userID, req); err != nil {
		global.GVA_LOG.Error("开始导入失败!", zap.Error(err))
		response.FailWithMessage("开始导入失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("已开始导入", c)
}

func (a *ContactImportApi) GetImportJob(c *gin.Context) {
	var req sysReq.ContactImportJobGetReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	job, err := contactImportService.GetJob(userID, req.ID)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"job": job}, c)
}

func (a *ContactImportApi) GetImportHistory(c *gin.Context) {
	var req sysReq.ContactImportJobListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	userID := utils.GetUserID(c)
	list, total, err := contactImportService.ListJobs(userID, req)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "获取成功", c)
}

func (a *ContactImportApi) DeleteImportJob(c *gin.Context) {
	var req sysReq.ContactImportJobDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	if err := contactImportService.DeleteJob(userID, req.ID); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (a *ContactImportApi) GetImportErrors(c *gin.Context) {
	var req sysReq.ContactImportErrorsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}
	userID := utils.GetUserID(c)
	db := global.GVA_DB.Model(&sysModel.SysContactImportRowError{}).Where("sys_user_id = ? AND job_id = ?", userID, req.JobID)
	if strings.TrimSpace(req.Type) != "" {
		db = db.Where("errors_json LIKE ?", "%\""+strings.TrimSpace(req.Type)+"\"%")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	var list []sysModel.SysContactImportRowError
	if err := db.Order("row_index asc").Limit(req.PageSize).Offset(req.PageSize * (req.Page - 1)).Find(&list).Error; err != nil {
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "获取成功", c)
}

func (a *ContactImportApi) ExportFailed(c *gin.Context) {
	var req sysReq.ContactImportExportFailedReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	content, contentType, filename, err := contactImportService.ExportFailedCSV(userID, req.JobID)
	if err != nil {
		global.GVA_LOG.Error("导出失败!", zap.Error(err))
		response.FailWithMessage("导出失败:"+err.Error(), c)
		return
	}
	token := utils.MD5V([]byte(filename + time.Now().String()))
	exp := time.Now().Add(10 * time.Minute)
	contactImportDownloadMutex.Lock()
	contactImportDownloadCache[token] = gin.H{"content": content, "contentType": contentType, "filename": filename}
	contactImportDownloadExpiration[token] = exp
	contactImportDownloadMutex.Unlock()
	response.OkWithData(sysResp.ContactImportExportFailedResp{Token: token}, c)
}

func (a *ContactImportApi) DownloadByToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.FailWithMessage("token required", c)
		return
	}
	contactImportDownloadMutex.RLock()
	val, ok := contactImportDownloadCache[token]
	exp := contactImportDownloadExpiration[token]
	contactImportDownloadMutex.RUnlock()
	if !ok || time.Now().After(exp) {
		response.FailWithMessage("token expired", c)
		return
	}
	m, ok := val.(gin.H)
	if !ok {
		response.FailWithMessage("invalid token", c)
		return
	}
	content, _ := m["content"].([]byte)
	contentType, _ := m["contentType"].(string)
	filename, _ := m["filename"].(string)
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, content)
}

func (a *ContactImportApi) DownloadTemplate(c *gin.Context) {
	format := strings.ToLower(strings.TrimSpace(c.Query("format")))
	if format == "" {
		format = "xlsx"
	}
	var content []byte
	var contentType string
	var filename string
	var err error
	if format == "csv" {
		content, contentType, filename, err = contactImportService.TemplateCSV()
	} else {
		content, contentType, filename, err = contactImportService.TemplateXLSX()
	}
	if err != nil {
		response.FailWithMessage("生成失败:"+err.Error(), c)
		return
	}
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, content)
}
