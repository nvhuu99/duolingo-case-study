package main

import (
	"duolingo/lib/config-reader"
	"flag"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
)

func main() {
	_, caller, _, _ := runtime.Caller(0)
	dir := filepath.Dir(caller)
	file := filepath.Join(dir, "..", "..", "..", "command", "migrate", "migrate.go")
	config := config.NewJsonReader(filepath.Join(dir, "..", "..", "config"))
	
	flag.Parse()
	
	cmd := exec.Command(
		"go", "run", file, 
		"--db", config.Get("db.campaign.driver", ""),
		"--db-name", config.Get("db.campaign.name", ""),
		"--host", config.Get("db.campaign.host", ""),
		"--port", config.Get("db.campaign.port", ""),
		"--user", config.Get("db.campaign.user", ""),
		"--pwd", config.Get("db.campaign.password", ""),
		"--src", config.Get("db.campaign.migration.source", ""),
		"--src-uri", config.Get("db.campaign.migration.uri", ""),
		flag.Arg(0),
	)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	log.Println(string(output))
}
