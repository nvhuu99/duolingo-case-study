package local_file

import (
	"bufio"
	wt "duolingo/lib/log/writer"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"
	"time"
)

type LocalFileOutput struct {
	Dir string
}

func NewLocalFileOutPut(dir string) *LocalFileOutput {
	return &LocalFileOutput{ dir }
}

func (out *LocalFileOutput) Flush(items []*wt.Writable) error {
	groupByPath := make(map[string][]byte)
	for _, item := range items {
		rotation, err := time.Parse("20060102150405", item.Rotation)
		if err != nil {
			return err
		}
		year := fmt.Sprintf("%d", rotation.Year())
		month := fmt.Sprintf("%02d", int(rotation.Month()))
		day := fmt.Sprintf("%02d", rotation.Day())

		parts := strings.Split(item.URI, "/")
		pathParts := parts[0:len(parts) - 1]
		pathParts = slices.Concat(pathParts, []string{ year, month, day})
		nameParts := strings.Split(parts[len(parts) - 1], ".")
		name := strings.Join(nameParts[0:len(nameParts)-1], ".")
		ext := nameParts[len(nameParts) - 1]
		
		filename := name + "_" + item.Rotation + "." + ext
		filePath := strings.Join(pathParts, "/")
		location := path.Join(out.Dir, filePath, filename)

		if _, exists := groupByPath[location]; exists {
			groupByPath[location] = slices.Concat(groupByPath[location], item.Content, []byte("\n")) 
		} else {
			groupByPath[location] = slices.Concat(item.Content, []byte("\n"))
		}
	}

	for location, bytes := range groupByPath {
		if err := out.writeToFile(location, bytes); err != nil {
			return err
		}
	}

	return nil
}


func (src *LocalFileOutput) writeToFile(location string, content []byte) error {
	os.MkdirAll(path.Dir(location), 0755)
	file, err := os.OpenFile(location, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	wr := bufio.NewWriter(file)
	defer wr.Flush()

	_, err = wr.Write(content)

	return err
}