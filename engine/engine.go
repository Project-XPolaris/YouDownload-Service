package engine

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/anacrolix/torrent"
	"github.com/myanimestream/arigo"
	config2 "github.com/projectxpolaris/youdownload-server/config"
	"github.com/projectxpolaris/youdownload-server/database"
	"github.com/sirupsen/logrus"
	"time"
)

var (
	Logger = logrus.WithFields(logrus.Fields{
		"scope": "Engine",
	})
)

type Engine struct {
	TorrentClient *torrent.Client
	Pool          *TaskPool
	TorrentConfig *torrent.ClientConfig
	Database      *Database
	Config        *EngineConfig
	Aria2Client   *arigo.Client
}

func NewEngine(engineConfig *EngineConfig) (*Engine, error) {
	config := NewConfig()
	config.DataDir = engineConfig.DownloadDir
	config.ListenPort = engineConfig.TorrentPort
	client, err := NewClient(config)
	if err != nil {
		return nil, err
	}
	pool := TaskPool{
		Client: client,
		Tasks:  []Task{},
	}
	boltdatabase, err := OpenDatabase(engineConfig.DatabaseDir)
	if err != nil {
		return nil, err
	}
	aria2Client, err := arigo.Dial(config2.Instance.Aria2Url, "")
	if err != nil {
		return nil, err
	}
	if err != nil {
		panic(err)
	}
	engine := &Engine{
		TorrentClient: client,
		Pool:          &pool,
		TorrentConfig: config,
		Database:      boltdatabase,
		Config:        engineConfig,
		Aria2Client:   &aria2Client,
	}
	pool.Engine = engine
	//restore task
	savedTasks, err := boltdatabase.ReadSavedTask()
	if err != nil {
		return nil, err
	}
	Logger.WithField("count", len(savedTasks)).Info("read saved task from boltdatabase")
	for _, savedTask := range savedTasks {
		task, err := pool.newTorrentTaskFromSaveTask(savedTask, engine)
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

	//restore file download task
	var saveFileTask []database.FileTask
	err = database.Instance.Find(&saveFileTask).Error
	if err != nil {
		return nil, err
	}
	for _, fileTask := range saveFileTask {
		gid := engine.Aria2Client.GetGID(fileTask.Gid)
		_, err = gid.GetFiles()
		if err != nil {
			continue
		}
		gidStatus, err := gid.TellStatus(fileTask.Gid, "status", "files")
		if err != nil {
			continue
		}
		downloadTask := DownloadTask{
			TaskId:     fileTask.Id,
			SavePath:   gidStatus.Files[0].Path,
			Status:     Aria2StatusToTaskStatus[gidStatus.Status],
			SaveTask:   &SaveFileDownloadTask{},
			OnPrepare:  make(chan struct{}),
			OnComplete: make(chan struct{}),
			OnStop:     make(chan struct{}),
			CreateTime: time.Now(),
			Gid:        &gid,
		}
		gidUris, err := gid.GetURIs()
		if err != nil {
			continue
		}
		downloadTask.Url = gidUris[0].URI

		engine.Pool.Tasks = append(engine.Pool.Tasks, &downloadTask)
	}
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				for _, task := range engine.Pool.Tasks {
					switch task.(type) {
					case *DownloadTask:
						//save := task.GetSaveTask().(*SaveFileDownloadTask)
						//if save == nil {
						//	continue
						//}
						//save.BytesComplete = task.ByteComplete()
						//save.Length = task.Length()
						//save.Save(engine.Database)
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
	return engine, nil
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
	go task.RunPiecesChangeSub(e)
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
	go task.RunPiecesChangeSub(e)
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

	task := NewDownloadTask(link, e.Config.DownloadDir)
	go func() {
		for {
			select {
			case <-task.OnPrepare:
				saveTask := SaveFileDownloadTask{
					TaskId:     task.TaskId,
					Url:        task.Url,
					SavePath:   task.SavePath,
					Status:     task.Status,
					Name:       task.TaskId,
					CreateTime: task.CreateTime,
				}
				task.SaveTask = &saveTask
				//err := saveTask.Save(e.Database)
				//if err != nil {
				//	Logger.Error(err)
				//}
			case <-task.OnComplete:
				//task.SaveTask.Save(e.Database)
				return
			}
		}

	}()
	go task.Run(e)
	e.Pool.Lock()
	e.Pool.Tasks = append(e.Pool.Tasks, task)
	e.Pool.Unlock()
	return task
}
