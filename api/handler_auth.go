package api

import (
	"errors"
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youdownload-server/config"
	"github.com/projectxpolaris/youdownload-server/hub"
	"github.com/projectxpolaris/youdownload-server/service"
	"github.com/projectxpolaris/youdownload-server/youplus"
	"net/http"
)

type UserAuthRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var LoginHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody UserAuthRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	// public account
	if !config.Instance.AuthEnable {
		needInit, err := service.CheckNeedInit(hub.PublicUid)
		if err != nil {
			AbortError(context, err, http.StatusInternalServerError)
		}
		context.JSON(haruka.JSON{
			"username": requestBody.Username,
			"needInit": needInit,
		})
		return
	}
	auth, err := youplus.DefaultClient.FetchUserAuth(requestBody.Username, requestBody.Password)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	if !auth.Success {
		AbortError(context, errors.New("user auth failed"), http.StatusBadRequest)
		return
	}
	needInit, err := service.CheckNeedInit(auth.Uid)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"token":    auth.Token,
		"needInit": needInit,
		"success":  true,
	})

}

type InitUserRequestBody struct {
	DataPath string `json:"dataPath"`
}

var InitUser haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody InitUserRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	dataPath := requestBody.DataPath
	if config.Instance.PathEnable {
		realPath, err := youplus.DefaultClient.GetRealPath(dataPath, context.Param["token"].(string))
		if err != nil {
			AbortError(context, err, http.StatusBadRequest)
			return
		}
		dataPath = realPath
	}
	err = service.InitUser(context.Param["uid"].(string), dataPath)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}
