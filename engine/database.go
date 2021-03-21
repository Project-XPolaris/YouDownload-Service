package engine

import (
	"github.com/anacrolix/torrent/metainfo"
	"github.com/asdine/storm/v3"
)

type Database struct {
	DB *storm.DB
}

func (d *Database) Save(savedTask *SavedTask) error {
	err := d.DB.Save(savedTask)
	if err != nil {
		return err
	}
	return nil
}
func (d *Database) ReadSavedTask() ([]*SavedTask, error) {
	var savedTasks []*SavedTask
	err := d.DB.All(&savedTasks)
	if err != nil {
		return nil, err
	}
	return savedTasks, nil
}
func (d *Database) RemoveTask(id string) error {
	var savedTask SavedTask
	err := d.DB.One("TaskId", id, &savedTask)
	if err != nil {
		return err
	}
	err = d.DB.Drop(&savedTask)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) UpdateTaskStatus(task *SavedTask, status TaskStatus) error {
	err := d.DB.UpdateField(task, "Status", status)
	if err != nil {
		return err
	}
	return nil
}
func OpenDatabase() (*Database, error) {
	db, err := storm.Open("./data.db")
	if err != nil {
		return nil, err
	}
	return &Database{DB: db}, nil
}

type SavedTask struct {
	ID       int `storm:"id,increment"`
	TaskId   string
	Status   TaskStatus
	MetaInfo *metainfo.MetaInfo
}

func NewSavedTask(taskId string, metaInfo metainfo.MetaInfo, status TaskStatus) *SavedTask {
	return &SavedTask{
		TaskId:   taskId,
		MetaInfo: &metaInfo,
		Status:   status,
	}
}
