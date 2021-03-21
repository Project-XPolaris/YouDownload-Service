package engine

import (
	"github.com/anacrolix/torrent"
	"github.com/rs/xid"
	"sync"
	"time"
)

type TaskStatus int

const (
	Estimate TaskStatus = iota + 1
	Downloading
	Stop
	Complete
)

type TaskPool struct {
	Client *torrent.Client
	Tasks  []Task
	Engine *Engine
	sync.Mutex
}

func (p *TaskPool) FindTaskById(id string) Task {
	for _, task := range p.Tasks {
		if task.Id() == id {
			return task
		}
	}
	return nil
}

type Task interface {
	Id() string
	Name() string
	ByteComplete() int64
	Length() int64
	Start() error
	Stop() error
	Delete() error
	GetSpeed() int64
	TaskStatus() TaskStatus
	SavedTaskId() int
}
type TorrentTask struct {
	TaskId    string
	Torrent   *torrent.Torrent
	Status    TaskStatus
	Speed     int64
	SavedTask *SavedTask
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
	return nil
}

func (t *TorrentTask) Stop() error {
	if t.Status != Downloading {
		return nil
	}
	t.Torrent.DisallowDataDownload()
	t.Status = Stop
	return nil
}

func (t *TorrentTask) Length() int64 {
	if t.Torrent.Info() != nil {
		return t.Torrent.Length()
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

func (t *TorrentTask) RunPiecesChangeSub() {
	sub := t.Torrent.SubscribePieceStateChanges()
	for {
		<-sub.Values
		if t.Status != Downloading {
			continue
		}
		if t.Torrent.BytesCompleted() == t.Torrent.Length() {
			t.Status = Complete
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
			savedTask := NewSavedTask(t.TaskId, t.Torrent.Metainfo(), Downloading)
			err := engine.Database.Save(savedTask)
			if err != nil {
				Logger.Error(err)
			}
			t.SavedTask = savedTask
		}
		err := engine.Database.UpdateTaskStatus(t.SavedTask, Downloading)
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
		TaskId: id,
		Status: Estimate,
	}
	t, err := p.Client.AddMagnet(link)
	task.Torrent = t
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (p *TaskPool) newTorrentTaskFromSaveTask(savedTask *SavedTask) (*TorrentTask, error) {
	t, err := p.Client.AddTorrent(savedTask.MetaInfo)
	if err != nil {
		return nil, err
	}
	task := &TorrentTask{
		TaskId:    savedTask.TaskId,
		Torrent:   t,
		Status:    savedTask.Status,
		SavedTask: savedTask,
	}
	go task.RunDownloadProgress(p.Engine)
	go task.RunPiecesChangeSub()
	go task.RunRateStaticSub()
	p.Lock()
	p.Tasks = append(p.Tasks, task)
	p.Unlock()
	return task, nil
}

func (p *TaskPool) newTorrentTaskFromFile(filePath string) (*TorrentTask, error) {
	id := xid.New().String()
	task := &TorrentTask{
		TaskId: id,
		Status: Estimate,
	}
	t, err := p.Client.AddTorrentFromFile(filePath)
	task.Torrent = t
	if err != nil {
		return nil, err
	}
	return task, nil
}
