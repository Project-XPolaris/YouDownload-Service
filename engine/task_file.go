package engine

import (
	"context"
	"fmt"
	"github.com/cavaliercoder/grab"
	"github.com/projectxpolaris/youdownload-server/config"
	"github.com/rs/xid"
	"path/filepath"
	"time"
)

type DownloadTask struct {
	TaskId     string
	Request    *grab.Request
	Response   *grab.Response
	Url        string
	SavePath   string
	Cancel     context.CancelFunc
	Status     TaskStatus
	SaveTask   *SaveFileDownloadTask
	OnPrepare  chan struct{}
	OnComplete chan struct{}
	CreateTime time.Time
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
	} else if t.SaveTask != nil {
		return t.SaveTask.BytesComplete
	}
	return 0
}

func (t *DownloadTask) Length() int64 {
	if t.Response != nil {
		return t.Response.Size
	} else if t.SaveTask != nil {
		return t.SaveTask.Length
	}
	return 0
}

func (t *DownloadTask) Start() error {
	t.Status = Downloading
	t.SaveTask.Status = Downloading
	return nil
}

func (t *DownloadTask) Stop() error {
	t.Cancel()
	t.Status = Stop
	t.SaveTask.Status = Stop
	return nil
}

func (t *DownloadTask) Delete() error {
	return nil
}

func (t *DownloadTask) GetSpeed() int64 {
	if t.Response != nil {
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
		TaskId:     xid.New().String(),
		SavePath:   config.Instance.DownloadDir,
		Url:        link,
		Status:     Downloading,
		OnPrepare:  make(chan struct{}),
		OnComplete: make(chan struct{}),
		CreateTime: time.Now(),
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
	Logger.WithField("id", t.TaskId).WithField("url", request.URL()).Info("Downloading")
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
			t.SaveTask.Status = Complete
			t.SaveTask.Save(e.Database)
			Logger.WithField("id", t.TaskId).Info("task complete")
			return
		case <-ctx.Done():
			t.Status = Complete
			t.SaveTask.Save(e.Database)
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
