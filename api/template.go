package api

import (
	"github.com/projectxpolaris/youdownload-server/engine"
)

var taskStatusMapping map[engine.TaskStatus]string = map[engine.TaskStatus]string{
	engine.Estimate:    "Estimate",
	engine.Downloading: "Downloading",
	engine.Stop:        "Stop",
	engine.Complete:    "Complete",
}

type BaseTaskTemplate struct {
	Id       string      `json:"id"`
	Name     string      `json:"name"`
	Complete int64       `json:"complete"`
	Length   int64       `json:"length"`
	Progress float64     `json:"progress"`
	Status   string      `json:"status"`
	Speed    int64       `json:"speed"`
	ETA      int64       `json:"eta"`
	Extra    interface{} `json:"extra"`
	Type     string      `json:"type"`
}

func (t *BaseTaskTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	task := dataModel.(engine.Task)
	t.Name = task.Name()
	t.Complete = task.ByteComplete()
	t.Length = task.Length()
	t.Id = task.Id()
	if t.Length > 0 {
		t.Progress = float64(t.Complete) / float64(t.Length)
	}
	t.Status = taskStatusMapping[task.TaskStatus()]
	t.Speed = task.GetSpeed()
	if t.Speed != 0 {
		t.ETA = task.Length() / t.Speed
	}
	if torrentTask, ok := task.(*engine.TorrentTask); ok {
		extra := TorrentTaskExtra{}
		extra.Assign(torrentTask)
		t.Extra = extra
	}
	switch task.(type) {
	case *engine.TorrentTask:
		t.Type = "Torrent"
	case *engine.DownloadTask:
		t.Type = "Download"
	}
	return nil
}

type TorrentTaskExtra struct {
	Files []TorrentFile `json:"files"`
	Peer  []TorrentPeer `json:"peer"`
}

type TorrentFile struct {
	Name     string `json:"name"`
	Length   int64  `json:"length"`
	Path     string `json:"path"`
	Priority int    `json:"priority"`
}

type TorrentPeer struct {
	ClientName string `json:"clientName"`
	Network    string `json:"network"`
	Remote     string `json:"remote"`
	Port       int    `json:"port"`
}

func (t *TorrentTaskExtra) Assign(task *engine.TorrentTask) {
	if task.Torrent == nil {
		return
	}
	if task.Torrent.Files() != nil {
		t.Files = []TorrentFile{}
		for _, file := range task.Torrent.Files() {
			file := TorrentFile{
				Name:     file.DisplayPath(),
				Length:   file.Length(),
				Path:     file.Path(),
				Priority: int(file.Priority()),
			}
			t.Files = append(t.Files, file)
		}
	}
	t.Peer = []TorrentPeer{}
	for _, conn := range task.Torrent.PeerConns() {
		peer := TorrentPeer{
			ClientName: conn.PeerClientName,
			Network:    conn.Network,
			Remote:     conn.RemoteAddr.String(),
			Port:       conn.PeerListenPort,
		}
		t.Peer = append(t.Peer, peer)
	}
}
