package api

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/lejianwen/rustdesk-api/v2/model"
    "github.com/lejianwen/rustdesk-api/v2/service"
)

type ActivationExternal struct {
}

func (e *ActivationExternal) Peers(c *gin.Context) {
    online := c.Query("online")
    now := time.Now().Unix()
    q := service.DB.Model(&model.Peer{})
    if online == "1" {
        q = q.Where("last_online_time > ?", now-120)
    } else if online == "0" {
        q = q.Where("last_online_time <= ? OR last_online_time = 0", now-120)
    }
    list := make([]*model.Peer, 0)
    q.Find(&list)
    c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": list}})
}