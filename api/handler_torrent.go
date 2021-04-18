package api

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youdownload-server/hub"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type NewMagnetTaskRequestBody struct {
	Link string `json:"link"`
}

var createMargetTask haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody NewMagnetTaskRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
	}
	service, err := hub.DefaultHub.GetService(context.Param["uid"].(string))
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
	}
	task, err := service.Engine.CreateMagnetTask(requestBody.Link)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
	}
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	template := BaseTaskTemplate{}
	template.Serializer(task, map[string]interface{}{})

	context.JSON(haruka.JSON{
		"success": true,
		"task":    template,
	})
}

var createTorrentTask haruka.RequestHandler = func(context *haruka.Context) {
	err := context.Request.ParseMultipartForm(32 << 20)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}
	file, handler, err := context.Request.FormFile("file")
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
		return
	}

	defer file.Close()
	service, err := hub.DefaultHub.GetService(context.Param["uid"].(string))
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
	}
	filePath := filepath.Join(service.Engine.Config.TempDir, handler.Filename)
	filePathAbs, _ := filepath.Abs(filePath)

	f, err := os.OpenFile(filePathAbs, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	torrentTask, err := service.Engine.CreateTorrentTask(filePathAbs)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	template := BaseTaskTemplate{}
	template.Serializer(torrentTask, map[string]interface{}{})

	context.JSON(haruka.JSON{
		"success": true,
		"task":    template,
	})
}
