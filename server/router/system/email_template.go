package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type EmailTemplateRouter struct{}

func (r *EmailTemplateRouter) InitEmailTemplateRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	group := Router.Group("emailTemplate").Use(middleware.OperationRecord())
	groupWithoutRecord := Router.Group("emailTemplate")
	groupWithoutAuth := PublicRouter.Group("emailTemplate")
	{
		group.POST("create", emailTemplateApi.CreateEmailTemplate)
		group.PUT("update", emailTemplateApi.UpdateEmailTemplate)
		group.DELETE("delete", emailTemplateApi.DeleteEmailTemplate)
		group.DELETE("deleteByIds", emailTemplateApi.DeleteEmailTemplateByIds)
		group.POST("copy", emailTemplateApi.CopyEmailTemplate)
		group.POST("batchStatus", emailTemplateApi.BatchStatusEmailTemplate)
		group.POST("move", emailTemplateApi.MoveEmailTemplate)
		group.POST("preview", emailTemplateApi.PreviewEmailTemplate)
		group.POST("testSend", emailTemplateApi.TestSendEmailTemplate)
		group.POST("export", emailTemplateApi.ExportEmailTemplate)
		group.POST("import", emailTemplateApi.ImportEmailTemplate)
	}
	{
		groupWithoutRecord.GET("list", emailTemplateApi.GetEmailTemplateList)
		groupWithoutRecord.GET("detail", emailTemplateApi.GetEmailTemplateDetail)
	}
	{
		groupWithoutAuth.GET("downloadByToken", emailTemplateApi.DownloadEmailTemplateByToken)
	}
}

