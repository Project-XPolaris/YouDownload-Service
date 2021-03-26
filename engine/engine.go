package engine

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/anacrolix/torrent"
	"github.com/cavaliercoder/grab"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"time"
)

var (
	DefaultEngine *Engine
	Logger        = logrus.WithFields(logrus.Fields{
		"scope": "Engine",
	})
)

type Engine struct {
	TorrentClient      *torrent.Client
	FileDownloadClient *grab.Client
	Pool               *TaskPool
	Config             *torrent.ClientConfig
	Database           *Database
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
		TorrentClient:      client,
		FileDownloadClient: grab.NewClient(),
		Pool:               &pool,
		Config:             config,
		Database:           database,
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
	saveFileDownloadTasks, err := database.ReadSavedFileDownloadTask()
	if err != nil {
		return err
	}
	for _, saveFileDownloadTask := range saveFileDownloadTasks {
		task := engine.Pool.newFileTaskFromSaveTask(saveFileDownloadTask)
		engine.Pool.Lock()
		engine.Pool.Tasks = append(engine.Pool.Tasks, task)
		engine.Pool.Unlock()
		if task.Status == Downloading {
			go task.Run(engine)
		}
	}
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				for _, task := range engine.Pool.Tasks {
					switch task.(type) {
					case *DownloadTask:
						save := task.GetSaveTask().(*SaveFileDownloadTask)
						if save == nil {
							continue
						}
						save.BytesComplete = task.ByteComplete()
						save.Length = task.Length()
						save.Save(engine.Database)
					case *TorrentTask:
						save := task.GetSaveTask().(*SavedTorrentTask)
						if save == nil {
							continue
						}
						save.BytesComplete = task.ByteComplete()
						save.Length = task.Length()
						save.Save(engine.Database)
					}
				}
			}
		}

	}()
	Logger.Info("engine init success")
	return nil
}

func (e *Engine) Stop() error {
	e.TorrentClient.Close()
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
		err = task.GetSaveTask().Save(e.Database)
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
		if downloadTask, ok := task.(*DownloadTask); ok {
			err = e.startDownloadTask(downloadTask)
			if err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}
		err = task.GetSaveTask().UpdateTaskStatus(e.Database, Downloading)

		if err != nil {
			return err
		}
	}
	return nil
}
func (e *Engine) startDownloadTask(task *DownloadTask) error {
	go task.Run(e)
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
	err = task.GetSaveTask().RemoveTask(e.Database)
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
func (e *Engine) CreateMagnetTask(link string) (*TorrentTask, error) {
	//new task
	task, err := e.Pool.newTorrentTaskFromMagnetLink(link)
	if err != nil {
		return task, err
	}
	for _, existTask := range e.Pool.Tasks {
		if existTorrentTask, ok := existTask.(*TorrentTask); ok {
			if task.Torrent.InfoHash().String() == existTorrentTask.Torrent.InfoHash().String() {
				return existTorrentTask, nil
			}
		}
	}
	e.Pool.Lock()
	e.Pool.Tasks = append(e.Pool.Tasks, task)
	e.Pool.Unlock()

	// run task
	go task.RunDownloadProgress(e)
	go task.RunPiecesChangeSub()
	go task.RunRateStaticSub()
	return task, err
}
func (e *Engine) CreateTorrentTask(torrentFilePath string) (*TorrentTask, error) {
	task, err := e.Pool.newTorrentTaskFromFile(torrentFilePath)
	if err != nil {
		return nil, err
	}
	for _, existTask := range e.Pool.Tasks {
		if existTorrentTask, ok := existTask.(*TorrentTask); ok {
			if task.Torrent.InfoHash().String() == existTorrentTask.Torrent.InfoHash().String() {
				return existTorrentTask, nil
			}
		}
	}
	e.Pool.Lock()
	e.Pool.Tasks = append(e.Pool.Tasks, task)
	e.Pool.Unlock()
	go task.RunDownloadProgress(e)
	go task.RunPiecesChangeSub()
	go task.RunRateStaticSub()
	return task, err
}

func (e *Engine) CreateDownloadTask(link string) Task {
	for _, task := range e.Pool.Tasks {
		if fileDownloadTask, ok := task.(*DownloadTask); ok {
			if fileDownloadTask.Url == link {
				return task
			}
		}
	}
	task := NewDownloadTask(link)
	go func() {
		<-task.OnPrepare
		saveTask := SaveFileDownloadTask{
			TaskId:     task.TaskId,
			Url:        task.Url,
			SavePath:   task.SavePath,
			Status:     task.Status,
			Name:       filepath.Base(task.Response.Filename),
			CreateTime: task.CreateTime,
		}
		task.SaveTask = &saveTask
		err := saveTask.Save(e.Database)
		if err != nil {
			Logger.Error(err)
		}
	}()
	go task.Run(e)
	e.Pool.Lock()
	e.Pool.Tasks = append(e.Pool.Tasks, task)
	e.Pool.Unlock()
	return task
}
