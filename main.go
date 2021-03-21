package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	srv "github.com/kardianos/service"
	"github.com/projectxpolaris/youdownload-server/api"
	"github.com/projectxpolaris/youdownload-server/engine"
	"github.com/sirupsen/logrus"
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
	}
	return nil
}
func Program() {
	err := engine.NewEngine()
	if err != nil {
		Logger.Fatal(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go api.RunApiApplication()
	Logger.Info("application running")
	<-done
	Logger.Info("graceful shutdown")
	err = engine.DefaultEngine.Stop()
	if err != nil {
		Logger.Error(err)
	}
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
	if opts.Install {
		InstallAsService()
		return
	}
	if opts.Uninstall {
		UnInstall()
		return
	}
	Program()
}
