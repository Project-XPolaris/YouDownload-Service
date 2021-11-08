package storage

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v3"
)

var DefaultTaskRepository = TaskRepository{}

type Task struct {
	Uid    string
	TaskId string
}
type TaskRepository struct {
	db *badger.DB
}

func (r *TaskRepository) init(db *badger.DB) error {
	r.db = db
	return nil
}
func (r *TaskRepository) save(task *Task) error {
	if r.db == nil {
		return nil
	}
	keyName := []byte(fmt.Sprintf("task_%s_%s", task.Uid, task.TaskId))
	err := r.db.Update(func(txn *badger.Txn) error {
		rawSaveData, err := json.Marshal(task)
		if err != nil {
			return err
		}
		return txn.Set(keyName, rawSaveData)
	})
	if err != nil {
		return err
	}
	return nil
}
func (r *TaskRepository) GetTaskWithUid(uid string) ([]*Task, error) {
	result := make([]*Task, 0)
	err := r.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(fmt.Sprintf("task_%s", uid))
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				task := &Task{}
				err := json.Unmarshal(v, item)
				if err != nil {
					return err
				}
				result = append(result, task)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *TaskRepository) GetTaskWithId(uid string, id string) (*Task, error) {
	result, err := r.GetTaskWithUid(uid)
	if err != nil {
		return nil, err
	}
	for _, task := range result {
		if task.TaskId == id {
			return task, nil
		}
	}
	return nil, nil
}
