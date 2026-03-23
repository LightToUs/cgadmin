package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	sysReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SenderEmailAccountApi struct{}

func (a *SenderEmailAccountApi) GetSenderEmailAccountList(c *gin.Context) {
	userID := utils.GetUserID(c)
	data, err := senderEmailAccountService.GetList(userID)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithData(data, c)
}

func (a *SenderEmailAccountApi) CreateSenderEmailAccount(c *gin.Context) {
	var req sysReq.SenderEmailAccountCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	_, err := senderEmailAccountService.Create(userID, req)
	if err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

func (a *SenderEmailAccountApi) UpdateSenderEmailAccount(c *gin.Context) {
	var req sysReq.SenderEmailAccountUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	if err := senderEmailAccountService.Update(userID, req); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (a *SenderEmailAccountApi) DeleteSenderEmailAccount(c *gin.Context) {
	var req sysReq.SenderEmailAccountDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	if err := senderEmailAccountService.Delete(userID, req.ID); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (a *SenderEmailAccountApi) SetDefaultSenderEmailAccount(c *gin.Context) {
	var req sysReq.SenderEmailAccountSetDefaultReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	if err := senderEmailAccountService.SetDefault(userID, req.ID); err != nil {
		global.GVA_LOG.Error("设置默认失败!", zap.Error(err))
		response.FailWithMessage("设置默认失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("设置成功", c)
}

func (a *SenderEmailAccountApi) UpdateSenderEmailAccountQuota(c *gin.Context) {
	var req sysReq.SenderEmailAccountQuotaUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	if err := senderEmailAccountService.UpdateQuota(userID, req); err != nil {
		global.GVA_LOG.Error("更新限额失败!", zap.Error(err))
		response.FailWithMessage("更新限额失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (a *SenderEmailAccountApi) TestSenderEmailAccountConnection(c *gin.Context) {
	var req sysReq.SenderEmailAccountTestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := senderEmailAccountService.TestConnection(req); err != nil {
		global.GVA_LOG.Error("测试失败!", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("连接成功，测试邮件已发送", c)
}

func (a *SenderEmailAccountApi) TestSenderEmailAccountConnectionByID(c *gin.Context) {
	var req sysReq.SenderEmailAccountTestByIDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	if err := senderEmailAccountService.TestConnectionByID(userID, req.ID); err != nil {
		global.GVA_LOG.Error("测试失败!", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("连接成功，测试邮件已发送", c)
}
