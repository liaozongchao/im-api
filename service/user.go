package service

import (
	"context"
	"demoapi/model"
	"demoapi/schema"
	"demoapi/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	data := schema.Login{}
	c.Bind(&data)
	user, err := model.FindUserByName(data.Username)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "操作错误"})
		return
	}
	if user.Uid == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "用户不存在"})
		return
	}
	if user.Password != utils.GenMd5(data.Password) {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "密码错误"})
		return
	}

	sessKey := setSession(user.Uid)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"sessKey": sessKey,
		"data":    user,
	})
}

func Register(c *gin.Context) {
	data := schema.Register{}
	c.Bind(&data)
	if data.Password != data.Repassword {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "密码和确认密码不一致"})
		return
	}
	insertData := &model.User{
		Username: data.Username,
		Password: utils.GenMd5(data.Password),
		Email:    "",
		Phone:    "",
	}
	user, err := model.CreateUser(insertData)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "操作错误"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": user,
	})
}

func setSession(uid uint64) string {
	nowtime := time.Now().Unix()
	sesskey := utils.GenMd5(fmt.Sprintf("%d%d", uid, nowtime))
	rkey := model.Rkonline(uid)

	utils.RDB.Set(context.TODO(), rkey, sesskey, time.Minute*time.Duration(0))
	utils.RDB.ExpireAt(context.TODO(), rkey, time.Now().Add(time.Minute*60*24*2))
	return sesskey
}

func ChatMsg(c *gin.Context) {
	data := schema.ChatMsg{}
	c.Bind(&data)
	ctx := context.Background()
	var rkey string
	if data.MsgType == 1 {
		if data.FromId > data.ToId {
			rkey = fmt.Sprintf("msg_%d_%d", data.ToId, data.FromId)
		} else {
			rkey = fmt.Sprintf("msg_%d_%d", data.FromId, data.ToId)
		}
	}
	if data.MsgType == 2 {
		rkey = fmt.Sprintf("msg_%d_%d", 0, data.FromId)
	}
	var chats []string
	var err error
	if data.IsRev == 1 {
		chats, err = utils.RDB.ZRevRange(ctx, rkey, data.Start, data.End).Result()
	} else {
		chats, err = utils.RDB.ZRange(ctx, rkey, data.Start, data.End).Result()
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "数据错误"})
		return
	}
	newChats := utils.ReverseStringArray(chats)
	var tempChats []*model.Message
	for _, v := range newChats {
		msg := &model.Message{}
		json.Unmarshal([]byte(v), msg)
		tempChats = append(tempChats, msg)
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": tempChats,
	})
}

/***********************************************/
func EditUser(c *gin.Context) {
	data := schema.EditUser{}
	c.Bind(&data)
	user, err := model.FindUserByUid(data.Uid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "操作错误"})
		return
	}
	if user.Uid == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "用户不存在"})
		return
	}
	updateData := &model.User{
		Uid:      data.Uid,
		Username: data.Username,
		Info:     data.Info,
		Avatar:   data.Avatar,
	}
	user, err = model.UpdateUser(updateData)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "操作错误"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": user,
	})
}
