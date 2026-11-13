package router

import (
    "github.com/gin-gonic/gin"
    _ "github.com/lejianwen/rustdesk-api/v2/docs/api"
    "github.com/lejianwen/rustdesk-api/v2/global"
    "github.com/lejianwen/rustdesk-api/v2/http/controller/api"
    "github.com/lejianwen/rustdesk-api/v2/http/middleware"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
    "net/http"
)

func ApiInit(g *gin.Engine) {

    //g.Use(middleware.Cors())
    //swagger
    if global.Config.App.ShowSwagger == 1 {
        g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.InstanceName("api")))
    }
    // 加载 HTML 模板
    g.LoadHTMLGlob("resources/templates/*")

    // Flutter 客户端账号接口（根路径，无 /api 前缀，适配 AccountApi）
    {
        u := &api.User{}
        // 注册：公开，是否可用由 app.register 控制（与网页端同一开关）
        g.POST("/register", u.Register)
        g.POST("/recover-password", u.RecoverPassword)
        g.POST("/add-device", u.AddDevice)
        g.GET("/pending-count", u.PendingCount)
        g.POST("/check-invite", u.CheckInvite)
        // 改密码：公开，凭 username + old_password 校验（与客户端 AccountApi 一致）
        g.POST("/change-password", u.ChangePassword)
        g.POST("/register-web", u.RegisterWeb)
    }
    
    frg := g.Group("/api")

    {
        i := &api.Index{}
        frg.GET("/", i.Index)
        frg.GET("/version", i.Version)

        frg.POST("/heartbeat", i.Heartbeat)
        {
            ac := &api.Activation{}
            frg.POST("/activate", ac.Activate)
            frg.POST("/report", ac.Report)
            frg.GET("/device/info", ac.DeviceInfo)
            frg.GET("/peer/all", ac.PeerAll)
            frg.GET("/orphan-host/pull", ac.PullOrphanHost)
            frg.GET("/lease-host/ips", ac.LeaseHostIps)
            frg.GET("/server-config", ac.ServerConfig)
        }
    }

    {
        l := &api.Login{}
        // 如果返回oidc则可以通过oidc登录
        frg.GET("/login-options", l.LoginOptions)
        frg.POST("/login", l.Login)

    }

    {
        o := &api.Oauth{}
        // [method:POST] [uri:/api/oidc/auth]
        frg.POST("/oidc/auth", o.OidcAuth)
        // [method:GET] [uri:/api/oidc/auth-query?code=abc&id=xxxxx&uuid=xxxxx]
        frg.GET("/oidc/auth-query", o.OidcAuthQuery)
        //api/oauth/callback
        frg.GET("/oauth/callback", o.OauthCallback)
        frg.GET("/oauth/login", o.OauthCallback)
        frg.GET("/oauth/msg", o.Message)

        frg.GET("/oidc/callback", o.OauthCallback)
        frg.GET("/oidc/login", o.OauthCallback)
        frg.GET("/oidc/msg", o.Message)
    }
    {
        pe := &api.Peer{}
        //提交系统信息
        frg.POST("/sysinfo", pe.SysInfo)
        frg.POST("/sysinfo_ver", pe.SysInfoVer)
        frg.GET("/peer/server_ip", pe.ServerIp)
    }

    if global.Config.App.WebClient == 1 {
        WebClientRoutes(frg)
    }

    {
        au := &api.Audit{}
        //[method:POST] [uri:/api/audit/conn]
        frg.POST("/audit/conn", au.AuditConn)
        //[method:POST] [uri:/api/audit/file]
        frg.POST("/audit/file", au.AuditFile)
    }

    {
        ext := &api.ActivationExternal{}
        frg.GET("/external/peers", ext.Peers)
    }

    frg.Use(middleware.RustAuth())
    {
        u := &api.User{}
        frg.GET("/user/info", u.Info)
        frg.POST("/currentUser", u.Info)
    }
    {
        l := &api.Login{}
        frg.POST("/logout", l.Logout)
    }
    {
        gr := &api.Group{}
        frg.GET("/users", gr.Users)
        frg.GET("/peers", gr.Peers)
        // /api/device-group/accessible?current=1&pageSize=100
        frg.GET("/device-group/accessible", gr.Device)
    }

    {
        ab := &api.Ab{}
        //获取地址
        frg.GET("/ab", ab.Ab)
        //更新地址
        frg.POST("/ab", ab.UpAb)
    }

    PersonalRoutes(frg)
    //访问静态文件
    g.StaticFS("/upload", http.Dir(global.Config.Gin.ResourcesPath+"/public/upload"))
}

func PersonalRoutes(frg *gin.RouterGroup) {
    {
        ab := &api.Ab{}
        frg.POST("/ab/personal", ab.Personal)
        //[method:POST] [uri:/api/ab/settings] Request
        frg.POST("/ab/settings", ab.Settings)
        // [method:POST] [uri:/api/ab/shared/profiles?current=1&pageSize=100]
        frg.POST("/ab/shared/profiles", ab.SharedProfiles)
        //[method:POST] [uri:/api/ab/peers?current=1&pageSize=100&ab=1]
        frg.POST("/ab/peers", ab.Peers)
        // [method:POST] [uri:/api/ab/tags/1]
        frg.POST("/ab/tags/:guid", ab.PTags)
        //[method:POST] api/ab/peer/add/1
        frg.POST("/ab/peer/add/:guid", ab.PeerAdd)
        //[method:DELETE] [uri:/api/ab/peer/1]
        frg.DELETE("/ab/peer/:guid", ab.PeerDel)
        //[method:PUT] [uri:/api/ab/peer/update/1]
        frg.PUT("/ab/peer/update/:guid", ab.PeerUpdate)
        //[method:POST] [uri:/api/ab/tag/add/1]
        frg.POST("/ab/tag/add/:guid", ab.TagAdd)
        //[method:PUT] [uri:/api/ab/tag/rename/1]
        frg.PUT("/ab/tag/rename/:guid", ab.TagRename)
        //[method:PUT] [uri:/api/ab/tag/update/1]
        frg.PUT("/ab/tag/update/:guid", ab.TagUpdate)
        //[method:DELETE] [uri:/api/ab/tag/1]
        frg.DELETE("/ab/tag/:guid", ab.TagDel)

    }

}

func WebClientRoutes(frg *gin.RouterGroup) {
    w := &api.WebClient{}
    {
        frg.POST("/shared-peer", w.SharedPeer)
    }
    {
        frg.POST("/server-config", middleware.RustAuth(), w.ServerConfig)
        frg.POST("/server-config-v2", middleware.RustAuth(), w.ServerConfigV2)
    }

}
