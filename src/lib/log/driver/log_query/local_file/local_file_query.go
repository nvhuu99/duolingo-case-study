package local_file

import (
	"bufio"
	"duolingo/lib/log"
	lq "duolingo/lib/log/log_query"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type LocalFileQuery struct {
	path        string
	dateFrom    time.Time
	dateTo      time.Time
	level       string
	filePrefix  string
	filter      map[string]any
	parsedFiles []string
}

func FileQuery(path string, dateFrom time.Time, dateTo time.Time) *LocalFileQuery {
	reader := &LocalFileQuery{
		path:        path,
		dateFrom:    dateFrom,
		dateTo:      dateTo,
		filter:      make(map[string]any),
		parsedFiles: []string{},
	}
	return reader
}

func (reader *LocalFileQuery) Info() lq.LogQuery {
	reader.level = string(log.LogLevelAsString[log.LevelInfo])
	return reader
}

func (reader *LocalFileQuery) Error() lq.LogQuery {
	reader.level = string(log.LogLevelAsString[log.LevelError])
	return reader
}

func (reader *LocalFileQuery) Debug() lq.LogQuery {
	reader.level = string(log.LogLevelAsString[log.LevelDebug])
	return reader
}

func (reader *LocalFileQuery) Filters(conditions map[string]any) lq.LogQuery {
	reader.filter = conditions
	return reader
}

func (reader *LocalFileQuery) First(filter func(*log.Log) bool) (*log.Log, error) {
	var first *log.Log
	err := reader.loop(func(log *log.Log) lq.LoopAction {
		if filter == nil || filter(log) {
			first = log
			return lq.LoopCancel
		}
		return lq.LoopContinue
	})
	return first, err
}

func (reader *LocalFileQuery) Sum(extractor func(*log.Log) float64) (float64, error) {
	var accumulation float64
	err := reader.loop(func(log *log.Log) lq.LoopAction {
		accumulation += extractor(log)
		return lq.LoopContinue
	})
	return accumulation, err
}

func (reader *LocalFileQuery) Avg(extractor func(*log.Log) float64) (float64, error) {
	var accumulation float64
	var count int64

	err := reader.loop(func(log *log.Log) lq.LoopAction {
		accumulation += extractor(log)
		count++
		return lq.LoopContinue
	})
	if count == 0 {
		return 0, nil
	}
	average := float64(accumulation) / float64(count)

	return average, err
}

func (reader *LocalFileQuery) Count(filter func(*log.Log) bool) (int, error) {
	count := 0
	err := reader.loop(func(log *log.Log) lq.LoopAction {
		if filter == nil || filter(log) {
			count++
		}
		return lq.LoopContinue
	})
	return count, err
}

func (reader *LocalFileQuery) All() ([]*log.Log, error) {
	logs := []*log.Log{}
	err := reader.loop(func(log *log.Log) lq.LoopAction {
		logs = append(logs, log)
		return lq.LoopContinue
	})
	return logs, err
}

func (reader *LocalFileQuery) Each(callback func(*log.Log) lq.LoopAction) error {
	err := reader.loop(func(log *log.Log) lq.LoopAction {
		return callback(log)
	})
	return err
}

func (reader *LocalFileQuery) Any(filter func(*log.Log) bool) (bool, error) {
	var calledOnce bool
	err := reader.loop(func(log *log.Log) lq.LoopAction {
		if filter == nil || filter(log) {
			calledOnce = true
			return lq.LoopCancel
		}
		return lq.LoopContinue
	})
	return calledOnce, err
}

func (reader *LocalFileQuery) ParsedFiles() []string {
	return reader.parsedFiles
}

func (reader *LocalFileQuery) listFiles() ([]string, error) {
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
func (reader *LocalFileQuery) loop(callback func(*log.Log) lq.LoopAction) error {
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

func (reader *LocalFileQuery) readLines(path string, callback func(*log.Log) lq.LoopAction) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		bytes := scanner.Bytes()
		var asMap map[string]any
		if err = json.Unmarshal(bytes, &asMap); err != nil {
			return err
		}
		if !matchFilters(asMap, reader.filter) {
			continue
		}
		asLog := new(log.Log)
		if err := json.Unmarshal(bytes, asLog); err != nil {
			return err
		}
		if loopAction := callback(asLog); loopAction == lq.LoopCancel {
			return nil
		}
	}

	return nil
}

func matchFilters(log map[string]any, conditions map[string]any) bool {
	cdStack := []map[string]any{}
	lgStack := []map[string]any{}

	cdIterator := conditions
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
