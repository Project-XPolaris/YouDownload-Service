package api

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youdownload-server/hub"
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
	service, err := hub.DefaultHub.GetService(context.Param["uid"].(string))
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
	}
	task := service.Engine.CreateDownloadTask(requestBody.Link)
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
