package engine

import (
	"github.com/anacrolix/torrent"
	"github.com/myanimestream/arigo"
	"sync"
	"time"
)

type TaskStatus int

var Aria2StatusToTaskStatus = map[arigo.DownloadStatus]TaskStatus{
	arigo.StatusActive:    Estimate,
	arigo.StatusWaiting:   Downloading,
	arigo.StatusCompleted: Complete,
	arigo.StatusPaused:    Stop,
}

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
	GetCreateTime() time.Time
	GetInfo() interface{}
}
