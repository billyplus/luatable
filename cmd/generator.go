package main

import (
	"fmt"
	"github.com/pkg/errors"
	excel "github.com/tealeg/xlsx"
	// excel "github.com/360EntSecGroup-Skylar/excelize"
	"github.com/billyplus/luatable/worker"
	"github.com/billyplus/luatable/xlsx"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type generator struct {
	wg         *sync.WaitGroup
	dispatcher *worker.Dispatcher
	jobQueue   chan worker.Job
	errChan    chan error
	sheetChan  chan *WorkSheet
	quit       chan bool
}

func newGenerator(maxworker int) *generator {
	gen := &generator{}
	gen.dispatcher = worker.NewDispater(maxworker)
	gen.jobQueue = make(chan worker.Job, 10)
	gen.errChan = make(chan error)
	gen.sheetChan = make(chan *WorkSheet, 10)
	gen.wg = &sync.WaitGroup{}
	gen.quit = make(chan bool)

	go gen.dispatcher.Run(gen.jobQueue, gen.errChan, gen.wg)
	go gen.handleError()
	return gen
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func (gen *generator) handleError() {
	for {
		select {
		case err := <-gen.errChan:
			fmt.Println(err.Error())
			if errs, ok := err.(stackTracer); ok {
				st := errs.StackTrace()
				fmt.Printf("%+v\n", st) // top two frames
				// for _, f := range err.StackTrace() {
				// 		fmt.Printf("%+s:%d", f)
				// }
			}
		case <-gen.quit:
			break
		}
	}

}

func (gen *generator) stop() {

}

func (gen *generator) GenConfig(conf Config) {
	go gen.iterateWorksheet(conf.Out)

	func(sheetChan chan *WorkSheet, errChan chan error) {
		filepath.Walk(conf.DataPath, func(path string, fileinfo os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if fileinfo.IsDir() {
				return nil
			}
			filename := fileinfo.Name()
			if filepath.Ext(filename) == ".xlsx" {
				if strings.Contains(filename, "~$") {
					return nil
				}
				filename = filepath.Join(conf.DataPath, fileinfo.Name())
				gen.iterateXlsx(filename)
			}
			return nil
		})

	}(gen.sheetChan, gen.errChan)

	fmt.Println("wait")
	gen.wg.Wait()
	time.Sleep(500 * time.Microsecond)
	return
}

func (gen *generator) iterateWorksheet(configs []ExportConfig) {
	for {
		sheet := <-gen.sheetChan
		for _, conf := range configs {
			task := NewTask(sheet, conf)
			gen.wg.Add(1)
			gen.jobQueue <- task.Run
			fmt.Printf("处理{ file: %s.%s, target: %s}\n", sheet.Name, conf.Format, conf.Filter)
		}
	}
}

func (gen *generator) iterateXlsx(file string) {
	xlsfile, err := excel.OpenFile(file)
	if err != nil {
		gen.errChan <- err
		return
	}
	sheets, err := xlsfile.ToSlice()
	if err != nil {
		gen.errChan <- err
		return
	}
	// for _, sh := range xlsfile.GetSheetMap() {
	for i := range sheets {
		// data := xlsfile.GetRows(sh)
		sh := xlsfile.Sheets[i].Name
		data := sheets[i]
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
					gen.errChan <- err
					continue
				}
				sheet.KeyCount = int(count)
			}
			gen.sheetChan <- sheet
		}
	}
	return
}

func checkValidSheet(data [][]string) bool {
	typ := data[0][1]
	switch typ {
	case "base":
		if len(data) < 9 {
			return false
		}
	case "tiny":
		if len(data) < 6 {
			return false
		}
	default:
		return false
	}
	return true
}

// WorkSheet 对应一张有效的数据表
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
	default:
		return xlsx.NewTinyReader(sheet.Name, sheet.Data, filter, 1, 2, 3, 4)
	}
}
