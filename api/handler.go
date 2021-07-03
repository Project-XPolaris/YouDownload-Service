package api

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youdownload-server/config"
	"github.com/projectxpolaris/youdownload-server/hub"
	"github.com/projectxpolaris/youdownload-server/service"
	"github.com/projectxpolaris/youdownload-server/youplus"
	"net/http"
	"os"
	"path/filepath"
)

type ReadDirectoryRequestBody struct {
	Path string `json:"path"`
}

var readDirectoryHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody ReadDirectoryRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	if config.Instance.PathEnable {
		token := context.Param["token"].(string)
		items, err := youplus.DefaultClient.ReadDir(requestBody.Path, token)
		if err != nil {
			AbortError(context, err, http.StatusInternalServerError)
			return
		}
		data := make([]BaseFileItemTemplate, 0)
		for _, item := range items {
			template := BaseFileItemTemplate{}
			template.AssignWithYouPlusItem(item)
			data = append(data, template)
		}
		context.JSON(haruka.JSON{
			"path":  requestBody.Path,
			"sep":   "/",
			"files": data,
			"back":  filepath.Dir(requestBody.Path),
		})
		return
	} else {
		infos, err := service.ReadDirectory(requestBody.Path)
		if err != nil {
			AbortError(context, err, http.StatusInternalServerError)
			return
		}
		data := make([]BaseFileItemTemplate, 0)
		for _, info := range infos {
			template := BaseFileItemTemplate{}
			template.Assign(info, requestBody.Path)
			data = append(data, template)
		}
		context.JSON(haruka.JSON{
			"path":  requestBody.Path,
			"sep":   string(os.PathSeparator),
			"files": data,
			"back":  filepath.Dir(requestBody.Path),
		})
	}
}
var initEngineHandler haruka.RequestHandler = func(context *haruka.Context) {
	uid := context.GetQueryString("uid")
	_, err := hub.DefaultHub.GetService(uid)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}
var serviceInfoHandler haruka.RequestHandler = func(context *haruka.Context) {
	context.JSON(haruka.JSON{
		"name":       "YouDownload serivce",
		"authEnable": config.Instance.AuthEnable,
	})
}
