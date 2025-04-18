package config

// ConfigReader defines an interface for retrieving configuration values.
//
// Usage: retrieve a string value from the configuration file.
//
//	 `database.yml
//	     nested_config:
//	         first_key: value
//	         sec_key: value
//	`
//	 // keys are seperated by '.', the first key is the config file name
//	 val := reader.Get("database.nested_config.first_key", "")
type ConfigReader interface {
	Get(path string, val string) string
	GetArr(path string, val []string) []string
	GetInt(path string, val int) int
	GetIntArr(path string, val []int) []int
}
