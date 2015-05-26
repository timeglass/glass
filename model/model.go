package model

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/timeglass/glass/_vendor/github.com/boltdb/bolt"
	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"
)

type Model struct {
	repoDir     string
	repoDirHash string
}

func New(dir string) *Model {
	hash := md5.Sum([]byte(dir))

	return &Model{
		repoDir:     dir,
		repoDirHash: fmt.Sprintf("%x", hash),
	}
}

func (m *Model) Open() (*bolt.DB, error) {
	u, err := user.Current()
	if err != nil {
		return nil, errwrap.Wrapf("Failed to detect current user for opening database: {{err}}", err)
	}

	dbdir := filepath.Join(u.HomeDir, ".timeglass")
	err = os.MkdirAll(dbdir, 0777)
	if err != nil {
		return nil, errwrap.Wrapf(fmt.Sprintf("Failed to create directory '%s' for opening database: {{err}}", dbdir), err)
	}

	dbpath := filepath.Join(dbdir, m.repoDirHash+".db")
	opts := &bolt.Options{Timeout: time.Millisecond * 100}
	db, err := bolt.Open(dbpath, 0600, opts)
	if err != nil {
		return nil, errwrap.Wrapf(fmt.Sprintf("Failed to open database '%s': {{err}}", dbpath), err)
	}

	return db, nil
}

func (m *Model) Close(db *bolt.DB) {
	err := db.Close()

	//@todo handle error?
	_ = err
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
			info = &Daemon{}
			return nil
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

func (m *Model) ReadConfig() (*Config, error) {
	conf := DefaultConfig

	p := filepath.Join(m.repoDir, "timeglass.json")
	f, err := os.Open(p)
	if err != nil {
		if os.IsNotExist(err) {
			return conf, nil
		}

		return nil, errwrap.Wrapf(fmt.Sprintf("Error opening configuration file '%s', it does exist but: {{err}}", p), err)
	}

	dec := json.NewDecoder(f)

	defer f.Close()
	err = dec.Decode(conf)
	if err != nil {
		return nil, errwrap.Wrapf(fmt.Sprintf("Error decoding '%s' as JSON, please check for syntax errors: {{err}}", p), err)
	}

	if time.Duration(conf.MBU) < time.Minute {
		return nil, fmt.Errorf("configuration 'mbu': An MBU of less then 1min is not supported, received: '%s'", conf.MBU)
	}

	return conf, nil
}
