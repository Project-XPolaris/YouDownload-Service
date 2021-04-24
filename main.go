package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	srv "github.com/kardianos/service"
	"github.com/projectxpolaris/youdownload-server/api"
	"github.com/projectxpolaris/youdownload-server/config"
	"github.com/projectxpolaris/youdownload-server/hub"
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
	err = os.MkdirAll(config.Instance.DownloadDir, os.ModePerm)
	if err != nil {
		Logger.Fatal(err)
	}
	hub.InitHub()
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
	// flags
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		logrus.Fatal(err)
		return
	}
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
