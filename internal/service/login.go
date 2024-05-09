package service

import (
	"context"
	"fmt"
	"imapi/internal/model"
	"imapi/internal/schema"
	"imapi/internal/utils"
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

	token := setToken(user.Uid)

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"token": token,
		"data":  schema.GetUser(user),
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

	token := setToken(user.Uid)

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"token": token,
		"data":  schema.GetUser(user),
	})
}

func setToken(uid uint64) string {
	nowtime := time.Now().Unix()
	token := utils.GenMd5(fmt.Sprintf("%d%d", uid, nowtime))
	rkey := model.Rktoken(uid)

	utils.RDB.Set(context.TODO(), rkey, token, time.Minute*time.Duration(0))
	utils.RDB.ExpireAt(context.TODO(), rkey, time.Now().Add(time.Minute*60*24*2))
	return token
}
