package model

import (
	"fmt"
	"net/url"
	"time"

	"github.com/boltdb/bolt"
	"github.com/hashicorp/errwrap"
)

type Model struct {
	dbpath string
}

func New(dbpath string) *Model {
	return &Model{dbpath}
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

func (m *Model) ReadDaemonAddr() (*url.URL, error) {
	db, err := m.Open()
	if err != nil {
		return nil, err
	}

	defer m.Close(db)

	//@todo open db, read value and close
	return nil, nil
}
