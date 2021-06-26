package api

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youdownload-server/config"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

var (
	Logger = logrus.WithFields(logrus.Fields{
		"scope": "Api",
	})
)

func RunApiApplication() {
	Logger.Info("Start api service")
	e := haruka.NewEngine()
	e.UseCors(cors.AllowAll())
	e.Router.GET("/tasks", taskInfoHandler)
	e.Router.POST("/task/magnet", createMargetTask)
	e.Router.POST("/task/file", createTorrentTask)
	e.Router.POST("/task/download/file", newFileDownloadTask)
	e.Router.POST("/task/start", startTaskHandler)
	e.Router.POST("/task/stop", stopTaskHandler)
	e.Router.POST("/task/delete", deleteTask)
	e.Router.POST("/util/readDir", readDirectoryHandler)
	e.Router.GET("/info", serviceInfoHandler)
	e.Router.POST("/hub/init", initEngineHandler)
	e.Router.POST("/user/auth", LoginHandler)
	e.Router.POST("/user/init", InitUser)
	e.UseMiddleware(&AuthMiddleware{})
	e.RunAndListen(config.Instance.Addr)
}
