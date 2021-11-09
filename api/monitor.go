package api

import (
	"context"
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/serializer"
	"github.com/projectxpolaris/youdownload-server/hub"
	"time"
)

func RunMonitor() {
	go func() {
		for {
			if hub.DefaultHub.Services != nil {
				for _, item := range hub.DefaultHub.Services {
					data := serializer.SerializeMultipleTemplate(item.Engine.Pool.Tasks, &BaseTaskTemplate{}, map[string]interface{}{})
					DefaultNotificationManager.sendJSONToUser(haruka.JSON{
						"event": EventTaskStatus,
						"data":  data,
					}, item.Uid)
				}
			}
			<-time.After(2 * time.Second)
		}
	}()
	<-context.Background().Done()
}
