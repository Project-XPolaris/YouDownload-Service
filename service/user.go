package service

import (
	"github.com/projectxpolaris/youdownload-server/storage"
	"os"
)

func CheckNeedInit(username string) (bool, error) {
	user, err := storage.DefaultUserRepository.GetOrCreate(username)
	if err != nil {
		return false, err
	}
	return len(user.DataPath) == 0, nil
}

func InitUser(uid string, dataPath string) error {
	user, err := storage.DefaultUserRepository.GetOrCreate(uid)
	if err != nil {
		return err
	}
	err = os.MkdirAll(dataPath, 0777)
	if err != nil {
		return err
	}
	user.DataPath = dataPath
	err = storage.DefaultUserRepository.Save(user)
	return err
}
