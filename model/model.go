package model

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"
	"github.com/hashicorp/errwrap"
)

type Model struct {
	dbpath string
}

func New(dir string) *Model {
	return &Model{filepath.Join(dir, "timer.db")}
}

func (m *Model) Open() (*bolt.DB, error) {
	opts := &bolt.Options{Timeout: time.Millisecond * 100}

	db, err := bolt.Open(m.dbpath, 0600, opts)
	if err != nil {
		return nil, errwrap.Wrapf(fmt.Sprintf("Failed to open database '%s': {{err}}", m.dbpath), err)
	}

	return db, nil
}

func (m *Model) Close(db *bolt.DB) {
	//@todo handle error?
	db.Close()
}

func (m *Model) UpsertDaemonInfo(info *Daemon) error {
	db, err := m.Open()
	if err != nil {
		return err
	}

	defer m.Close(db)
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(MetaBucketName))
		if err != nil {
			return errwrap.Wrapf(fmt.Sprintf("Failed to create meta bucket: {{err}}"), err)
		}

		data, err := info.Serialize()
		if err != nil {
			return errwrap.Wrapf(fmt.Sprintf("Failed to serialize db deamon '%s': {{err}}", info.Addr), err)
		}

		b.Put([]byte(DeamonKeyName), data)
		return nil
	})
}

func (m *Model) ReadDaemonInfo() (*Daemon, error) {
	var info *Daemon
	db, err := m.Open()
	if err != nil {
		return nil, err
	}

	defer m.Close(db)
	return info, db.View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte(MetaBucketName))
		if b == nil {
			return fmt.Errorf("Failed to open daemon bucket")
		}

		data := b.Get([]byte(DeamonKeyName))
		if data != nil {
			info, err = NewDaemonFromSerialized(data)
			if err != nil {
				return errwrap.Wrapf(fmt.Sprintf("Failed to deserialize daemon from db: {{err}}, data: '%s'", string(data)), err)
			}
		}

		return nil
	})
}
