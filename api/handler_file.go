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
	task := engine.DefaultEngine.CreateDownloadTask(requestBody.Link)
	template := BaseTaskTemplate{}
	err = template.Serializer(task, nil)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
	}
	context.JSON(haruka.JSON{
		"success": true,
		"task":    template,
	})
}
