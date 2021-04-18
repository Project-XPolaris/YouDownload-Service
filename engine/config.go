package engine

import (
	"github.com/anacrolix/torrent"
	"github.com/projectxpolaris/youdownload-server/config"
)

type EngineConfig struct {
	DatabaseDir string
	DownloadDir string
	TempDir     string
	TorrentPort int
}

func NewConfig() *torrent.ClientConfig {
	torrentClientConfig := torrent.NewDefaultClientConfig()
	torrentClientConfig.DataDir = config.Instance.DownloadDir
	return torrentClientConfig
}
