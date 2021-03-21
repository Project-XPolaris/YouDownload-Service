package engine

import "github.com/anacrolix/torrent"

func NewClient(config *torrent.ClientConfig) (*torrent.Client, error) {
	return torrent.NewClient(config)
}
