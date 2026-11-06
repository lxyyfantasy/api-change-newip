package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/lejianwen/rustdesk-api/v2/global"
    "github.com/lejianwen/rustdesk-api/v2/http/response"
    "github.com/lejianwen/rustdesk-api/v2/service"
    "github.com/lejianwen/rustdesk-api/v2/utils"
)

// AdminPrivilege ...
func AdminPrivilege() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := service.AllService.UserService.CurUser(c)

		if !service.AllService.UserService.IsAdmin(u) {
			response.Fail(c, 403, response.TranslateMsg(c, "NoAccess"))
			c.Abort()
			return
		}

		c.Next()
	}
}

func AdminOpPassword() gin.HandlerFunc {
    return func(c *gin.Context) {
        if !global.Config.Admin.OpPasswordEnable {
            c.Next()
            return
        }
        pwd := c.GetHeader("X-Op-Password")
        if pwd == "" {
            response.Fail(c, 401, "需要操作密码")
            c.Abort()
            return
        }
        ok, _, err := utils.VerifyPassword(global.Config.Admin.OpPassword, pwd)
        if err != nil || !ok {
            response.Fail(c, 401, "操作密码错误")
            c.Abort()
            return
        }
        c.Next()
    }
}