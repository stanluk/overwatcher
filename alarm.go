package main

import (
	"fmt"
)

type Alarmer interface {
	AlarmNow(message string) error
}

var alarmer Alarmer

func InstallAlarmer(alr Alarmer) {
	alarmer = alr
}

func AlarmNow(message string) error {
	if alarmer == nil {
		return fmt.Errorf("No alarmer interface installed")
	}
	return alarmer.AlarmNow(message)
}
