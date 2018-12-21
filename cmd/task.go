package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type Task struct {
	Sheet  *WorkSheet
	Config ExportConfig
}

func NewTask(sheet *WorkSheet, conf ExportConfig) *Task {
	task := &Task{
		Sheet:  sheet,
		Config: conf,
	}
	return task
}

func (task *Task) Run() error {
	reader := readerFatory(task.Sheet, task.Config.Filter)
	result, err := reader.ReadAll()
	if err != nil {
		return err
	}
	return writeStringToFile(result, filepath.Join(task.Config.Path, sheet.Name+"."+task.Config.Format))
}

func writeStringToFile(value, path string) error {
	return ioutil.WriteFile(path, []byte(value), 0644)

}
