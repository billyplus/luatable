package main

import (
	// "fmt"
	"github.com/billyplus/luatable"
	"github.com/billyplus/luatable/encode"
	"github.com/pkg/errors"
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

func (task *Task) Run() (err error) {
	defer func() { // 用defer来捕获到panic异常
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = errors.WithStack(e)
			} else {
				err = errors.Errorf("错误：%v", e)
			}
			err = errors.WithMessagef(err, "表%s", task.Sheet.Name)
		}
	}()

	// 创建目录
	if err = os.MkdirAll(task.Config.Path, 0644); err != nil {
		return task.error(err)
	}

	reader := readerFatory(task.Sheet, task.Config.Filter)
	result, err := reader.ReadAll()
	if err != nil {
		return task.error(err)
	}
	// fmt.Println(result)
	var value interface{}
	err = luatable.Unmarshal(result, &value)
	if err != nil {
		return task.error(err)
	}
	var enc encode.EncodeFunc
	switch task.Config.Format {
	case "xml":
		enc = encode.EncodeXML
	case "json":
		enc = encode.EncodeJSON
	default:
		return errors.Errorf("不支持的格式类型{out.format}=%s", task.Config.Format)
	}
	// 编码成json或xml
	data, err := enc(value)
	if err != nil {
		return task.error(err)
	}
	outfile := filepath.Join(task.Config.Path, task.Sheet.Name+"."+task.Config.Format)
	// 写入文件
	return ioutil.WriteFile(outfile, data, 0644)
}

func (task *Task) error(err error) error {
	return errors.Wrapf(err, "表%s", task.Sheet.Name)
}
