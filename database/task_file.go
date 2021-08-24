package database

import "gorm.io/gorm"

type FileTask struct {
	gorm.Model
	Id      string
	Gid     string
	UserUid string
}
