package api

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youdownload-server/config"
	"github.com/projectxpolaris/youdownload-server/service"
	"net/http"
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
	if len(requestBody.Path) == 0 {
		homePath, err := filepath.Abs(config.Instance.DownloadDir)
		if err != nil {
			AbortError(context, err, http.StatusBadRequest)
			return
		}
		requestBody.Path = homePath
	}
	items, err := service.ReadDirectory(requestBody.Path)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	abs, _ := filepath.Abs(requestBody.Path)
	context.JSON(map[string]interface{}{
		"path":  abs,
		"sep":   string(filepath.Separator),
		"files": items,
	})
}