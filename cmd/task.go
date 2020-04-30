package main

import (
	// "fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/billyplus/luatable"
	"github.com/billyplus/luatable/encode"
	"github.com/pkg/errors"
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

	if task.Config.Format == "ts" {
		return task.genTS()
	} else {
		return task.genConfig()
	}

}

func (task *Task) error(err error) error {
	return errors.Wrapf(err, "表%s", task.Sheet.Name)
}

func (task *Task) genTS() error {
	// 创建目录
	if err := os.MkdirAll(task.Config.Path, 0644); err != nil {
		return task.error(err)
	}

	data, err := GenTSFile(task.Sheet, task.Config.Filter)
	if err != nil {
		return task.error(err)
	}
	outfile := filepath.Join(task.Config.Path, task.Sheet.Name+"."+task.Config.Format)
	// 写入文件
	return ioutil.WriteFile(outfile, []byte(data), 0644)
}

func (task *Task) genConfig() error {
	// 创建目录
	if err := os.MkdirAll(task.Config.Path, 0644); err != nil {
		return task.error(err)
	}

	reader := readerFatory(task.Sheet, task.Config.Filter)
	result, err := reader.ReadAll()
	if err != nil {
		return task.error(err)
	}
	// 写lua，用于调试
	if task.Config.GenLua {
		// 创建目录
		if err = os.MkdirAll("./lua", 0644); err == nil {
			outfile := filepath.Join("./lua", task.Sheet.Name+"."+task.Config.Filter+".lua")
			// 写入文件
			ioutil.WriteFile(outfile, []byte(result), 0644)
		}
	}
	var value interface{}
	err = luatable.Unmarshal(result, &value)
	if err != nil {
		return task.error(err)
	}
	outfile := filepath.Join(task.Config.Path, task.Sheet.Name+"."+task.Config.Format)

	var enc encode.EncodeFunc
	switch task.Config.Format {
	case "xml":
		enc = encode.EncodeXML
	case "json":
		enc = encode.EncodeJSON
	case "lua":
		{
			if err = ioutil.WriteFile(outfile, []byte(result), 0644); err != nil {
				return task.error(err)
			}
			return nil
		}
	default:
		return errors.Errorf("不支持的格式类型{out.format}=%s", task.Config.Format)
	}
	// 编码成json或xml
	data, err := enc(value)
	if err != nil {
		return task.error(err)
	}
	// 写入文件
	if err = ioutil.WriteFile(outfile, data, 0644); err != nil {
		return task.error(err)
	}
	return nil
}
