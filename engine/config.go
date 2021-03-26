package engine

import (
	"github.com/anacrolix/torrent"
	"github.com/projectxpolaris/youdownload-server/config"
)

func NewConfig() *torrent.ClientConfig {
	torrentClientConfig := torrent.NewDefaultClientConfig()
	torrentClientConfig.DataDir = config.Instance.DownloadDir
	return torrentClientConfig
}
