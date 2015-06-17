package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"
)

//@todo repeats in model/model.go
var GlassUserDir = ".timeglass"

type Logger struct {
	file *os.File
	hash string

	*log.Logger
}

func NewLogger(dir string, w io.Writer) (*Logger, error) {
	l := &Logger{
		hash: fmt.Sprintf("%x", md5.Sum([]byte(dir))),
	}

	p, err := l.Path()
	if err != nil {
		return nil, err
	}

	l.file, err = os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	mw := io.MultiWriter(l.file, w)

	l.Logger = log.New(mw, "", log.Ldate|log.Lmicroseconds)
	return l, nil
}

func (l *Logger) Path() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", errwrap.Wrapf("Failed to detect current user for writing logs: {{err}}", err)
	}

	return filepath.Join(u.HomeDir, GlassUserDir, fmt.Sprintf("%s.log", l.hash)), nil
}

func (l *Logger) Close() error {
	return l.Close()
}
