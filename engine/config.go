package engine

import "github.com/anacrolix/torrent"

func NewConfig() *torrent.ClientConfig {
	config := torrent.NewDefaultClientConfig()
	config.DataDir = "./download"
	return config
}
