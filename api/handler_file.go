package api

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youdownload-server/engine"
	"net/http"
)

type NewFileTaskRequestBody struct {
	Link string `json:"link"`
}

var newFileDownloadTask haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody NewFileTaskRequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortError(context, err, http.StatusBadRequest)
	}
	engine.DefaultEngine.CreateDownloadTask(requestBody.Link)
	context.JSON(haruka.JSON{
		"success": true,
	})
}
