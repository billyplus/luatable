package main

import (
	"fmt"
	excel "github.com/360EntSecGroup-Skylar/excelize"
	"github.com/billyplus/luatable/worker"
	"github.com/billyplus/luatable/xlsx"
	"os"
	"path/filepath"
	"strconv"
)

func genConfig(conf Config) {
	dispatcher := worker.NewDispater(20)
	jobQueue := make(chan worker.Job, 5)
	errChan := make(chan error)
	sheetChan := make(chan *WorkSheet, 3)

	go dispatcher.Run(jobQueue, errChan)

	go func() {
		for {
			err := <-errChan
			fmt.Println(err.Error())
		}
	}()

	go filepath.Walk(conf.DataPath, func(path string, fileinfo os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if fileinfo.IsDir {
			return nil
		}
		filename := fileinfo.Name
		go iterateXlsx(filename, sheetChan, errChan)
	})

	iterateWorksheet(sheetChan, conf.ExportConfig)
	return
}

func iterateWorksheet(sheetChan chan *WorkSheet, configs []ExportConfig) {
	for {
		sheet := <-sheetChan
		for _, conf := range configs {
			task := NewTask(sheet, conf)
			jobQueue <- task.Run
		}
	}
}

func iterateXlsx(file string, sheetChan chan *WorkSheet, errChan chan error) {
	xlsfile, err := excel.OpenFile(file)
	if err != nil {
		errChan <- err
		return
	}
	result := make(map[string]*WorkSheet)
	// sname := xlsfile.GetSheetName(1)
	for _, sh := range xlsfile.GetSheetMap() {
		data := xlsfile.GetRows(sh)
		if checkValidSheet(data) {

			sheet := &WorkSheet{
				Type:       data[0][1],
				ServerPath: data[0][3],
				ClientPath: data[1][3],
				Data:       data[3:],
				Name:       sh,
			}
			if sheet.Type == "base" {
				count, err := strconv.ParseUint(data[1][1], 10, 64)
				if err != nil {
					errChan <- err
					continue
				}
				sheet.KeyCount = int(count)
			}
			sheetChan <- sheet
		}
	}
	return
}

func checkValidSheet(data [][]string) bool {
	typ := data[0][1]
	if len(data) < 9 {
		return false
	}
	if typ == "base" || typ == "tiny" {
		return true
	}
	return false
}

type WorkSheet struct {
	Type       string
	Data       [][]string
	ServerPath string
	ClientPath string
	Name       string
	KeyCount   int
}

func readerFatory(sheet *WorkSheet, filter string) xlsx.Reader {
	switch sheet.Type {
	case "base":
		return xlsx.NewBaseReader(sheet.Name, sheet.Data, filter, sheet.KeyCount, 1, 3, 2, 4)
	case "tiny":
		return xlsx.NewTinyReader(sheet.Name, sheet.Data, filter, 1, 2, 3, 4)
	}
}
