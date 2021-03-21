package api

import "github.com/projectxpolaris/youdownload-server/engine"

var taskStatusMapping map[engine.TaskStatus]string = map[engine.TaskStatus]string{
	engine.Estimate:    "Estimate",
	engine.Downloading: "Downloading",
	engine.Stop:        "Stop",
	engine.Complete:    "Complete",
}

type BaseTaskTemplate struct {
	Id       string  `json:"id"`
	Name     string  `json:"name"`
	Complete int64   `json:"complete"`
	Length   int64   `json:"length"`
	Progress float64 `json:"progress"`
	Status   string  `json:"status"`
	Speed    int64   `json:"speed"`
	ETA      int64   `json:"eta"`
}

func (t *BaseTaskTemplate) Serializer(dataModel interface{}, context map[string]interface{}) error {
	task := dataModel.(engine.Task)
	t.Name = task.Name()
	t.Complete = task.ByteComplete()
	t.Length = task.Length()
	t.Id = task.Id()
	if t.Length > 0 {
		t.Progress = float64(t.Complete) / float64(t.Length)
	}
	t.Status = taskStatusMapping[task.TaskStatus()]
	t.Speed = task.GetSpeed()
	if t.Speed != 0 {
		t.ETA = task.Length() / t.Speed
	}
	return nil
}
