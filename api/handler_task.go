package api

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/serializer"
	"github.com/projectxpolaris/youdownload-server/hub"
	"net/http"
)

var taskInfoHandler haruka.RequestHandler = func(context *haruka.Context) {
	service, err := hub.DefaultHub.GetService(context.Param["uid"].(string))
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	data := serializer.SerializeMultipleTemplate(service.Engine.Pool.Tasks, &BaseTaskTemplate{}, map[string]interface{}{})
	context.JSON(haruka.JSON{
		"list": data,
	})
}

var stopTaskHandler haruka.RequestHandler = func(context *haruka.Context) {
	id := context.GetQueryString("id")
	service, err := hub.DefaultHub.GetService(context.Param["uid"].(string))
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	err = service.Engine.StopTask(id)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}
var startTaskHandler haruka.RequestHandler = func(context *haruka.Context) {
	id := context.GetQueryString("id")
	service, err := hub.DefaultHub.GetService(context.Param["uid"].(string))
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	err = service.Engine.StartTask(id)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

var deleteTask haruka.RequestHandler = func(context *haruka.Context) {
	id := context.GetQueryString("id")
	service, err := hub.DefaultHub.GetService(context.Param["uid"].(string))
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	err = service.Engine.DeleteTask(id)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}
