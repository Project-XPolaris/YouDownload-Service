package engine

import (
	"github.com/anacrolix/torrent"
	"github.com/rs/xid"
	"path/filepath"
	"time"
)

type TorrentTask struct {
	TaskId     string
	Torrent    *torrent.Torrent
	Status     TaskStatus
	Speed      int64
	SavedTask  *SavedTorrentTask
	CreateTime time.Time
	OnComplete chan struct{}
}
type TorrentFile struct {
	Path   string `json:"path"`
	Name   string `json:"name"`
	Length int64  `json:"length"`
}
type TorrentInfo struct {
	Files []*TorrentFile `json:"files"`
}

func (t *TorrentTask) GetCreateTime() time.Time {
	return t.CreateTime
}

func (t *TorrentTask) GetSaveTask() SaveTask {
	return t.SavedTask
}

func (t *TorrentTask) SavedTaskId() int {
	return t.SavedTask.ID
}

func (t *TorrentTask) GetSpeed() int64 {
	return t.Speed
}

func (t *TorrentTask) Delete() error {
	t.Torrent.Drop()
	return nil
}

func (t *TorrentTask) TaskStatus() TaskStatus {
	return t.Status
}

func (t *TorrentTask) Start() error {
	if t.Status != Stop {
		return nil
	}
	t.Torrent.AllowDataDownload()
	t.Status = Downloading
	t.SavedTask.Status = Downloading
	return nil
}

func (t *TorrentTask) Stop() error {
	if t.Status != Downloading {
		return nil
	}
	t.Torrent.DisallowDataDownload()
	t.Status = Stop
	t.SavedTask.Status = Stop
	return nil
}

func (t *TorrentTask) Length() int64 {
	if t.Torrent.Info() != nil {
		return t.Torrent.Length()
	} else if t.SavedTask != nil {
		return t.SavedTask.Length
	}
	return 0
}

func (t *TorrentTask) Id() string {
	return t.TaskId
}

func (t *TorrentTask) Name() string {
	return t.Torrent.Name()
}

func (t *TorrentTask) ByteComplete() int64 {
	return t.Torrent.BytesCompleted()
}
func (t *TorrentTask) GetInfo() interface{} {
	info := t.Torrent.Info()
	if info == nil {
		return map[string]interface{}{}
	}
	torrentInfo := TorrentInfo{Files: []*TorrentFile{}}
	for _, file := range info.Files {
		torrentInfo.Files = append(torrentInfo.Files, &TorrentFile{Name: filepath.Base(file.Path[0]), Path: file.Path[0], Length: file.Length})
	}
	return torrentInfo
}
func (t *TorrentTask) RunPiecesChangeSub(engine *Engine) {
	sub := t.Torrent.SubscribePieceStateChanges()
	for {
		<-sub.Values
		if t.Status != Downloading {
			continue
		}
		if t.Torrent.BytesCompleted() == t.Torrent.Length() {
			t.Status = Complete
			t.OnComplete <- struct{}{}
			t.SavedTask.Status = Complete
			t.SavedTask.Save(engine.Database)
			return
		}
	}
}
func (t *TorrentTask) RunRateStaticSub() {
	lastByteComplete := t.Torrent.BytesCompleted()
	for {
		<-time.After(1 * time.Second)
		t.Speed = t.Torrent.BytesCompleted() - lastByteComplete
		lastByteComplete = t.Torrent.BytesCompleted()
	}
}
func (t *TorrentTask) RunDownloadProgress(engine *Engine) {
	if t.Status == Estimate {
		<-t.Torrent.GotInfo()
		if t.SavedTask == nil {
			savedTask := NewSavedTask(t.TaskId, t.Torrent.Metainfo(), Downloading, t.Name(), t.CreateTime)
			err := savedTask.Save(engine.Database)
			if err != nil {
				Logger.Error(err)
			}
			t.SavedTask = savedTask
		}
		err := t.SavedTask.UpdateTaskStatus(engine.Database, Downloading)
		if err != nil {
			Logger.Error(err)
		}
		t.Torrent.DownloadAll()
		t.Status = Downloading
		return
	}

	if t.Status == Downloading {
		<-t.Torrent.GotInfo()
		t.Torrent.DownloadAll()
		return
	}

	if t.Status == Stop {
		t.Torrent.DownloadAll()
		t.Torrent.DisallowDataDownload()
	}

	if t.Status == Complete {
		return
	}
}

func (p *TaskPool) newTorrentTaskFromMagnetLink(link string) (*TorrentTask, error) {
	id := xid.New().String()
	task := &TorrentTask{
		TaskId:     id,
		Status:     Estimate,
		CreateTime: time.Now(),
		OnComplete: make(chan struct{}),
	}
	t, err := p.Client.AddMagnet(link)
	task.Torrent = t
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (p *TaskPool) newTorrentTaskFromSaveTask(savedTask *SavedTorrentTask, engine *Engine) (*TorrentTask, error) {
	t, err := p.Client.AddTorrent(savedTask.MetaInfo)
	if err != nil {
		return nil, err
	}
	task := &TorrentTask{
		TaskId:     savedTask.TaskId,
		Torrent:    t,
		Status:     savedTask.Status,
		SavedTask:  savedTask,
		CreateTime: time.Now(),
		OnComplete: make(chan struct{}),
	}
	go task.RunDownloadProgress(p.Engine)
	go task.RunPiecesChangeSub(engine)
	go task.RunRateStaticSub()
	p.Lock()
	p.Tasks = append(p.Tasks, task)
	p.Unlock()
	return task, nil
}

func (p *TaskPool) newTorrentTaskFromFile(filePath string) (*TorrentTask, error) {
	id := xid.New().String()
	task := &TorrentTask{
		TaskId:     id,
		Status:     Estimate,
		CreateTime: time.Now(),
		OnComplete: make(chan struct{}),
	}
	t, err := p.Client.AddTorrentFromFile(filePath)
	task.Torrent = t
	if err != nil {
		return nil, err
	}
	return task, nil
}
