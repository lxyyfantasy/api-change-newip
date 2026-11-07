package api

import (
    "github.com/gin-gonic/gin"
    "github.com/lejianwen/rustdesk-api/v2/global"
    "github.com/lejianwen/rustdesk-api/v2/model"
    requstform "github.com/lejianwen/rustdesk-api/v2/http/request/api"
    apiResp "github.com/lejianwen/rustdesk-api/v2/http/response/api"
    "github.com/lejianwen/rustdesk-api/v2/service"
    "github.com/lejianwen/rustdesk-api/v2/utils"
    "net/http"
)

type User struct {
}

// Info 用户信息
// @Tags 用户
// @Summary 用户信息
// @Description 用户信息
// @Accept  json
// @Produce  json
// @Success 200 {object} apiResp.UserPayload
// @Failure 500 {object} response.Response
// @Router /currentUser [get]
// @Security token
func (u *User) Info(c *gin.Context) {
    user := service.AllService.UserService.CurUser(c)
    up := (&apiResp.UserPayload{}).FromUser(user)
    c.JSON(http.StatusOK, up)
}

// accountResp 客户端 AccountApi 约定返回 {code, msg}，code == "OK" 为成功
func accountResp(c *gin.Context, code, msg string) {
    c.JSON(http.StatusOK, gin.H{
        "code": code,
        "msg":  msg,
    })
}

// Register 客户端注册接口（公开，无 /api 前缀）
// 与网页端共用 app.register 开关和 UserService.Register 逻辑
func (u *User) Register(c *gin.Context) {
    if !global.Config.App.Register {
        accountResp(c, "ERROR", "注册功能未开启")
        return
    }
    form := &requstform.RegisterForm{}
    if err := c.ShouldBindJSON(form); err != nil {
        accountResp(c, "ERROR", "参数错误: "+err.Error())
        return
    }
    status := model.StatusCode(global.Config.App.RegisterStatus)

    // 事务：建用户 → 存答案 → 删邀请码（最后一步，且检查影响行数）
    tx := service.DB.Begin()
    user := &model.User{
        Username: form.Username, Email: "", Password: form.Password,
        GroupId: 1, Status: status,
    }
    if err := service.AllService.UserService.CreateTx(tx, user); err != nil {
        tx.Rollback()
        accountResp(c, "ERROR", "注册失败：用户名已存在")
        return
    }
    if err := service.AllService.ActivationService.SaveRegInfoTx(tx, form.Username, form.Answer1, form.Answer2, form.Answer3); err != nil {
        tx.Rollback()
        accountResp(c, "ERROR", "注册失败：保存答案出错")
        return
    }
    rows, err := service.AllService.ActivationService.UseInviteCodeTx(tx, form.InviteCode)
    if err != nil || rows != 1 {
        tx.Rollback()
        accountResp(c, "INVITE_ERROR", "邀请码已被使用")
        return
    }
    tx.Commit()
    accountResp(c, "OK", "注册成功")
}

// 改的是当前登录账号自己的密码，不是 RustDesk ID 密码
func (u *User) ChangePassword(c *gin.Context) {
    form := &requstform.ChangePasswordForm{}
    if err := c.ShouldBindJSON(form); err != nil {
        accountResp(c, "ERROR", "参数错误: "+err.Error())
        return
    }
    cur := service.AllService.UserService.InfoByUsername(form.Username)
    if cur == nil || cur.Id == 0 {
        accountResp(c, "ERROR", "用户不存在")
        return
    }
    ok, _, err := utils.VerifyPassword(cur.Password, form.OldPassword)
    if err != nil || !ok {
        accountResp(c, "ERROR", "原密码错误")
        return
    }
    if err := service.AllService.UserService.UpdatePassword(cur, form.NewPassword); err != nil {
        accountResp(c, "ERROR", "密码修改失败: "+err.Error())
        return
    }
    accountResp(c, "OK", "密码修改成功")
}

// RecoverPassword 找回密码
func (u *User) RecoverPassword(c *gin.Context) {
    form := &requstform.RecoverPasswordForm{}
    if err := c.ShouldBindJSON(form); err != nil {
        accountResp(c, "ERROR", "参数错误: "+err.Error())
        return
    }
    if !service.AllService.ActivationService.CheckRegInfo(form.Username, form.Answer1, form.Answer2, form.Answer3) {
        accountResp(c, "ERROR", "账号或答案错误")
        return
    }
    cur := service.AllService.UserService.InfoByUsername(form.Username)
    if cur == nil || cur.Id == 0 {
        accountResp(c, "ERROR", "用户不存在")
        return
    }
    if err := service.AllService.UserService.UpdatePassword(cur, form.NewPassword); err != nil {
        accountResp(c, "ERROR", "密码修改失败: "+err.Error())
        return
    }
    accountResp(c, "OK", "密码修改成功")
}

