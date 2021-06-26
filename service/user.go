package service

import (
	"github.com/projectxpolaris/youdownload-server/database"
	"os"
)

func CheckNeedInit(username string) (bool, error) {
	var user database.User
	err := database.Instance.FirstOrCreate(&user, database.User{Uid: username}).Error
	if err != nil {
		return false, err
	}
	return len(user.DataPath) == 0, nil
}

func InitUser(uid string, dataPath string) error {
	var user database.User
	err := database.Instance.Where("uid = ?", uid).Find(&user).Error
	if err != nil {
		return err
	}
	err = os.MkdirAll(dataPath, 0777)
	if err != nil {
		return err
	}
	user.DataPath = dataPath
	err = database.Instance.Save(&user).Error
	return err
}
