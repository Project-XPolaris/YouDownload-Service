package api

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/serializer"
	"github.com/projectxpolaris/youdownload-server/engine"
	"net/http"
)

var taskInfoHandler haruka.RequestHandler = func(context *haruka.Context) {
	data := serializer.SerializeMultipleTemplate(engine.DefaultEngine.Pool.Tasks, &BaseTaskTemplate{}, map[string]interface{}{})
	context.JSON(haruka.JSON{
		"list": data,
	})
}

var stopTaskHandler haruka.RequestHandler = func(context *haruka.Context) {
	id := context.GetQueryString("id")
	err := engine.DefaultEngine.StopTask(id)
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
	err := engine.DefaultEngine.StartTask(id)
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
	err := engine.DefaultEngine.DeleteTask(id)
	if err != nil {
		AbortError(context, err, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}
