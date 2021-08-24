package engine

import (
	"errors"
	"fmt"
	"github.com/myanimestream/arigo"
	"github.com/projectxpolaris/youdownload-server/database"
	"github.com/rs/xid"
	"path/filepath"
	"time"
)

type DownloadTask struct {
	TaskId     string
	Url        string
	SavePath   string
	Status     TaskStatus
	SaveTask   *SaveFileDownloadTask
	OnPrepare  chan struct{}
	OnComplete chan struct{}
	OnStop     chan struct{}
	CreateTime time.Time
	Gid        *arigo.GID
}

func (t *DownloadTask) GetCreateTime() time.Time {
	return t.CreateTime
}

func (t *DownloadTask) GetSaveTask() SaveTask {
	return t.SaveTask
}

func (t *DownloadTask) Id() string {
	return t.TaskId
}

func (t *DownloadTask) Name() string {
	if t.Gid == nil {
		return t.TaskId
	}
	files, err := t.Gid.GetFiles()
	if err != nil || len(files) == 0 {
		return t.Id()
	}
	if len(files) == 1 {
		return filepath.Base(files[0].Path)
	}
	return fmt.Sprintf("%s and other %d files", filepath.Base(files[0].Path), len(files)-1)
}

func (t *DownloadTask) ByteComplete() int64 {
	if t.Gid == nil {
		return 0
	}
	files, err := t.Gid.GetFiles()
	if err != nil {
		return 0
	}
	var totalCompleteSize int64 = 0
	for _, file := range files {
		totalCompleteSize += int64(file.CompletedLength)
	}
	return totalCompleteSize
}

func (t *DownloadTask) Length() int64 {
	if t.Gid == nil {
		return 0
	}
	files, err := t.Gid.GetFiles()
	if err != nil {
		return 0
	}
	var totalSize int64 = 0
	for _, file := range files {
		totalSize += int64(file.Length)
	}
	return totalSize
}

func (t *DownloadTask) Start() error {
	if t.Gid == nil {
		return errors.New("task not found")
	}
	t.Gid.Unpause()
	t.Status = Downloading
	t.SaveTask.Status = Downloading
	return nil
}

func (t *DownloadTask) Stop() error {
	if t.Gid == nil {
		return errors.New("task not found")
	}
	t.Status = Stop
	t.SaveTask.Status = Stop
	err := t.Gid.Pause()
	if err != nil {
		return err
	}
	//t.OnStop <- struct{}{}
	return nil
}

func (t *DownloadTask) Delete() error {
	if t.Gid == nil {
		return errors.New("task not found")
	}
	err := t.Gid.ForceRemove()
	if err != nil {
		return err
	}
	err = database.Instance.Unscoped().Model(&database.FileTask{}).Where("id = ?", t.TaskId).Error
	if err != nil {
		return err
	}
	return nil
}

func (t *DownloadTask) GetSpeed() int64 {
	return 0
}

func (t *DownloadTask) TaskStatus() TaskStatus {
	return t.Status
}

func (t *DownloadTask) SavedTaskId() int {
	return t.SaveTask.ID
}
func (t *DownloadTask) GetInfo() interface{} {
	return nil
}
func NewDownloadTask(link string, savePath string) *DownloadTask {
	return &DownloadTask{
		TaskId:     xid.New().String(),
		SavePath:   savePath,
		Url:        link,
		Status:     Downloading,
		OnPrepare:  make(chan struct{}),
		OnComplete: make(chan struct{}),
		OnStop:     make(chan struct{}),
		CreateTime: time.Now(),
	}
}
func (t *DownloadTask) Download() {
	go func() {
		t.Gid.WaitForDownload()
		t.OnComplete <- struct{}{}
	}()
}
func (t *DownloadTask) Run(e *Engine) {
	gid, err := e.Aria2Client.AddURI([]string{t.Url}, &arigo.Options{Dir: t.SavePath})
	if err != nil {
		fmt.Println(err)
		return
	}
	t.Gid = &gid
	Logger.WithField("id", t.TaskId).WithField("url", t.Url).Info("Downloading")
	t.Download()
	// update taskinfo
	var count int64 = -1
	database.Instance.Model(&database.FileTask{}).Where("gid = ?", t.Gid.GID).Count(&count)
	if count == 0 {
		database.Instance.Create(&database.FileTask{Gid: t.Gid.GID, UserUid: e.Config.Uid, Id: t.TaskId})
	}
	t.OnPrepare <- struct{}{}
	// update with request result
	//run for done chan
	go func() {
		select {
		case <-t.OnComplete:
			t.Status = Complete
			t.SaveTask.Status = Complete
			//t.SaveTask.Save(e.Database)
			Logger.WithField("id", t.TaskId).Info("task complete")
			return
		case <-t.OnStop:
			t.Status = Stop
			//t.SaveTask.Save(e.Database)
			Logger.WithField("id", t.TaskId).Info("task interrupt")
			return
		}
	}()
}

func (p *TaskPool) newFileTaskFromSaveTask(savedTask *SaveFileDownloadTask) *DownloadTask {
	return &DownloadTask{
		TaskId:     savedTask.TaskId,
		SavePath:   savedTask.SavePath,
		Url:        savedTask.Url,
		Status:     savedTask.Status,
		SaveTask:   savedTask,
		CreateTime: savedTask.CreateTime,
	}
}
