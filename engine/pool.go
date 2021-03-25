package engine

import (
	"context"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/cavaliercoder/grab"
	"github.com/rs/xid"
	"path/filepath"
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
	GetSaveTask() SaveTask
}
type TorrentTask struct {
	TaskId    string
	Torrent   *torrent.Torrent
	Status    TaskStatus
	Speed     int64
	SavedTask *SavedTorrentTask
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
			savedTask := NewSavedTask(t.TaskId, t.Torrent.Metainfo(), Downloading, t.Name())
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

func (p *TaskPool) newTorrentTaskFromSaveTask(savedTask *SavedTorrentTask) (*TorrentTask, error) {
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

type DownloadTask struct {
	TaskId    string
	Request   *grab.Request
	Response  *grab.Response
	Url       string
	SavePath  string
	Cancel    context.CancelFunc
	Status    TaskStatus
	SaveTask  *SaveFileDownloadTask
	OnPrepare chan struct{}
}

func (t *DownloadTask) GetSaveTask() SaveTask {
	return t.SaveTask
}

func (t *DownloadTask) Id() string {
	return t.TaskId
}

func (t *DownloadTask) Name() string {
	if t.Response != nil {
		return filepath.Base(t.Response.Filename)
	} else if t.SaveTask != nil {
		return t.SaveTask.Name
	}
	return t.Id()
}

func (t *DownloadTask) ByteComplete() int64 {
	if t.Response != nil {
		return t.Response.BytesComplete()
	}
	return 0
}

func (t *DownloadTask) Length() int64 {
	if t.Response != nil {
		return t.Response.Size
	}
	return 0
}

func (t *DownloadTask) Start() error {
	return nil
}

func (t *DownloadTask) Stop() error {
	t.Cancel()
	return nil
}

func (t *DownloadTask) Delete() error {
	return nil
}

func (t *DownloadTask) GetSpeed() int64 {
	if t.Request != nil {
		return int64(t.Response.BytesPerSecond())
	}
	return 0
}

func (t *DownloadTask) TaskStatus() TaskStatus {
	return t.Status
}

func (t *DownloadTask) SavedTaskId() int {
	return t.SaveTask.ID
}

func NewDownloadTask(link string) *DownloadTask {
	return &DownloadTask{
		TaskId:    xid.New().String(),
		SavePath:  "./download",
		Url:       link,
		Status:    Downloading,
		OnPrepare: make(chan struct{}),
	}
}

func (t *DownloadTask) Run(e *Engine) {
	request, err := grab.NewRequest(t.SavePath, t.Url)
	if err != nil {
		fmt.Println(err)
		return
	}
	request.BufferSize = 1024
	ctx, cancel := context.WithCancel(context.Background())
	t.Cancel = cancel
	request = request.WithContext(ctx)
	Logger.WithField("id", t.Id).WithField("url", request.URL()).Info("Downloading")
	t.Request = request
	// request download url
	response := e.FileDownloadClient.Do(request)
	t.Response = response
	t.OnPrepare <- struct{}{}
	// update with request result
	//run for done chan
	go func() {
		select {
		case <-response.Done:
			t.Status = Complete
			Logger.WithField("id", t.Id).Info("task complete")
		case <-ctx.Done():
			t.Status = Stop
			Logger.WithField("id", t.Id).Info("task interrupt")
		}
	}()
}

func (p *TaskPool) newFileTaskFromSaveTask(savedTask *SaveFileDownloadTask) *DownloadTask {
	return &DownloadTask{
		TaskId:   savedTask.TaskId,
		SavePath: savedTask.SavePath,
		Url:      savedTask.Url,
		Status:   savedTask.Status,
		SaveTask: savedTask,
	}
}
