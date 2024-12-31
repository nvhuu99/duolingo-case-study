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
		"--db", config.Get("db.duolingo.driver", ""),
		"--db-name", config.Get("db.duolingo.name", ""),
		"--host", config.Get("db.duolingo.host", ""),
		"--port", config.Get("db.duolingo.port", ""),
		"--user", config.Get("db.duolingo.user", ""),
		"--pwd", config.Get("db.duolingo.password", ""),
		"--src", config.Get("db.duolingo.migration.source", ""),
		"--src-uri", config.Get("db.duolingo.migration.uri", ""),
		flag.Arg(0),
	)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	log.Println(string(output))
}
