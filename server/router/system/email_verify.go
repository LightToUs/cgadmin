package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type EmailVerifyRouter struct{}

func (r *EmailVerifyRouter) InitEmailVerifyRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	group := Router.Group("emailVerify").Use(middleware.OperationRecord())
	groupWithoutRecord := Router.Group("emailVerify")
	{
		group.POST("start", emailVerifyApi.StartEmailVerify)
	}
	{
		groupWithoutRecord.GET("job", emailVerifyApi.GetEmailVerifyJob)
		groupWithoutRecord.GET("history", emailVerifyApi.GetEmailVerifyHistory)
	}
}

