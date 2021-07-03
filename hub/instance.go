package hub

import (
	"errors"
	"github.com/projectxpolaris/youdownload-server/database"
	"github.com/projectxpolaris/youdownload-server/engine"
	"os"
	"path"
	"sync"
)

var (
	DefaultHub *DownloadHub
)

func InitHub() {
	DefaultHub = &DownloadHub{
		Services:    []*DownloadService{},
		TorrentPort: 42069,
	}
}

type DownloadService struct {
	Uid    string
	Engine *engine.Engine
}
type DownloadHub struct {
	Services    []*DownloadService
	TorrentPort int
	sync.Mutex
}

func (h *DownloadHub) createService(uid string) (*DownloadService, error) {
	var user database.User
	err := database.Instance.Where("uid = ?", uid).Find(&user).Error
	if err != nil {
		return nil, err
	}
	if len(user.DataPath) == 0 {
		return nil, errors.New("user not init")
	}
	dataPath := path.Join(user.DataPath, "download")
	err = os.MkdirAll(dataPath, os.ModePerm)
	if err != nil {
		return nil, err
	}
	tempPath := path.Join(user.DataPath, "tmp")
	err = os.MkdirAll(tempPath, os.ModePerm)
	if err != nil {
		return nil, err
	}
	h.Lock()
	torrentPort := h.TorrentPort + 1
	h.TorrentPort += 1
	engineConfig := &engine.EngineConfig{
		DatabaseDir: user.DataPath,
		DownloadDir: dataPath,
		TempDir:     tempPath,
		TorrentPort: torrentPort,
	}
	h.Unlock()
	e, err := engine.NewEngine(engineConfig)
	if err != nil {
		return nil, err
	}
	h.Lock()
	defer h.Unlock()
	service := &DownloadService{
		Uid:    uid,
		Engine: e,
	}
	h.Services = append(h.Services, service)
	return service, nil
}
func (h *DownloadHub) GetService(uid string) (*DownloadService, error) {
	for _, service := range h.Services {
		if service.Uid == uid {
			return service, nil
		}
	}
	return h.createService(uid)
}
