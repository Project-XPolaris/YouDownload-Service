package api

import "github.com/allentom/haruka"

func AbortError(context *haruka.Context, err error, status int) {
	context.JSONWithStatus(haruka.JSON{
		"success": false,
		"reason":  err.Error(),
	}, status)
}
