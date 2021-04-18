package engine

import (
	"github.com/anacrolix/torrent/metainfo"
	"github.com/asdine/storm/v3"
	"path"
	"time"
)

type Database struct {
	DB *storm.DB
}

func (d *Database) ReadSavedTask() ([]*SavedTorrentTask, error) {
	var savedTasks []*SavedTorrentTask
	err := d.DB.All(&savedTasks)
	if err != nil {
		return nil, err
	}
	return savedTasks, nil
}

func (d *Database) ReadSavedFileDownloadTask() ([]*SaveFileDownloadTask, error) {
	var savedTasks []*SaveFileDownloadTask
	err := d.DB.All(&savedTasks)
	if err != nil {
		return nil, err
	}
	return savedTasks, nil
}
func OpenDatabase(databasePath string) (*Database, error) {
	db, err := storm.Open(path.Join(databasePath, "./data.db"))
	if err != nil {
		return nil, err
	}
	return &Database{DB: db}, nil
}

type SaveTask interface {
	Save(d *Database) error
	RemoveTask(d *Database) error
	UpdateTaskStatus(d *Database, status TaskStatus) error
}

type SavedTorrentTask struct {
	ID            int `storm:"id,increment"`
	Name          string
	TaskId        string
	Status        TaskStatus
	Length        int64
	BytesComplete int64
	MetaInfo      *metainfo.MetaInfo
	CreateTime    time.Time
}

func (s *SavedTorrentTask) Save(database *Database) error {
	err := database.DB.Save(s)
	if err != nil {
		return err
	}
	return nil
}

func (s *SavedTorrentTask) RemoveTask(database *Database) error {
	err := database.DB.Drop(s)
	if err != nil {
		return err
	}
	return nil
}

func (s *SavedTorrentTask) UpdateTaskStatus(database *Database, status TaskStatus) error {
	err := database.DB.UpdateField(s, "Status", status)
	if err != nil {
		return err
	}
	return nil
}

func NewSavedTask(taskId string, metaInfo metainfo.MetaInfo, status TaskStatus, name string, createTime time.Time) *SavedTorrentTask {
	return &SavedTorrentTask{
		TaskId:     taskId,
		MetaInfo:   &metaInfo,
		Status:     status,
		Name:       name,
		CreateTime: createTime,
	}
}

type SaveFileDownloadTask struct {
	ID            int `storm:"id,increment"`
	Name          string
	TaskId        string
	Url           string
	SavePath      string
	Length        int64
	BytesComplete int64
	Status        TaskStatus
	CreateTime    time.Time
}

func (s *SaveFileDownloadTask) Save(database *Database) error {
	err := database.DB.Save(s)
	if err != nil {
		return err
	}
	return nil
}

func (s *SaveFileDownloadTask) RemoveTask(database *Database) error {
	err := database.DB.Drop(s)
	if err != nil {
		return err
	}
	return nil
}

func (s *SaveFileDownloadTask) UpdateTaskStatus(database *Database, status TaskStatus) error {
	err := database.DB.UpdateField(s, "Status", status)
	if err != nil {
		return err
	}
	return nil
}
