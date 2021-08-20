package youplus

import (
	entry "github.com/project-xpolaris/youplustoolkit/youplus/entity"
	"github.com/projectxpolaris/youdownload-server/config"
)

var DefaultEntry *entry.EntityClient

type AppExport struct {
	Addrs []string `json:"addrs"`
}

func InitEntity() {
	DefaultEntry = entry.NewEntityClient(config.Instance.Entity.Name, config.Instance.Entity.Version, &entry.EntityExport{}, DefaultRPCClient)
	DefaultEntry.HeartbeatRate = 3000
}
