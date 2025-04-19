package local

import (
	"bufio"
	"duolingo/lib/log"
	jf "duolingo/lib/log/driver/formatter/json"
	formatter "duolingo/lib/log/formatter"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type LocalReader struct {
	path        string
	dateFrom    time.Time
	dateTo      time.Time
	level       string
	filePrefix  string
	filter      map[string]any
	formatter   formatter.Formatter
	parsedFiles []string
}

func LogQuery(path string, dateFrom time.Time, dateTo time.Time) *LocalReader {
	reader := &LocalReader{
		path:        path,
		dateFrom:    dateFrom,
		dateTo:      dateTo,
		filter:      make(map[string]any),
		formatter:   new(jf.JsonFormatter),
		parsedFiles: []string{},
	}
	return reader
}

func (reader *LocalReader) ExpectJson() *LocalReader {
	reader.formatter = new(jf.JsonFormatter)
	return reader
}

func (reader *LocalReader) Info() *LocalReader {
	reader.level = string(log.LogLevelAsString[log.LevelInfo])
	return reader
}

func (reader *LocalReader) Error() *LocalReader {
	reader.level = string(log.LogLevelAsString[log.LevelError])
	return reader
}

func (reader *LocalReader) Debug() *LocalReader {
	reader.level = string(log.LogLevelAsString[log.LevelDebug])
	return reader
}

func (reader *LocalReader) FilePrefix(prefix string) *LocalReader {
	reader.filePrefix = prefix
	return reader
}

func (reader *LocalReader) Filter(conditions map[string]any) *LocalReader {
	reader.filter = conditions
	return reader
}

func (reader *LocalReader) Sum(extractFunc func(map[string]any) int64) (int64, error) {
	var accumulation int64
	err := reader.loop(func(log map[string]any) LoopAction {
		accumulation += extractFunc(log)
		return LoopContinue
	})
	return accumulation, err
}

func (reader *LocalReader) Avg(extractFunc func(map[string]any) int64) (float64, error) {
	var accumulation int64
	var count int64

	err := reader.loop(func(log map[string]any) LoopAction {
		accumulation += extractFunc(log)
		count++
		return LoopContinue
	})
	if count == 0 {
		return 0, nil
	}
	average := float64(accumulation) / float64(count)

	return average, err
}

func (reader *LocalReader) Count() (int, error) {
	count := 0
	err := reader.loop(func(log map[string]any) LoopAction {
		count++
		return LoopContinue
	})
	return count, err
}

func (reader *LocalReader) All() ([]map[string]any, error) {
	logs := []map[string]any{}
	err := reader.loop(func(log map[string]any) LoopAction {
		logs = append(logs, log)
		return LoopContinue
	})
	return logs, err
}

func (reader *LocalReader) Each(callback func(map[string]any) LoopAction) error {
	err := reader.loop(func(log map[string]any) LoopAction {
		return callback(log)
	})
	return err
}

func (reader *LocalReader) Any() (bool, error) {
	var calledOnce bool
	err := reader.loop(func(log map[string]any) LoopAction {
		calledOnce = true
		return LoopCancel
	})
	return calledOnce, err
}

func (reader *LocalReader) First() (map[string]any, error) {
	var firstLog map[string]any
	err := reader.loop(func(log map[string]any) LoopAction {
		firstLog = log
		return LoopCancel
	})
	return firstLog, err
}

func (reader *LocalReader) ParsedFiles() []string {
	return reader.parsedFiles
}

func (reader *LocalReader) listFiles() ([]string, error) {
	files := []string{}
	err := filepath.WalkDir(reader.path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		parts := strings.Split(filepath.ToSlash(path), "/")
		if len(parts) < 4 {
			return nil
		}

		fileName := parts[len(parts)-1]
		if reader.filePrefix != "" && !strings.HasPrefix(fileName, reader.filePrefix) {
			return nil
		}
		if reader.level != "" && !strings.HasSuffix(fileName, "."+reader.level) {
			return nil
		}

		year, err1 := strconv.Atoi(parts[len(parts)-4])
		month, err2 := strconv.Atoi(parts[len(parts)-3])
		day, err3 := strconv.Atoi(parts[len(parts)-2])
		if err1 != nil || err2 != nil || err3 != nil {
			return nil
		}
		fileDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		if fileDate.Before(reader.dateFrom) || fileDate.After(reader.dateTo) {
			return nil
		}

		files = append(files, path)
		return nil
	})

	return files, err
}

// ignore unmarshal error
func (reader *LocalReader) loop(callback func(map[string]any) LoopAction) error {
	files, err := reader.listFiles()
	if err != nil {
		return err
	}

	for _, path := range files {
		reader.parsedFiles = append(reader.parsedFiles, path)
		if err := reader.readLines(path, callback); err != nil {
			return err
		}
	}

	return nil
}

func (reader *LocalReader) readLines(path string, callback func(map[string]any) LoopAction) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var raw map[string]any
		if err = json.Unmarshal(scanner.Bytes(), &raw); err != nil {
			continue
		}
		if match(raw, reader.filter) {
			if loopAction := callback(raw); loopAction == LoopCancel {
				return nil
			}
		}
	}

	return nil
}

func match(log map[string]any, condition map[string]any) bool {
	cdStack := []map[string]any{}
	lgStack := []map[string]any{}

	cdIterator := condition
	lgIterator := log

	for {
		for key := range cdIterator {
			if _, hasKey := lgIterator[key]; !hasKey {
				return false
			}
			if isMap(cdIterator[key]) {
				if !isMap(lgIterator[key]) {
					return false
				}
				cdStack = append(cdStack, cdIterator[key].(map[string]any))
				lgStack = append(lgStack, lgIterator[key].(map[string]any))
			} else {
				if cdIterator[key] != lgIterator[key] &&
					fmt.Sprintf("%v", cdIterator[key]) != fmt.Sprintf("%v", lgIterator[key]) &&
					!reflect.DeepEqual(cdIterator[key], lgIterator[key]) {
					return false
				}
			}
		}

		if len(cdStack) == 0 {
			break
		}

		cdStack, cdIterator = cdStack[:len(cdStack)-1], cdStack[len(cdStack)-1]
		lgStack, lgIterator = lgStack[:len(lgStack)-1], lgStack[len(lgStack)-1]
	}

	return true
}

func isMap(v any) bool {
	return reflect.TypeOf(v).Kind() == reflect.Map
}
