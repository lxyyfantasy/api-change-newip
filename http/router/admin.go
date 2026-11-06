package router

import (
    "github.com/gin-gonic/gin"
    _ "github.com/lejianwen/rustdesk-api/v2/docs/admin"
    "github.com/lejianwen/rustdesk-api/v2/global"
    "github.com/lejianwen/rustdesk-api/v2/http/controller/admin"
    "github.com/lejianwen/rustdesk-api/v2/http/controller/admin/my"
    "github.com/lejianwen/rustdesk-api/v2/http/middleware"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

func Init(g *gin.Engine) {

    //swagger
    //g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    if global.Config.App.ShowSwagger == 1 {
        g.GET("/admin/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.InstanceName("admin")))
    }

    adg := g.Group("/api/admin")
    LoginBind(adg)
    adg.POST("/user/register", (&admin.User{}).Register)

    ConfigBind(adg)

    adg.Use(middleware.BackendUserAuth())
    //FileBind(adg)
    UserBind(adg)
    GroupBind(adg)
    TagBind(adg)
    AddressBookBind(adg)
    PeerBind(adg)
    ActivationBind(adg)
    OauthBind(adg)
    LoginLogBind(adg)
    AuditBind(adg)
    AddressBookCollectionBind(adg)
    AddressBookCollectionRuleBind(adg)
    UserTokenBind(adg)

    //deprecated by ConfigBind
    //rs := &admin.Rustdesk{}
    //adg.GET("/server-config", rs.ServerConfig)
    //adg.GET("/app-config", rs.AppConfig)
    //deprecated end

    ShareRecordBind(adg)
    MyBind(adg)

    RustdeskCmdBind(adg)
    DeviceGroupBind(adg)
    //访问静态文件
    //g.StaticFS("/upload", http.Dir(global.Config.Gin.ResourcesPath+"/upload"))
}

func RustdeskCmdBind(adg *gin.RouterGroup) {
    cont := &admin.Rustdesk{}
    rg := adg.Group("/rustdesk").Use(middleware.AdminPrivilege())
    rg.POST("/sendCmd", middleware.AdminOpPassword(), cont.SendCmd)
    rg.GET("/cmdList", cont.CmdList)
    rg.POST("/cmdDelete", middleware.AdminOpPassword(), cont.CmdDelete)
    rg.POST("/cmdCreate", middleware.AdminOpPassword(), cont.CmdCreate)
}
func LoginBind(rg *gin.RouterGroup) {
    cont := &admin.Login{}
    rg.POST("/login", cont.Login)
    rg.GET("/captcha", cont.Captcha)
    rg.POST("/logout", cont.Logout)
    rg.GET("/login-options", cont.LoginOptions)
    rg.POST("/oidc/auth", cont.OidcAuth)
    rg.GET("/oidc/auth-query", cont.OidcAuthQuery)
}

func UserBind(rg *gin.RouterGroup) {
    aR := rg.Group("/user")
    {
        cont := &admin.User{}
        aR.GET("/current", cont.Current)
        aR.POST("/changeCurPwd", cont.ChangeCurPwd)
        aR.POST("/myOauth", cont.MyOauth)
        //aR.GET("/myPeer", cont.MyPeer)
        aR.POST("/groupUsers", cont.GroupUsers)
    }
    aRP := rg.Group("/user").Use(middleware.AdminPrivilege())
    {
        cont := &admin.User{}
        aRP.GET("/list", cont.List)
        aRP.GET("/detail/:id", cont.Detail)
        aRP.POST("/create", middleware.AdminOpPassword(), cont.Create)
        aRP.POST("/update", middleware.AdminOpPassword(), cont.Update)
        aRP.POST("/delete", middleware.AdminOpPassword(), cont.Delete)
        aRP.POST("/changePwd", middleware.AdminOpPassword(), cont.UpdatePassword)
    }
}

func GroupBind(rg *gin.RouterGroup) {
    aR := rg.Group("/group").Use(middleware.AdminPrivilege())
    {
        cont := &admin.Group{}
        aR.GET("/list", cont.List)
        aR.GET("/detail/:id", cont.Detail)
        aR.POST("/create", middleware.AdminOpPassword(), cont.Create)
        aR.POST("/update", middleware.AdminOpPassword(), cont.Update)
        aR.POST("/delete", middleware.AdminOpPassword(), cont.Delete)
    }
}

func DeviceGroupBind(rg *gin.RouterGroup) {
    aR := rg.Group("/device_group").Use(middleware.AdminPrivilege())
    {
        cont := &admin.DeviceGroup{}
        aR.GET("/list", cont.List)
        aR.GET("/detail/:id", cont.Detail)
        aR.POST("/create", middleware.AdminOpPassword(), cont.Create)
        aR.POST("/update", middleware.AdminOpPassword(), cont.Update)
        aR.POST("/delete", middleware.AdminOpPassword(), cont.Delete)
    }
}

func TagBind(rg *gin.RouterGroup) {
    aR := rg.Group("/tag").Use(middleware.AdminPrivilege())
    {
        cont := &admin.Tag{}
        aR.GET("/list", cont.List)
        aR.GET("/detail/:id", cont.Detail)
        aR.POST("/create", middleware.AdminOpPassword(), cont.Create)
        aR.POST("/update", middleware.AdminOpPassword(), cont.Update)
        aR.POST("/delete", middleware.AdminOpPassword(), cont.Delete)
    }
}

func AddressBookBind(rg *gin.RouterGroup) {
    aR := rg.Group("/address_book")
    {
        cont := &admin.AddressBook{}
        aR.POST("/shareByWebClient", cont.ShareByWebClient)

        arp := aR.Use(middleware.AdminPrivilege(), middleware.AdminOpPassword())
        arp.GET("/list", cont.List)
        arp.POST("/create", cont.Create)
        arp.POST("/update", cont.Update)
        arp.POST("/delete", cont.Delete)
        arp.POST("/batchCreate", cont.BatchCreate)
        arp.POST("/batchCreateFromPeers", cont.BatchCreateFromPeers)
    }
}
func PeerBind(rg *gin.RouterGroup) {
    aR := rg.Group("/peer")
    aR.POST("/simpleData", (&admin.Peer{}).SimpleData)
    aR.Use(middleware.AdminPrivilege())
    {
        cont := &admin.Peer{}
        aR.GET("/list", cont.List)
        aR.GET("/detail/:id", cont.Detail)
        aR.POST("/create", middleware.AdminOpPassword(), cont.Create)
        aR.POST("/update", middleware.AdminOpPassword(), cont.Update)
        aR.POST("/delete", middleware.AdminOpPassword(), cont.Delete)
        aR.POST("/batchDelete", middleware.AdminOpPassword(), cont.BatchDelete)
        aR.POST("/setServerIp", middleware.AdminOpPassword(), cont.SetServerIp)
    }
}

func OauthBind(rg *gin.RouterGroup) {
    aR := rg.Group("/oauth")
    {
        cont := &admin.Oauth{}
        aR.POST("/confirm", cont.Confirm)
        aR.POST("/bind", cont.ToBind)
        aR.POST("/bindConfirm", cont.BindConfirm)
        aR.POST("/unbind", cont.Unbind)
        aR.GET("/info", cont.Info)
    }
    arp := aR.Use(middleware.AdminPrivilege())
    {
        cont := &admin.Oauth{}
        arp.GET("/list", cont.List)
        arp.GET("/detail/:id", cont.Detail)
        arp.POST("/create", middleware.AdminOpPassword(), cont.Create)
        arp.POST("/update", middleware.AdminOpPassword(), cont.Update)
        arp.POST("/delete", middleware.AdminOpPassword(), cont.Delete)
    }
}
func LoginLogBind(rg *gin.RouterGroup) {
    cont := &admin.LoginLog{}
    aR := rg.Group("/login_log").Use(middleware.AdminPrivilege())
    aR.GET("/list", cont.List)
    aR.POST("/delete", middleware.AdminOpPassword(), cont.Delete)
    aR.POST("/batchDelete", middleware.AdminOpPassword(), cont.BatchDelete)
}
func AuditBind(rg *gin.RouterGroup) {
    cont := &admin.Audit{}
    aR := rg.Group("/audit_conn").Use(middleware.AdminPrivilege())
    aR.GET("/list", cont.ConnList)
    aR.POST("/delete", middleware.AdminOpPassword(), cont.ConnDelete)
    aR.POST("/batchDelete", middleware.AdminOpPassword(), cont.BatchConnDelete)
    afR := rg.Group("/audit_file").Use(middleware.AdminPrivilege())
    afR.GET("/list", cont.FileList)
    afR.POST("/delete", middleware.AdminOpPassword(), cont.FileDelete)
    afR.POST("/batchDelete", middleware.AdminOpPassword(), cont.BatchFileDelete)
}
func AddressBookCollectionBind(rg *gin.RouterGroup) {
    aR := rg.Group("/address_book_collection").Use(middleware.AdminPrivilege())
    {
        cont := &admin.AddressBookCollection{}
        aR.GET("/list", cont.List)
        aR.GET("/detail/:id", cont.Detail)
        aR.POST("/create", middleware.AdminOpPassword(), cont.Create)
        aR.POST("/update", middleware.AdminOpPassword(), cont.Update)
        aR.POST("/delete", middleware.AdminOpPassword(), cont.Delete)
    }
}
func AddressBookCollectionRuleBind(rg *gin.RouterGroup) {
    aR := rg.Group("/address_book_collection_rule").Use(middleware.AdminPrivilege())
    {
        cont := &admin.AddressBookCollectionRule{}
        aR.GET("/list", cont.List)
        aR.GET("/detail/:id", cont.Detail)
        aR.POST("/create", middleware.AdminOpPassword(), cont.Create)
        aR.POST("/update", middleware.AdminOpPassword(), cont.Update)
        aR.POST("/delete", middleware.AdminOpPassword(), cont.Delete)
    }
}
func UserTokenBind(rg *gin.RouterGroup) {
    aR := rg.Group("/user_token").Use(middleware.AdminPrivilege())
    cont := &admin.UserToken{}
    aR.GET("/list", cont.List)
    aR.POST("/delete", middleware.AdminOpPassword(), cont.Delete)
    aR.POST("/batchDelete", middleware.AdminOpPassword(), cont.BatchDelete)
}
func ConfigBind(rg *gin.RouterGroup) {
    aR := rg.Group("/config")
    rs := &admin.Config{}

    aR.GET("/admin", rs.AdminConfig)

    aR.Use(middleware.BackendUserAuth())
    aR.GET("/server", rs.ServerConfig)
    aR.GET("/app", rs.AppConfig)

}

/*
func FileBind(rg *gin.RouterGroup) {
    aR := rg.Group("/file")
    {
        cont := &admin.File{}
        aR.POST("/notify", cont.Notify)
        aR.OPTIONS("/oss_token", nil)
        aR.OPTIONS("/upload", nil)
        aR.GET("/oss_token", cont.OssToken)
        aR.POST("/upload", cont.Upload)
    }
}*/

func MyBind(rg *gin.RouterGroup) {
    {
        cont := &my.ShareRecord{}
        rg.GET("/my/share_record/list", cont.List)
        rg.POST("/my/share_record/delete", middleware.AdminOpPassword(), cont.Delete)
        rg.POST("/my/share_record/batchDelete", middleware.AdminOpPassword(), cont.BatchDelete)
    }

    {
        cont := &my.AddressBook{}
        rg.GET("/my/address_book/list", cont.List)
        rg.POST("/my/address_book/create", middleware.AdminOpPassword(), cont.Create)
        rg.POST("/my/address_book/update", middleware.AdminOpPassword(), cont.Update)
        rg.POST("/my/address_book/delete", middleware.AdminOpPassword(), cont.Delete)
        rg.POST("/my/address_book/batchCreateFromPeers", middleware.AdminOpPassword(), cont.BatchCreateFromPeers)
        rg.POST("/my/address_book/batchUpdateTags", middleware.AdminOpPassword(), cont.BatchUpdateTags)
    }

    {
        cont := &my.Tag{}
        rg.GET("/my/tag/list", cont.List)
        rg.POST("/my/tag/create", middleware.AdminOpPassword(), cont.Create)
        rg.POST("/my/tag/update", middleware.AdminOpPassword(), cont.Update)
        rg.POST("/my/tag/delete", middleware.AdminOpPassword(), cont.Delete)
    }

    {
        cont := &my.AddressBookCollection{}
        rg.GET("/my/address_book_collection/list", cont.List)
        rg.POST("/my/address_book_collection/create", middleware.AdminOpPassword(), cont.Create)
        rg.POST("/my/address_book_collection/update", middleware.AdminOpPassword(), cont.Update)
        rg.POST("/my/address_book_collection/delete", middleware.AdminOpPassword(), cont.Delete)
    }

    {
        cont := &my.AddressBookCollectionRule{}
        rg.GET("/my/address_book_collection_rule/list", cont.List)
        rg.POST("/my/address_book_collection_rule/create", middleware.AdminOpPassword(), cont.Create)
        rg.POST("/my/address_book_collection_rule/update", middleware.AdminOpPassword(), cont.Update)
        rg.POST("/my/address_book_collection_rule/delete", middleware.AdminOpPassword(), cont.Delete)
    }
    {
        cont := &my.Peer{}
        rg.GET("/my/peer/list", cont.List)
    }

    {
        cont := &my.LoginLog{}
        rg.GET("/my/login_log/list", cont.List)
        rg.POST("/my/login_log/delete", middleware.AdminOpPassword(), cont.Delete)
        rg.POST("/my/login_log/batchDelete", middleware.AdminOpPassword(), cont.BatchDelete)
    }
}

func ShareRecordBind(rg *gin.RouterGroup) {
    aR := rg.Group("/share_record").Use(middleware.AdminPrivilege())
    {
        cont := &admin.ShareRecord{}
        aR.GET("/list", cont.List)
        aR.POST("/delete", middleware.AdminOpPassword(), cont.Delete)
        aR.POST("/batchDelete", middleware.AdminOpPassword(), cont.BatchDelete)
    }
}

func ActivationBind(rg *gin.RouterGroup) {
    cont := &admin.Activation{}
    aR := rg.Group("/").Use(middleware.AdminPrivilege())
    aR.PUT("/server-config", middleware.AdminOpPassword(), cont.ServerConfigPut)
    aR.GET("/codes/take", middleware.AdminOpPassword(), cont.TakeCodes)
    aR.GET("/removed-devices", cont.RemovedDevices)
    aR.GET("/removed-accounts", cont.RemovedAccounts)
    aR.GET("/codes/used", cont.UsedCodes)
    aR.GET("/pending/list", cont.PendingList)
    aR.POST("/pending/add", middleware.AdminOpPassword(), cont.PendingAdd)
    aR.POST("/pending/delete", middleware.AdminOpPassword(), cont.PendingDelete)
    aR.GET("/assigned/list", cont.AssignedList)
    aR.GET("/orphan/pull", cont.PullOrphan)
}