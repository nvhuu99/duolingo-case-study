package config_reader

import (
	"fmt"

	"github.com/tidwall/gjson"
)

type JsonConfigReader struct {
	source Source
	cache  map[string][][]byte
}

func NewJsonConfigReader() *JsonConfigReader {
	return &JsonConfigReader{
		cache: make(map[string][][]byte),
	}
}

func (reader *JsonConfigReader) LoadFromLocalFiles(configDir string) *JsonConfigReader {
	reader.source = NewLocalFile(configDir).AcceptExtentions(".json")
	return reader
}

func (reader *JsonConfigReader) Get(uri string, pattern string) string {
	return reader.get(uri, pattern).String()
}

func (reader *JsonConfigReader) GetInt(uri string, pattern string) int {
	return int(reader.get(uri, pattern).Int())
}

func (reader *JsonConfigReader) GetInt64(uri string, pattern string) int64 {
	return reader.get(uri, pattern).Int()
}

func (reader *JsonConfigReader) GetArr(uri string, pattern string) []string {
	result := []string{}
	data := reader.get(uri, pattern).Array()
	for r := range data {
		if data[r].Exists() {
			result = append(result, data[r].String())
		}
	}
	return result
}

func (reader *JsonConfigReader) GetIntArr(uri string, pattern string) []int {
	result := []int{}
	data := reader.get(uri, pattern).Array()
	for r := range data {
		if data[r].Exists() {
			result = append(result, int(data[r].Int()))
		}
	}
	return result
}

func (reader *JsonConfigReader) GetInt64Arr(source string, pattern string) []int64 {
	result := []int64{}
	data := reader.get(source, pattern).Array()
	for r := range data {
		if data[r].Exists() {
			result = append(result, data[r].Int())
		}
	}
	return result
}

func (reader *JsonConfigReader) get(uri string, pattern string) gjson.Result {
	if reader.source == nil {
		panic(fmt.Errorf(ErrSourceIsNotSet, "JsonConfigReader"))
	}

	var rawContents [][]byte
	var err error

	if _, exists := reader.cache[uri]; exists {
		rawContents = reader.cache[uri]
	} else {
		rawContents, err = reader.source.Load(uri)
		if err != nil {
			panic(fmt.Errorf(ErrSourceFailure, reader.source, err))
		}
		reader.cache[uri] = rawContents
	}

	for _, content := range rawContents {
		data := gjson.ParseBytes(content).Get(pattern)
		if data.Exists() {
			return data
		}
	}

	panic(fmt.Errorf(ErrConfigNotFound, pattern, reader.source))
}
