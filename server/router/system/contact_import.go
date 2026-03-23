package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type ContactImportRouter struct{}

func (r *ContactImportRouter) InitContactImportRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	group := Router.Group("contactImport").Use(middleware.OperationRecord())
	groupWithoutRecord := Router.Group("contactImport")
	groupWithoutAuth := PublicRouter.Group("contactImport")
	{
		group.POST("upload", contactImportApi.UploadContactImport)
		group.POST("googleSheet", contactImportApi.UploadGoogleSheet)
		group.POST("suggestMapping", contactImportApi.SuggestMapping)
		group.POST("validate", contactImportApi.ValidateImport)
		group.POST("start", contactImportApi.StartImport)
		group.POST("delete", contactImportApi.DeleteImportJob)
		group.POST("exportFailed", contactImportApi.ExportFailed)
	}
	{
		groupWithoutRecord.GET("job", contactImportApi.GetImportJob)
		groupWithoutRecord.GET("history", contactImportApi.GetImportHistory)
		groupWithoutRecord.GET("errors", contactImportApi.GetImportErrors)
	}
	{
		groupWithoutAuth.GET("downloadByToken", contactImportApi.DownloadByToken)
		groupWithoutAuth.GET("template", contactImportApi.DownloadTemplate)
	}
}

