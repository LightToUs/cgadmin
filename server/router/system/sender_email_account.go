package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type SenderEmailAccountRouter struct{}

func (r *SenderEmailAccountRouter) InitSenderEmailAccountRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	group := Router.Group("senderEmailAccount").Use(middleware.OperationRecord())
	groupWithoutRecord := Router.Group("senderEmailAccount")
	{
		group.POST("create", senderEmailAccountApi.CreateSenderEmailAccount)
		group.PUT("update", senderEmailAccountApi.UpdateSenderEmailAccount)
		group.PUT("updateQuota", senderEmailAccountApi.UpdateSenderEmailAccountQuota)
		group.DELETE("delete", senderEmailAccountApi.DeleteSenderEmailAccount)
		group.POST("setDefault", senderEmailAccountApi.SetDefaultSenderEmailAccount)
		group.POST("testById", senderEmailAccountApi.TestSenderEmailAccountConnectionByID)
	}
	{
		groupWithoutRecord.GET("list", senderEmailAccountApi.GetSenderEmailAccountList)
		groupWithoutRecord.POST("test", senderEmailAccountApi.TestSenderEmailAccountConnection)
	}
}
