package youplus

import (
	"context"
	"github.com/project-xpolaris/youplustoolkit/youplus/rpc"
	"github.com/projectxpolaris/youdownload-server/config"
)

var DefaultRPCClient *rpc.YouPlusRPCClient

func LoadYouPlusRPCClient() error {
	DefaultRPCClient = rpc.NewYouPlusRPCClient(config.Instance.YouPlusRPCAddr)
	DefaultRPCClient.KeepAlive = true
	DefaultRPCClient.MaxRetry = 1000
	return DefaultRPCClient.Connect(context.Background())
}
