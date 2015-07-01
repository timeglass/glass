package main

import (
	"io"
	"os"
	"path/filepath"

	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"
)

type Logger struct {
	file *os.File
	path string

	io.Writer
}

func NewLogger(w io.Writer) (*Logger, error) {
	l := &Logger{}
	path, err := SystemTimeglassPathCreateIfNotExist()
	if err != nil {
		return nil, errwrap.Wrapf("Failed to find Timeglass system path: {{err}}", err)
	}

	l.path = filepath.Join(path, "glass-daemon.log")
	l.file, err = os.OpenFile(l.path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	l.Writer = io.MultiWriter(l.file, w)
	return l, nil
}

func (l *Logger) Path() string {
	return l.path
}

func (l *Logger) Close() error {
	return l.file.Close()
}
