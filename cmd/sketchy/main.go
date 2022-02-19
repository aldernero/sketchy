package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	targetDir := initCmd.String("d", cwd, "Target directory")
	prefix := initCmd.String("p", "sketch", "Project prefix")
	if len(os.Args) < 2 {
		fmt.Println("expected 'init' or 'run' subcommands")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "init":
		initCmd.Parse(os.Args[2:])
		dirPath := path.Join(*targetDir, *prefix)
		if _, err := os.Stat(dirPath); errors.Is(err, fs.ErrExist) {
			log.Fatal("can't make directory: ", err)
		}
		err := os.Mkdir(dirPath, 0644)
		if err != nil {
			log.Fatal("error while creating directory: ", err)
		}
	}
}
