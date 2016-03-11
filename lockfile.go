package main

import (
	"bufio"
	"io"
	"os"
	"time"
)

type Lockfile struct {
	Path string
}

func (v *Lockfile) TryWriteTime(when time.Time) error {
	file, err := os.OpenFile(v.Path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(when.UTC().Format(time.RFC822Z))
	return err
}

func (v *Lockfile) LoadTime() (time.Time, error) {
	var ret time.Time
	file, err := os.Open(v.Path)
	if err != nil {
		return ret, err
	}
	defer file.Close()

	buf := bufio.NewReader(file)
	str, err := buf.ReadString('\n')
	if err != nil && err != io.EOF {
		return ret, err
	}
	ret, err = time.Parse(time.RFC822Z, str)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

func (v *Lockfile) TryWriteDuration(when time.Duration) error {
	file, err := os.OpenFile(v.Path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(when.String())
	return err
}

func (v *Lockfile) LoadDuration() (time.Duration, error) {
	var ret time.Duration
	file, err := os.Open(v.Path)
	if err != nil {
		return ret, err
	}
	defer file.Close()

	buf := bufio.NewReader(file)
	str, err := buf.ReadString('\n')
	if err != nil && err != io.EOF {
		return ret, err
	}
	ret, err = time.ParseDuration(str)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

func (v *Lockfile) Remove() {
	os.Remove(v.Path)
}
