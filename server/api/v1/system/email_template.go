package system

import (
	"net/http"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	commonReq "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	sysReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type EmailTemplateApi struct{}

var (
	emailTemplateExportTokenCache      = make(map[string]any)
	emailTemplateExportTokenExpiration = make(map[string]time.Time)
	emailTemplateExportTokenMutex      sync.RWMutex
)

func init() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			emailTemplateExportTokenMutex.Lock()
			now := time.Now()
			for token, exp := range emailTemplateExportTokenExpiration {
				if now.After(exp) {
					delete(emailTemplateExportTokenCache, token)
					delete(emailTemplateExportTokenExpiration, token)
				}
			}
			emailTemplateExportTokenMutex.Unlock()
		}
	}()
}

func (a *EmailTemplateApi) GetEmailTemplateList(c *gin.Context) {
	var req sysReq.EmailTemplateSearchReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	list, total, err := emailTemplateService.List(userID, req)
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

func (a *EmailTemplateApi) GetEmailTemplateDetail(c *gin.Context) {
	var q commonReq.GetById
	if err := c.ShouldBindQuery(&q); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	tpl, err := emailTemplateService.GetByID(userID, q.Uint())
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"template": tpl}, c)
}

func (a *EmailTemplateApi) CreateEmailTemplate(c *gin.Context) {
	var req sysReq.EmailTemplateCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	_, err := emailTemplateService.Create(userID, req)
	if err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

func (a *EmailTemplateApi) UpdateEmailTemplate(c *gin.Context) {
	var req sysReq.EmailTemplateUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	if err := emailTemplateService.Update(userID, req); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (a *EmailTemplateApi) DeleteEmailTemplate(c *gin.Context) {
	var req sysReq.EmailTemplateDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	if err := emailTemplateService.Delete(userID, req.ID); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (a *EmailTemplateApi) DeleteEmailTemplateByIds(c *gin.Context) {
	var req commonReq.IdsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	if err := emailTemplateService.DeleteByIDs(userID, req); err != nil {
		global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

func (a *EmailTemplateApi) CopyEmailTemplate(c *gin.Context) {
	var req sysReq.EmailTemplateCopyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	_, err := emailTemplateService.Copy(userID, req)
	if err != nil {
		global.GVA_LOG.Error("复制失败!", zap.Error(err))
		response.FailWithMessage("复制失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("复制成功", c)
}

func (a *EmailTemplateApi) BatchStatusEmailTemplate(c *gin.Context) {
	var req sysReq.EmailTemplateBatchStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	if err := emailTemplateService.BatchStatus(userID, req); err != nil {
		global.GVA_LOG.Error("更新状态失败!", zap.Error(err))
		response.FailWithMessage("更新状态失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (a *EmailTemplateApi) MoveEmailTemplate(c *gin.Context) {
	var req sysReq.EmailTemplateMoveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	if err := emailTemplateService.Move(userID, req); err != nil {
		global.GVA_LOG.Error("移动失败!", zap.Error(err))
		response.FailWithMessage("移动失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("移动成功", c)
}

func (a *EmailTemplateApi) PreviewEmailTemplate(c *gin.Context) {
	var req sysReq.EmailTemplatePreviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	out, err := emailTemplateService.Preview(userID, req)
	if err != nil {
		global.GVA_LOG.Error("预览失败!", zap.Error(err))
		response.FailWithMessage("预览失败:"+err.Error(), c)
		return
	}
	response.OkWithData(out, c)
}

func (a *EmailTemplateApi) TestSendEmailTemplate(c *gin.Context) {
	var req sysReq.EmailTemplateTestSendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	if err := emailTemplateService.TestSend(userID, req); err != nil {
		global.GVA_LOG.Error("发送失败!", zap.Error(err))
		response.FailWithMessage("发送失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("发送成功", c)
}

func (a *EmailTemplateApi) ExportEmailTemplate(c *gin.Context) {
	var req sysReq.EmailTemplateExportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	b, contentType, filename, err := emailTemplateService.Export(userID, req)
	if err != nil {
		global.GVA_LOG.Error("导出失败!", zap.Error(err))
		response.FailWithMessage("导出失败:"+err.Error(), c)
		return
	}

	token := utils.MD5V([]byte(filename + time.Now().String()))
	exp := time.Now().Add(10 * time.Minute)
	emailTemplateExportTokenMutex.Lock()
	emailTemplateExportTokenCache[token] = gin.H{
		"content":     b,
		"contentType": contentType,
		"filename":    filename,
	}
	emailTemplateExportTokenExpiration[token] = exp
	emailTemplateExportTokenMutex.Unlock()

	response.OkWithData(gin.H{"token": token}, c)
}

func (a *EmailTemplateApi) DownloadEmailTemplateByToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.FailWithMessage("token required", c)
		return
	}
	emailTemplateExportTokenMutex.RLock()
	val, ok := emailTemplateExportTokenCache[token]
	exp := emailTemplateExportTokenExpiration[token]
	emailTemplateExportTokenMutex.RUnlock()
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

func (a *EmailTemplateApi) ImportEmailTemplate(c *gin.Context) {
	var req sysReq.EmailTemplateImportReq
	_ = c.ShouldBind(&req)
	file, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	count, err := emailTemplateService.Import(userID, file, req)
	if err != nil {
		global.GVA_LOG.Error("导入失败!", zap.Error(err))
		response.FailWithMessage("导入失败:"+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"count": count}, c)
}

