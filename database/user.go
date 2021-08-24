package database

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Uid       string
	DataPath  string
	FileTasks []FileTask `gorm:"foreignKey:UserUid;references:Uid"`
}
