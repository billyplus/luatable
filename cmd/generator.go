package main

import (
	"fmt"
	excel360 "github.com/360EntSecGroup-Skylar/excelize"
	"github.com/billyplus/luatable/worker"
	"github.com/billyplus/luatable/xlsx"
	"github.com/pkg/errors"
	excel "github.com/tealeg/xlsx"
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
	config     *Config
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
			// 空文件，不处理
			if errors.Cause(err) == xlsx.ErrNoContent {
				continue
			}
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
	gen.config = &conf
	go gen.iterateWorksheet(conf.Out)

	func(sheetChan chan *WorkSheet, errChan chan error) {
		filepath.Walk(conf.DataPath, func(path string, fileinfo os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if fileinfo.IsDir() {
				if conf.SkipSubDir && path != "." {
					return filepath.SkipDir
				}
				return nil
			}

			// filename := fileinfo.Name()
			if filepath.Ext(path) == ".xlsx" {
				if strings.Contains(path, "~$") {
					return nil
				}
				filename := filepath.Join(conf.DataPath, path)
				fmt.Println("打开文件:", filename)
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
			fmt.Printf("生成 { file: %s.%s, target: %s}\n", sheet.Name, conf.Format, conf.Filter)
		}
	}
}

func (gen *generator) iterateXlsx(file string) {
	defer func() { // 用defer来捕获到panic异常
		if r := recover(); r != nil {
			var err error
			if e, ok := r.(error); ok {
				err = errors.WithStack(e)
			} else {
				err = errors.Errorf("错误：%v", e)
			}
			gen.errChan <- err
		}
	}()
	if gen.config.Use360 {
		gen.sheetsFromExcel360(file)
	} else {
		gen.sheetsFromXlsx(file)
	}
}

func (gen *generator) sheetsFromXlsx(file string) {
	xlsfile, err := excel.OpenFile(file)
	if err != nil {
		gen.errChan <- err
		return
	}

	for _, sheet := range xlsfile.Sheets {
		if checkValidSheet(sheet) {
			sh := sheet.Name
			data, err := xlsx.SheetToSlice(sheet)
			if err != nil {
				gen.errChan <- err
				continue
			}
			worksheet := &WorkSheet{
				Type: data[0][1],
				Data: data[3:],
				Name: sh,
				// ServerPath: data[0][3],
				// ClientPath: data[1][3],
			}
			if worksheet.Type == "base" {
				count, err := strconv.ParseUint(data[1][1], 10, 64)
				if err != nil {
					gen.errChan <- err
					continue
				}
				worksheet.KeyCount = int(count)
			}
			gen.sheetChan <- worksheet
		}
	}
}

func (gen *generator) sheetsFromExcel360(file string) {
	xlsfile, err := excel360.OpenFile(file)
	if err != nil {
		gen.errChan <- err
		return
	}

	// for index := range xlsfile.SheetCount {
	for i := 0; i < xlsfile.SheetCount; i++ {
		sh := xlsfile.GetSheetName(i)

		data := xlsfile.GetRows(sh)
		// data := xlsfile.GetRows(sh)
		// sheet := xlsfile.
		if checkValid360Sheet(data) {
			worksheet := &WorkSheet{
				Type: data[0][1],
				Data: data[3:],
				Name: sh,
				// ServerPath: data[0][3],
				// ClientPath: data[1][3],
			}
			if worksheet.Type == "base" {
				count, err := strconv.ParseUint(data[1][1], 10, 64)
				if err != nil {
					gen.errChan <- err
					continue
				}
				worksheet.KeyCount = int(count)
			}
			gen.sheetChan <- worksheet
		}
	}
}

func checkValid360Sheet(data [][]string) bool {
	rows := len(data)
	if rows < 6 || len(data[0]) < 2 {
		return false
	}

	typ := data[0][1]
	switch typ {
	case "base":
		if rows < 9 || len(data[1]) < 2 {
			return false
		}
	case "tiny":
		if rows < 6 {
			return false
		}
	default:
		return false
	}
	return true
}

func checkValidSheet(sheet *excel.Sheet) bool {
	rows := sheet.MaxRow
	if rows < 6 || len(sheet.Rows[0].Cells) < 2 {
		return false
	}

	typ := sheet.Rows[0].Cells[1].Value
	switch typ {
	case "base":
		if rows < 9 || len(sheet.Rows[1].Cells) < 2 {
			return false
		}
	case "tiny":
		if rows < 6 {
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
		return xlsx.NewBaseReader(sheet.Name, sheet.Data, filter, sheet.KeyCount, 1, 0, 3, 4)
	default:
		return xlsx.NewTinyReader(sheet.Name, sheet.Data, filter, 1, 2, 3, 4)
	}
}
