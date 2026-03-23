package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	sysReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type EmailVerifyApi struct{}

func (a *EmailVerifyApi) StartEmailVerify(c *gin.Context) {
	var req sysReq.EmailVerifyStartReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := emailVerifyService.ValidateStart(req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	job, err := emailVerifyService.Start(userID, req)
	if err != nil {
		global.GVA_LOG.Error("启动失败!", zap.Error(err))
		response.FailWithMessage("启动失败:"+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"job": job}, c)
}

func (a *EmailVerifyApi) GetEmailVerifyJob(c *gin.Context) {
	var req sysReq.EmailVerifyJobGetReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	job, err := emailVerifyService.GetJob(userID, req.ID)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"job": job}, c)
}

func (a *EmailVerifyApi) GetEmailVerifyHistory(c *gin.Context) {
	var req sysReq.EmailVerifyHistoryReq
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
	list, total, err := emailVerifyService.List(userID, req.Page, req.PageSize)
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

