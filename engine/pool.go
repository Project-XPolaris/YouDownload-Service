package engine

import (
	"github.com/anacrolix/torrent"
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
	GetCreateTime() time.Time
	GetInfo() interface{}
}
