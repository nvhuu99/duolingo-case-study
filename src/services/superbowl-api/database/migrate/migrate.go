package main

import (
	"duolingo/common"
	"duolingo/services/superbowl-api/bootstrap"
	"flag"
	"log"
	"os/exec"
)

func main() {
	bootstrap.Run()

	config := common.Config
	file := common.Dir("..","..", "command", "migrate", "migrate.go")
	
	flag.Parse()
	
	cmd := exec.Command(
		"go", "run", file, 
		"--db", config.Get("database.driver", ""),
		"--db-name", config.Get("database.name", ""),
		"--host", config.Get("database.host", ""),
		"--port", config.Get("database.port", ""),
		"--user", config.Get("database.user", ""),
		"--pwd", config.Get("database.password", ""),
		"--src", config.Get("database.migration.source", ""),
		"--src-uri", config.Get("database.migration.uri", ""),
		flag.Arg(0),
	)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	log.Println(string(output))
}
