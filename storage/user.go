package storage

import (
	"encoding/json"
	"errors"
	"github.com/ahmetb/go-linq/v3"
	"github.com/dgraph-io/badger/v3"
)

var DefaultUserRepository = UserRepository{}
var (
	usersKey = []byte("users")
)

type User struct {
	Uid      string
	DataPath string
}
type UserRepository struct {
	db    *badger.DB
	Users []*User
}

func (r *UserRepository) init(db *badger.DB) error {
	r.db = db
	err := db.View(func(txn *badger.Txn) error {
		r.Users = make([]*User, 0)
		data, err := txn.Get(usersKey)
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
		if err != nil {
			return err
		}

		rawData, err := data.ValueCopy(nil)
		if rawData == nil {
			return nil
		}
		err = json.Unmarshal(rawData, &r.Users)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
func (r *UserRepository) save() error {
	if r.db == nil {
		return nil
	}
	err := r.db.Update(func(txn *badger.Txn) error {
		rawSaveData, err := json.Marshal(r.Users)
		if err != nil {
			return err
		}
		return txn.Set(usersKey, rawSaveData)
	})
	if err != nil {
		return err
	}
	return nil
}
func (r *UserRepository) GetOrCreate(uid string) (*User, error) {
	rawResult := linq.From(r.Users).FirstWith(func(i interface{}) bool {
		return i.(*User).Uid == uid
	})
	if rawResult == nil {
		user := &User{
			Uid:      uid,
			DataPath: "",
		}
		r.Users = append(r.Users, user)
		err := r.save()
		if err != nil {
			return nil, err
		}
		return user, nil
	}
	return rawResult.(*User), nil
}
func (r *UserRepository) Save(saveUser *User) error {
	linq.From(r.Users).Where(func(i interface{}) bool {
		return i.(*User).Uid != saveUser.Uid
	}).ToSlice(&r.Users)
	r.Users = append(r.Users, saveUser)
	return r.save()
}
