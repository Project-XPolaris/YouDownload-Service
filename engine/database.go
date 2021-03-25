package engine

import (
	"github.com/anacrolix/torrent/metainfo"
	"github.com/asdine/storm/v3"
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
func OpenDatabase() (*Database, error) {
	db, err := storm.Open("./data.db")
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
	ID       int `storm:"id,increment"`
	Name     string
	TaskId   string
	Status   TaskStatus
	MetaInfo *metainfo.MetaInfo
}

func (s *SavedTorrentTask) Save(database *Database) error {
	err := database.DB.Save(s)
	if err != nil {
		return err
	}
	return nil
}

func (s *SavedTorrentTask) RemoveTask(database *Database) error {
	var savedTask SavedTorrentTask
	err := database.DB.One("TaskId", s.TaskId, &savedTask)
	if err != nil {
		return err
	}
	err = database.DB.Drop(&savedTask)
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

func NewSavedTask(taskId string, metaInfo metainfo.MetaInfo, status TaskStatus, name string) *SavedTorrentTask {
	return &SavedTorrentTask{
		TaskId:   taskId,
		MetaInfo: &metaInfo,
		Status:   status,
		Name:     name,
	}
}

type SaveFileDownloadTask struct {
	ID       int `storm:"id,increment"`
	Name     string
	TaskId   string
	Url      string
	SavePath string
	Status   TaskStatus
}

func (s *SaveFileDownloadTask) Save(database *Database) error {
	err := database.DB.Save(s)
	if err != nil {
		return err
	}
	return nil
}

func (s *SaveFileDownloadTask) RemoveTask(database *Database) error {
	var savedTask SavedTorrentTask
	err := database.DB.One("TaskId", s.TaskId, &savedTask)
	if err != nil {
		return err
	}
	err = database.DB.Drop(&savedTask)
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
