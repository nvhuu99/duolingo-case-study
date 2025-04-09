package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// JsonReader implement ConfigReader interface, it reads configuration values
// from JSON files located in a specific directory.
//
// Usage: retrieve a string value from the configuration file.
//    `database.json
//        {
//            "migration": {
//                "source": "local",
//                "uri": "migration"
//            }
//        }
//   `
//    // keys are seperated by '.', the first key is the config file name
//    uri := reader.Get("database.migration.uri", "")
type JsonReader struct {
	dir   string         // Directory containing JSON configuration files.
	cache map[string]any // Cache to store parsed JSON data by file name.
}

// NewJsonReader creates a new JsonReader for the specified directory.
func NewJsonReader(dir string) *JsonReader {
	reader := JsonReader{}
	reader.dir = dir
	reader.cache = make(map[string]any)
	return &reader
}

// Get retrieves a string value from the configuration at the given path.
// If the path does not exist or the value is not a string, it returns the provided default value.
func (reader *JsonReader) Get(path string, val string) string {
	if readResult, err := reader.read(path); err == nil {
		if converted, ok := readResult.(string); ok {
			return converted
		}
	}
	return val
}

// GetInt retrieves an integer value from the configuration at the given path.
// If the path does not exist or the value is not an integer, it returns the provided default value.
func (reader *JsonReader) GetInt(path string, val int) int {
	if readResult, err := reader.read(path); err == nil {
		if converted, ok := readResult.(float64); ok {
			return int(converted)
		}
	}
	return val
}

// GetArr retrieves an array of strings from the configuration at the given path.
// If the path does not exist or the value is not an array of strings, it returns the provided default array.
func (reader *JsonReader) GetArr(path string, val []string) []string {
	if readResult, err := reader.read(path); err == nil {
		if arr, ok := readResult.([]any); ok {
			converted := make([]string, len(arr))
			for i, v := range arr {
				if converted[i], ok = v.(string); !ok {
					return val
				}
			}
			return converted
		}
	}
	return val
}

// GetIntArr retrieves an array of integers from the configuration at the given path.
// If the path does not exist or the value is not an array of integers, it returns the provided default array.
func (reader *JsonReader) GetIntArr(path string, val []int) []int {
	if readResult, err := reader.read(path); err == nil {
		if arr, ok := readResult.([]any); ok {
			converted := make([]int, len(arr))
			for i, v := range arr {
				if _, ok := v.(float64); !ok { // JSON numbers are parsed as float64 by default.
					return val
				}
				converted[i] = int(v.(float64))
			}
			return converted
		}
	}
	return val
}

// read retrieves a configuration value from the JSON file, given a dot-separated path (e.g., "file.section.key").
// It traverses the JSON object hierarchy to find the specified key.
func (reader *JsonReader) read(path string) (any, error) {
	parts := strings.Split(path, ".")
	if len(parts) <= 1 {
		return nil, fmt.Errorf("config: invalid path")
	}
	name := parts[0]

	// Parse the JSON file corresponding to the top-level name.
	iterator, err := reader.parseJson(name)
	if err != nil {
		return nil, err
	}

	// Traverse the JSON object using the remaining path components.
	for i := 1; i < len(parts)-1; i++ {
		if _, exists := iterator[parts[i]]; !exists {
			return nil, fmt.Errorf("config: \"%v\" not exists in %v", path, name+".json")
		}
		next, ok := iterator[parts[i]].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("config: \"%v\" not exists in %v", path, name+".json")
		}
		iterator = next
	}

	// Retrieve the final key in the path.
	key := parts[len(parts)-1]
	if _, exists := iterator[key]; !exists {
		return nil, fmt.Errorf("config: \"%v\" not exists in %v", path, name+".json")
	}

	return iterator[key], nil
}

// parseJson parses a JSON file and caches the result.
// If the file has already been parsed, it retrieves the cached data.
func (reader *JsonReader) parseJson(name string) (map[string]any, error) {
	// Check if the file has already been cached.
	if value, exists := reader.cache[name]; exists {
		return value.(map[string]any), nil
	}

	// Build the file path and check if the file exists.
	uri := filepath.Join(reader.dir, name+".json")
	if _, err := os.Stat(uri); err != nil {
		return nil, err
	}

	// Read the JSON file.
	p, err := os.ReadFile(uri)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON data into a map.
	var data map[string]any
	if err = json.Unmarshal(p, &data); err != nil {
		return nil, err
	}

	// Cache the parsed JSON data for future use.
	reader.cache[name] = data
	return data, nil
}
