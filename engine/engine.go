package engine

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/anacrolix/torrent"
	"github.com/sirupsen/logrus"
)

var (
	DefaultEngine *Engine
	Logger        = logrus.WithFields(logrus.Fields{
		"scope": "Engine",
	})
)

type Engine struct {
	Client   *torrent.Client
	Pool     *TaskPool
	Config   *torrent.ClientConfig
	Database *Database
}

func NewEngine() error {
	config := NewConfig()
	client, err := NewClient(config)
	if err != nil {
		return err
	}
	pool := TaskPool{
		Client: client,
		Tasks:  []Task{},
	}
	database, err := OpenDatabase()
	if err != nil {
		return err
	}
	engine := &Engine{
		Client:   client,
		Pool:     &pool,
		Config:   config,
		Database: database,
	}
	pool.Engine = engine
	DefaultEngine = engine

	//restore task
	savedTasks, err := database.ReadSavedTask()
	if err != nil {
		return err
	}
	Logger.WithField("count", len(savedTasks)).Info("read saved task from database")
	for _, savedTask := range savedTasks {
		task, err := pool.newTorrentTaskFromSaveTask(savedTask)
		if err != nil {
			Logger.WithFields(logrus.Fields{
				"id":  task.TaskId,
				"err": err.Error(),
			}).Errorf("restore task error")
			continue
		}
		Logger.WithFields(logrus.Fields{
			"id":   task.TaskId,
			"name": task.Name(),
		}).Info("restore task success")
	}
	Logger.Info("engine init success")
	return nil
}

func (e *Engine) Stop() error {
	e.Client.Close()
	err := e.Database.DB.Close()
	if err != nil {
		return err
	}
	return nil
}

func (e *Engine) StopTask(id string) error {
	task := e.Pool.FindTaskById(id)
	if task != nil {
		err := task.Stop()
		if err != nil {
			return err
		}
		err = e.Database.UpdateTaskStatus(&SavedTask{ID: task.SavedTaskId()}, Stop)
		if err != nil {
			return err
		}
	}

	return nil
}
func (e *Engine) StartTask(id string) error {
	task := e.Pool.FindTaskById(id)
	if task != nil {
		err := task.Start()
		if err != nil {
			return err
		}
		err = e.Database.UpdateTaskStatus(&SavedTask{ID: task.SavedTaskId()}, Downloading)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) DeleteTask(id string) error {
	task := e.Pool.FindTaskById(id)
	if task == nil {
		return nil
	}
	err := task.Delete()
	if err != nil {
		return err
	}
	err = e.Database.RemoveTask(id)
	if err != nil {
		return err
	}
	e.Pool.Lock()
	defer e.Pool.Unlock()
	linq.From(e.Pool.Tasks).Where(func(i interface{}) bool {
		return i.(Task).Id() != id
	}).ToSlice(&e.Pool.Tasks)
	return nil
}
func (e *Engine) CreateMagnetTask(link string) error {
	//new task
	task, err := e.Pool.newTorrentTaskFromMagnetLink(link)
	if err != nil {
		return err
	}
	e.Pool.Lock()
	e.Pool.Tasks = append(e.Pool.Tasks, task)
	e.Pool.Unlock()

	// run task
	go task.RunDownloadProgress(e)
	go task.RunPiecesChangeSub()
	go task.RunRateStaticSub()
	return nil
}
func (e *Engine) CreateTorrentTask(torrentFilePath string) error {
	task, err := e.Pool.newTorrentTaskFromFile(torrentFilePath)
	if err != nil {
		return err
	}
	e.Pool.Lock()
	e.Pool.Tasks = append(e.Pool.Tasks, task)
	e.Pool.Unlock()
	go task.RunDownloadProgress(e)
	go task.RunPiecesChangeSub()
	go task.RunRateStaticSub()
	return nil
}
