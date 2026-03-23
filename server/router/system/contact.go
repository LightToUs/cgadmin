package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type ContactRouter struct{}

func (r *ContactRouter) InitContactRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	group := Router.Group("contact").Use(middleware.OperationRecord())
	groupWithoutRecord := Router.Group("contact")
	{
		group.POST("create", contactApi.CreateContact)
	}
	{
		groupWithoutRecord.GET("list", contactApi.GetContactList)
	}
}

