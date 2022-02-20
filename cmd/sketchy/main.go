package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
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
		err := os.Mkdir(dirPath, 0700)
		if err != nil {
			log.Fatal("error while creating directory: ", err)
		}
		err = copyTemplate(dirPath)
		if err != nil {
			log.Fatal("error while copying template files: ", err)
		}
	}
}

func copyTemplate(targetDir string) error {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("error while getting working directory: ", err)
	}
	configFname := path.Join(cwd, "template", "sketch.json")
	appFname := path.Join(cwd, "template", "main.go")
	exists, err := regularFileExists(configFname)
	if err != nil {
		return fmt.Errorf("error while checking for config file: %v", err)
	}
	if !exists {
		return fmt.Errorf("config file %s doesn't exist", configFname)
	}
	exists, err = regularFileExists(appFname)
	if err != nil {
		return fmt.Errorf("error while checking for config file: %v", err)
	}
	if !exists {
		return fmt.Errorf("app file %s doesn't exist", appFname)
	}
	err = copyFile(configFname, path.Join(targetDir, "sketch.json"))
	if err != nil {
		return fmt.Errorf("error while copying config file: %v", err)
	}
	err = copyFile(appFname, path.Join(targetDir, "main.go"))
	if err != nil {
		return fmt.Errorf("error while copying app file: %v", err)
	}
	return nil
}

func regularFileExists(fname string) (bool, error) {
	stat, err := os.Stat(fname)
	if err == nil {
		return stat.Mode().IsRegular(), nil
	} else {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
	}
	return false, err
}

func copyFile(src string, dst string) error {
	srcFd, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFd.Close()
	dstFd, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFd.Close()
	_, err = io.Copy(dstFd, srcFd)
	if err != nil {
		return err
	}
	err = dstFd.Sync()
	return err
}
