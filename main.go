package main

import (
	"context"
	"fmt"
	srv "github.com/kardianos/service"
	"github.com/project-xpolaris/youplustoolkit/util"
	entry "github.com/project-xpolaris/youplustoolkit/youplus/entity"
	"github.com/projectxpolaris/youdownload-server/api"
	"github.com/projectxpolaris/youdownload-server/config"
	"github.com/projectxpolaris/youdownload-server/database"
	"github.com/projectxpolaris/youdownload-server/hub"
	"github.com/projectxpolaris/youdownload-server/youplus"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var Logger = logrus.WithFields(logrus.Fields{
	"scope": "Main",
})

var svcConfig *srv.Config

func initService(workDir string) error {
	svcConfig = &srv.Config{
		Name:             "YouDownloadService",
		DisplayName:      "YouDownload Core Service",
		WorkingDirectory: workDir,
		Arguments:        []string{"run"},
	}
	return nil
}
func Program() {
	err := config.ReadConfig()
	if err != nil {
		Logger.Fatal(err)
	}
	// init youplus client
	if config.Instance.AuthEnable || config.Instance.PathEnable {
		Logger.Info("init youplus [check]")
		youplus.DefaultClient.Init(config.Instance.YouPlusUrl)
		info, err := youplus.DefaultClient.FetchInfo()
		if err != nil {
			Logger.Fatal(err)
		}
		if !info.Success {
			Logger.Fatal("init youplus [failed]")
		}
		Logger.Info("init youplus [pass]")
	}
	// init database
	Logger.Info("connect to database")
	err = database.Connect()
	if err != nil {
		Logger.Fatal(err)
	}
	err = os.MkdirAll(config.Instance.DownloadDir, os.ModePerm)
	if err != nil {
		Logger.Fatal(err)
	}
	hub.InitHub()
	if len(config.Instance.YouPlusRPCAddr) > 0 {
		Logger.Info("check youplus rpc [checking]")
		err = youplus.LoadYouPlusRPCClient()
		if err != nil {
			Logger.WithFields(logrus.Fields{
				"url": config.Instance.YouPlusRPCAddr,
			}).Fatal(err.Error())
		}

		Logger.WithFields(logrus.Fields{
			"url": config.Instance.YouPlusRPCAddr,
		}).Info("check youplus rpc service [pass]")

	}
	// youplus entity
	if config.Instance.Entity.Enable {
		Logger.Info("register entity")
		youplus.InitEntity()

		err := youplus.DefaultEntry.Register()
		if err != nil {
			Logger.Fatal(err.Error())
		}

		addrs, err := util.GetHostIpList()
		urls := make([]string, 0)
		for _, addr := range addrs {
			urls = append(urls, fmt.Sprintf("http://%s%s", addr, config.Instance.Addr))
		}
		if err != nil {
			Logger.Fatal(err.Error())
		}
		err = youplus.DefaultEntry.UpdateExport(entry.EntityExport{Urls: urls, Extra: map[string]interface{}{}})
		if err != nil {
			Logger.Fatal(err.Error())
		}

		err = youplus.DefaultEntry.StartHeartbeat(context.Background())
		if err != nil {
			Logger.Fatal(err.Error())
		}
		Logger.WithFields(logrus.Fields{
			"url": config.Instance.YouPlusRPCAddr,
		}).Info("success register entity")

	}
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go api.RunApiApplication()
	Logger.Info("application running")
	<-done
	Logger.Info("graceful shutdown")
}

type program struct{}

func (p *program) Start(s srv.Service) error {
	go Program()
	return nil
}

func (p *program) Stop(s srv.Service) error {
	return nil
}

func InstallAsService() {
	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	s.Uninstall()

	err = s.Install()
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("successful install service")
}

func UnInstall() {

	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	s.Uninstall()
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("successful uninstall service")
}

var opts struct {
	Install   bool `short:"i" long:"install" description:"install service"`
	Uninstall bool `short:"u" long:"uninstall" description:"uninstall service"`
}

func StartService() {
	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	err = s.Start()
	if err != nil {
		logrus.Fatal(err)
	}
}
func StopService() {
	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	err = s.Stop()
	if err != nil {
		logrus.Fatal(err)
	}
}
func RestartService() {
	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	err = s.Restart()
	if err != nil {
		logrus.Fatal(err)
	}
}
func RunApp() {
	app := &cli.App{
		Flags: []cli.Flag{},
		Commands: []*cli.Command{
			{
				Name:  "service",
				Usage: "service manager",
				Subcommands: []*cli.Command{
					{
						Name:  "install",
						Usage: "install service",
						Action: func(context *cli.Context) error {
							InstallAsService()
							return nil
						},
					},
					{
						Name:  "uninstall",
						Usage: "uninstall service",
						Action: func(context *cli.Context) error {
							UnInstall()
							return nil
						},
					},
					{
						Name:  "start",
						Usage: "start service",
						Action: func(context *cli.Context) error {
							StartService()
							return nil
						},
					},
					{
						Name:  "stop",
						Usage: "stop service",
						Action: func(context *cli.Context) error {
							StopService()
							return nil
						},
					},
					{
						Name:  "restart",
						Usage: "restart service",
						Action: func(context *cli.Context) error {
							RestartService()
							return nil
						},
					},
				},
				Description: "YouDownload service controller",
			},
			{
				Name:  "run",
				Usage: "run app",
				Action: func(context *cli.Context) error {
					Program()
					return nil
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
func main() {
	// service
	workPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		logrus.Fatal(err)
	}
	err = initService(workPath)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Info(fmt.Sprintf("work_path =  %s", workPath))
	RunApp()
}
