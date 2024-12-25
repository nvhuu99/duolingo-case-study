package main

import (
	"flag"
	"log"
	"os/exec"
	"duolingo/services/superbowl-api/common"
)

func main() {
	flag.Parse()
	file := common.Dir("..","..", "command", "migrate", "migrate.go")
	cmd := exec.Command(
		"go", "run", file, 
		"--db", common.Config().Get("database.driver", ""),
		"--db-name", common.Config().Get("database.name", ""),
		"--host", common.Config().Get("database.host", ""),
		"--port", common.Config().Get("database.port", ""),
		"--user", common.Config().Get("database.user", ""),
		"--pwd", common.Config().Get("database.password", ""),
		"--src", common.Config().Get("database.migration.source", ""),
		"--src-uri", common.Config().Get("database.migration.uri", ""),
		flag.Arg(0),
	)  
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	log.Println(string(output))
}
