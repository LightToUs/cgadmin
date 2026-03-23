package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type EmailTemplateFolderRouter struct{}

func (r *EmailTemplateFolderRouter) InitEmailTemplateFolderRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	group := Router.Group("emailTemplateFolder").Use(middleware.OperationRecord())
	groupWithoutRecord := Router.Group("emailTemplateFolder")
	{
		group.POST("create", emailTemplateFolderApi.CreateEmailTemplateFolder)
		group.PUT("update", emailTemplateFolderApi.UpdateEmailTemplateFolder)
		group.DELETE("delete", emailTemplateFolderApi.DeleteEmailTemplateFolder)
	}
	{
		groupWithoutRecord.GET("tree", emailTemplateFolderApi.GetEmailTemplateFolderTree)
	}
}

