package api

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youdownload-server/engine"
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
	err = engine.DefaultEngine.CreateMagnetTask(requestBody.Link)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
	}
	context.JSON(haruka.JSON{
		"success": true,
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

	filePath := filepath.Join("./tmp", handler.Filename)
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
	err = engine.DefaultEngine.CreateTorrentTask(filePathAbs)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}
