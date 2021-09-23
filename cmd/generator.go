package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	excel360 "github.com/360EntSecGroup-Skylar/excelize"
	"github.com/billyplus/luatable"
	"github.com/billyplus/luatable/encode"
	"github.com/billyplus/luatable/xlsx"
	"github.com/pkg/errors"
	excel "github.com/tealeg/xlsx"
)

type generator struct {
	wg     *sync.WaitGroup
	config *Config
	quit   chan bool
	errors []error
}

func newGenerator() *generator {
	gen := &generator{}
	gen.wg = &sync.WaitGroup{}
	gen.quit = make(chan bool)

	return gen
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func (gen *generator) Close() {
	if gen.quit != nil {
		close(gen.quit)
		gen.quit = nil
	}
}

func (gen *generator) PrintErrors() {
	fmt.Printf("一共有 %d 个错误", len(gen.errors))
	fmt.Println("-------------------------------------------")
	fmt.Println("")
	for i, err := range gen.errors {
		fmt.Printf("****** %d ******\n", i)
		fmt.Println(err.Error())
		fmt.Println("")
		// if errs, ok := err.(stackTracer); ok {
		// 	st := errs.StackTrace()
		// 	fmt.Printf("%v\n", st) // top two frames
		// 	// for _, f := range err.StackTrace() {
		// 	// 		fmt.Printf("%+s:%d", f)
		// 	// }
		// }
	}
	fmt.Println("")
	fmt.Println("-------------------------------------------")

}

func (gen *generator) GenConfig(conf Config) {
	gen.config = &conf
	fmt.Println("handle path:", conf.DataPath)

	md5list := make(map[string]string)
	data, err := ioutil.ReadFile(conf.MD5File)
	if err != nil {
		gen.errors = append(gen.errors, err)
		return
	}
	if err := json.Unmarshal(data, &md5list); err != nil {
		gen.errors = append(gen.errors, err)
		return
	}
	filepath.Walk(conf.DataPath, func(path string, fileinfo os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if fileinfo.IsDir() {
			if conf.SkipSubDir && path != conf.DataPath {
				return filepath.SkipDir
			}
			return nil
		}

		filename := fileinfo.Name()
		fpath := ""
		if filepath.Ext(filename) == ".xlsx" {
			if strings.HasPrefix(filename, "~$") || strings.HasPrefix(filename, ".~") {
				return nil
			}
			fpath = filepath.Join(conf.DataPath, filename)
			//check hash
			checksum, err := md5Hash(fpath)
			if err != nil {
				return nil
			}
			md5str, ok := md5list[filename]
			if ok && checksum == md5str {
				fmt.Println("skip file:", filename)
				return nil
			}
			fmt.Println("handle file:", filename)
			if err = gen.iterateXlsx(fpath); err != nil {
				// fmt.Printf("failed to handle xlsx err: %+v", err)
				gen.errors = append(gen.errors, err)
				// return nil
			} else {
				md5list[filename] = checksum
				if err = saveMd5Hash(md5list, conf.MD5File); err != nil {
					fmt.Println("failed to save md5", fpath)
					return err
				}
			}
		}
		return nil
	})

	for _, out := range gen.config.Out {
		if out.MergeJson {
			var mergedJson bytes.Buffer
			mergedJson.WriteByte('{')
			filepath.Walk(out.Path, func(path string, fileinfo os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if fileinfo.IsDir() {
					return nil
				}

				filename := fileinfo.Name()
				if !strings.HasSuffix(filename, ".json") {
					return nil
				}
				mergedJson.WriteString("\n    \"")
				mergedJson.WriteString(filename[:len(filename)-5])
				mergedJson.WriteString("\": ")
				fpath := filepath.Join(out.Path, filename)
				data, err := ioutil.ReadFile(fpath)
				if err != nil {
					return err
				}
				if data[len(data)-1] == '\n' {
					data = data[:len(data)-1]
				}
				mergedJson.Write(data)
				mergedJson.WriteByte(',')
				return nil
			})
			data := mergedJson.Bytes()
			n := len(data)
			if data[n-1] == ',' {
				data = data[:n-1]
			}
			data = append(data, '}')

			outpath := strings.Split(out.Path, "/")

			err = ioutil.WriteFile("./"+outpath[len(outpath)-1]+".json", data, 0644)
			if err != nil {
				gen.errors = append(gen.errors, err)
			}
		}
	}
	fmt.Println("wait")
	time.Sleep(500 * time.Microsecond)
	return
}

func saveMd5Hash(md5list map[string]string, path string) error {
	data, err := json.MarshalIndent(md5list, "", "    ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

func md5Hash(filepath string) (string, error) {
	//Initialize variable returnMD5String now in case an error has to be returned
	var returnMD5String string

	//Open the passed argument and check for any error
	file, err := os.Open(filepath)
	if err != nil {
		return returnMD5String, err
	}

	//Tell the program to call the following function when the current function returns
	defer file.Close()

	//Open a new hash interface to write to
	hash := md5.New()

	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}

	//Get the 16 bytes hash
	hashInBytes := hash.Sum(nil)[:16]

	//Convert the bytes to a string
	return hex.EncodeToString(hashInBytes), nil
}

func (gen *generator) iterateXlsx(file string) (err error) {
	// defer func() { // 用defer来捕获到panic异常
	// 	if r := recover(); r != nil {
	// 		if e, ok := r.(error); ok {
	// 			err = errors.WithStack(e)
	// 		} else {
	// 			err = errors.Errorf("错误：%v", e)
	// 		}
	// 	}
	// }()
	if gen.config.Use360 {
		err = gen.sheetsFromExcel360(file)
	} else {
		err = gen.sheetsFromXlsx(file)
	}
	return
}

func (gen *generator) sheetsFromXlsx(file string) error {
	xlsfile, err := excel.OpenFile(file)
	if err != nil {
		return err
	}
	lst := strings.Split(file, "/")
	filename := lst[len(lst)-1]

	for _, sheet := range xlsfile.Sheets {
		if checkValidSheet(sheet) {
			sh := sheet.Name
			data, err := xlsx.SheetToSlice(sheet)
			if err != nil {
				return err
			}
			worksheet := &WorkSheet{
				Type:     data[0][1],
				Data:     data[2:],
				Name:     sh,
				FileName: filename,
			}
			head := data[0][5]
			head = strings.ReplaceAll(head, " ", "")
			if strings.HasSuffix(head, "={") {
				head = head[:len(head)-2]
			}
			worksheet.Head = head
			if worksheet.Type == "base" {
				count, err := strconv.ParseUint(data[0][3], 10, 64)
				if err != nil {
					return err
				}
				worksheet.KeyCount = int(count)
			}
			for _, expConf := range gen.config.Out {
				// outfile := data[expConf.OutCellX][expConf.OutCellY]
				if err = genConfig(worksheet, expConf); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (gen *generator) sheetsFromExcel360(file string) error {
	xlsfile, err := excel360.OpenFile(file)
	if err != nil {
		return err
	}

	// for index := range xlsfile.SheetCount {
	for i := 0; i < xlsfile.SheetCount; i++ {
		sh := xlsfile.GetSheetName(i)

		data := xlsfile.GetRows(sh)
		// data := xlsfile.GetRows(sh)
		// sheet := xlsfile.
		if checkValid360Sheet(data) {
			data[7][0] = "comment"
			worksheet := &WorkSheet{
				Type: data[0][1],
				Data: data[2:],
				Name: sh,
			}
			head := data[0][5]
			head = strings.ReplaceAll(head, " ", "")
			if strings.HasSuffix(head, "={") {
				head = head[:len(head)-2]
			}
			worksheet.Head = head
			if worksheet.Type == "base" {
				count, err := strconv.ParseUint(data[0][3], 10, 64)
				if err != nil {
					return err
				}
				worksheet.KeyCount = int(count)
			}
			for _, expConf := range gen.config.Out {
				// outfile := data[expConf.OutCellX][expConf.OutCellY]
				if err = genConfig(worksheet, expConf); err != nil {
					return err
				}
			}
		}
	}

	return nil
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
	if rows < 4 || len(sheet.Rows[0].Cells) < 6 {
		return false
	}

	typ := sheet.Rows[0].Cells[1].Value
	switch typ {
	case "base":
		if rows < 7 || len(sheet.Rows[0].Cells) < 6 {
			return false
		}
	case "tiny":
		if rows < 4 {
			return false
		}
	default:
		return false
	}
	return true
}

// WorkSheet 对应一张有效的数据表
type WorkSheet struct {
	Type     string
	Data     [][]string
	Name     string
	Head     string
	FileName string
	KeyCount int
}

func readerFatory(sheet *WorkSheet, filter string) xlsx.Reader {
	switch sheet.Type {
	case "base":
		return xlsx.NewBaseReader(sheet.Name, sheet.Data, filter, sheet.KeyCount, 1, 2, 3, 4)
	default:
		return xlsx.NewTinyReader(sheet.Name, sheet.Data, filter, 1, 2, 3, 4)
	}
}

func genConfig(sheet *WorkSheet, cnf ExportConfig) (err error) {
	fmt.Println("gen config for sheet: ", sheet.Head)
	outfile := filepath.Join(cnf.Path, sheet.Head+"."+cnf.Format)
	// 创建目录
	if err = os.MkdirAll(filepath.Dir(outfile), 0755); err != nil {
		return
	}
	fmt.Println("success mkdirall ")

	reader := readerFatory(sheet, cnf.Filter)
	var result []byte
	fmt.Println("start readall ")
	result, err = reader.ReadAll()
	if err != nil {
		if err == xlsx.ErrNoContent {
			err = nil
		}
		return
	}

	if cnf.Format == "lua" {
		f, err := os.OpenFile(outfile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		tag := fmt.Sprintf("-- from %s %s \n%s=", sheet.FileName, sheet.Name, sheet.Head)

		if _, err = f.WriteString(tag); err != nil {
			return err
		}
		if _, err = f.Write(result); err != nil {
			return err
		}

		cmd := exec.Command("lua", outfile)
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(string(out))
			return errors.New(string(out))
			// return errors.Wrap(err, "check lua")
		}

		// f.Seek(0, 0)
		// if _, err = luaparse.Parse(f, sheet.Head); err != nil {
		// 	return errors.Wrap(err, "parse lua")
		// }
		return nil
	}
	// 其它格式

	var value interface{}
	err = luatable.Unmarshal(result, &value)
	if err != nil {
		return
	}
	fmt.Println("path", cnf.Path, " ", outfile)
	var enc encode.EncodeFunc
	switch cnf.Format {
	case "xml":
		enc = encode.EncodeXML
	case "json":
		enc = encode.EncodeJSON
	default:
		return errors.Errorf("不支持的格式类型{out.format}=%s", cnf.Format)
	}
	// 编码成json或xml
	var data []byte
	data, err = enc(value)
	if err != nil {
		return
	}
	// 写入文件
	if err = ioutil.WriteFile(outfile, data, 0644); err != nil {
		return
	}
	return nil
}

func genTS(sheet *WorkSheet, cnf ExportConfig) (err error) {
	// defer func() {
	// 	if err != nil {
	// 		err = errors.Wrapf(err, "表: %s ", sheet.Name)
	// 	}
	// }()
	// 创建目录
	if err = os.MkdirAll(cnf.Path, 0644); err != nil {
		return
	}

	var data []byte
	data, err = GenTSFile(sheet, cnf.Filter)
	if err != nil {
		return
	}
	outfile := filepath.Join(cnf.Path, sheet.Name+"."+cnf.Format)
	// 写入文件
	err = ioutil.WriteFile(outfile, []byte(data), 0644)
	return
}
