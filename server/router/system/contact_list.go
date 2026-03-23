package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type ContactListRouter struct{}

func (r *ContactListRouter) InitContactListRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	group := Router.Group("contactList").Use(middleware.OperationRecord())
	groupWithoutRecord := Router.Group("contactList")
	{
		group.POST("create", contactListApi.CreateContactList)
		group.PUT("update", contactListApi.UpdateContactList)
		group.DELETE("delete", contactListApi.DeleteContactList)
	}
	{
		groupWithoutRecord.GET("tree", contactListApi.GetContactListTree)
	}
}

