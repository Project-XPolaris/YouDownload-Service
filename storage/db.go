package storage

import (
	"github.com/dgraph-io/badger/v3"
	"log"
)

func InitDB() error {
	db, err := badger.Open(badger.DefaultOptions("dataStorage"))
	if err != nil {
		log.Fatal(err)
	}
	err = DefaultUserRepository.init(db)
	if err != nil {
		return err
	}
	return nil
}