// AddDevice 添加设备
func (u *User) AddDevice(c *gin.Context) {
    form := &requstform.AddDeviceForm{}
    if err := c.ShouldBindJSON(form); err != nil {
        accountResp(c, "ERROR", "参数错误: "+err.Error())
        return
    }
    if !service.AllService.ActivationService.CheckAddDeviceCode(form.DeviceCode) {
        accountResp(c, "CODE_ERROR", "添加设备码不对")
        return
    }
    if service.AllService.ActivationService.PendingCount() <= 0 {
        accountResp(c, "EMPTY", "设备库正在补充，请耐心等待")
        return
    }
    cur := service.AllService.UserService.InfoByUsername(form.Username)
    if cur == nil || cur.Id == 0 {
        accountResp(c, "ERROR", "用户不存在")
        return
    }

    tx := service.DB.Begin()
    ok, deviceId := service.AllService.ActivationService.TakePendingDeviceTx(tx, form.Username)
    if !ok {
        tx.Rollback()
        accountResp(c, "EMPTY", "设备库正在补充，请耐心等待")
        return
    }
    peer := service.AllService.PeerService.FindById(deviceId)
    if peer.RowId == 0 {
        tx.Rollback()
        accountResp(c, "ERROR", "设备信息不存在，请稍后重试")
        return
    }
    ab := service.AllService.AddressBookService.FromPeer(peer)
    ab.UserId = cur.Id
    if err := service.AllService.AddressBookService.CreateTx(tx, ab); err != nil {
        tx.Rollback()
        accountResp(c, "ERROR", "添加失败")
        return
    }
    // 最后一步才消费码，并检查影响行数
    rows, err := service.AllService.ActivationService.UseAddDeviceCodeTx(tx, form.DeviceCode)
    if err != nil || rows != 1 {
        tx.Rollback()
        accountResp(c, "CODE_ERROR", "添加设备码已被使用")
        return
    }
    tx.Commit()
    accountResp(c, "OK", "添加成功:"+deviceId)
}

// PendingCount 获取待分配设备
func (u *User) PendingCount(c *gin.Context) {
    cnt := service.AllService.ActivationService.PendingCount()
    c.JSON(http.StatusOK, gin.H{"code": "OK", "data": gin.H{"count": cnt}})
}

// CheckInvite 设置 ID 时校验：邀请码 + ID + 设备识别码
func (u *User) CheckInvite(c *gin.Context) {
	form := &struct {
		InviteCode string `json:"invite_code" binding:"required"`
		Id         string `json:"id"`
		DeviceCode string `json:"device_code"`
	}{}
	if err := c.ShouldBindJSON(form); err != nil {
		accountResp(c, "ERROR", "参数错误: "+err.Error())
		return
	}
	// 一、邀请码不对直接拒绝
	if !service.AllService.ActivationService.CheckInviteCode(form.InviteCode) {
		accountResp(c, "INVITE_ERROR", "邀请码错误")
		return
	}
	// 二、邀请码对
	// 1. 设备识别码有 → 只覆盖 ID，地址簿等其它信息不动
	if form.DeviceCode != "" {
		if same := service.AllService.PeerService.FindByDeviceCode(form.DeviceCode); same.RowId > 0 {
			if form.Id != "" {
				if p := service.AllService.PeerService.FindById(form.Id); p.RowId > 0 && p.DeviceCode != form.DeviceCode {
					accountResp(c, "ID_DUPLICATE", "ID 已存在")
					return
				}
			}
			service.DB.Model(&model.Peer{}).Where("row_id = ?", same.RowId).Updates(map[string]interface{}{
				"id":          form.Id,
				"device_code": form.DeviceCode,
			})
			service.DB.Model(&model.AddressBook{}).Where("id = ?", same.Id).Update("id", form.Id)
		} else {
			// 设备码没记录：ID 有且被别的设备占 → 拒；否则建新
			if form.Id != "" {
				if p := service.AllService.PeerService.FindById(form.Id); p.RowId > 0 {
					accountResp(c, "ID_DUPLICATE", "ID 已存在")
					return
				}
			}
			service.DB.Create(&model.Peer{Id: form.Id, DeviceCode: form.DeviceCode})
		}
		accountResp(c, "OK", "邀请码有效")
		return
	}
	// 2. 设备识别码没有，有 ID 且已存在 → 拒绝
	if form.Id != "" {
		if p := service.AllService.PeerService.FindById(form.Id); p.RowId > 0 {
			accountResp(c, "ID_DUPLICATE", "ID 已存在")
			return
		}
	}
	// 3. 设备识别码没有，没有 ID → 放行
	accountResp(c, "OK", "邀请码有效")
}