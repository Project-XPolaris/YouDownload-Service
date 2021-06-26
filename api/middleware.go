package api

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youdownload-server/config"
	"github.com/projectxpolaris/youdownload-server/hub"
	"github.com/projectxpolaris/youdownload-server/youplus"
	"github.com/sirupsen/logrus"
	"strings"
)

var noAuthPath = []string{"/info", "/user/auth"}

type AuthMiddleware struct {
}

func (m AuthMiddleware) OnRequest(ctx *haruka.Context) {
	if !config.Instance.AuthEnable {
		ctx.Param["uid"] = hub.PublicUid
		ctx.Param["username"] = hub.PublicUsername
		return
	}
	for _, targetPath := range noAuthPath {
		if ctx.Request.URL.Path == targetPath {
			return
		}
	}
	rawString := ctx.Request.Header.Get("Authorization")
	if len(rawString) == 0 {
		rawString = ctx.GetQueryString("token")
	}
	if len(rawString) > 0 {
		rawString = strings.Replace(rawString, "Bearer ", "", 1)
		response, err := youplus.DefaultAuthClient.CheckAuth(rawString)
		if err == nil {
			logrus.WithFields(logrus.Fields{
				"uid":  response.Uid,
				"user": response.Username,
			}).Info("user auth")
			ctx.Param["uid"] = response.Uid
			ctx.Param["username"] = response.Username
		} else {
			ctx.Param["uid"] = hub.PublicUid
			ctx.Param["username"] = hub.PublicUsername
		}
	} else {
		ctx.Param["uid"] = hub.PublicUid
		ctx.Param["username"] = hub.PublicUsername
	}
}
