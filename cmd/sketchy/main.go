package main

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"syscall"
)

const version = "v0.1.0"

//go:embed template/*
var template embed.FS

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) < 3 {
		if len(os.Args) == 2 && os.Args[1] == "version" {
			fmt.Printf("Sketchy %s\n", version)
			os.Exit(0)
		}
		fmt.Println("expected 'init' or 'run' subcommands")
		usage()
		os.Exit(1)
	}
	prefix := os.Args[2]
	dirPath := path.Join(cwd, prefix)
	switch os.Args[1] {
	case "init":
		if _, err := os.Stat(dirPath); errors.Is(err, fs.ErrExist) {
			log.Fatal("can't make directory: ", err)
		}
		err := os.Mkdir(dirPath, 0700)
		if err != nil {
			log.Fatal("error while creating directory: ", err)
		}
		err = os.Chdir(dirPath)
		if err != nil {
			log.Fatal("error while changing directory:", err)
		}
		copyFilesFromEmbedFS(&template, "template", dirPath)
		modInitCmd := exec.Command("go", "mod", "init", prefix)
		_, cmdErr := modInitCmd.Output()
		if cmdErr != nil {
			log.Fatal("error while creating go mod: ", cmdErr)
		}
		modTidyCmd := exec.Command("go", "mod", "tidy")
		_, cmdErr = modTidyCmd.Output()
		if cmdErr != nil {
			log.Fatal("error while go mod tidy: ", cmdErr)
		}
		err = os.Chdir(cwd)
		if err != nil {
			log.Fatal("error while changing directory:", err)
		}
	case "run":
		if _, err := os.Stat(dirPath); errors.Is(err, fs.ErrNotExist) {
			log.Fatalf("directory %s doesn't exist: %v", dirPath, err)
		}
		appPath := path.Join(dirPath, "main.go")
		pathExists, err := regularFileExists(appPath)
		if err != nil {
			log.Fatal("error while looking for sketch file:", err)
		}
		if !pathExists {
			log.Fatalf("sketch file %s doesn't exist", appPath)
		}
		err = os.Chdir(dirPath)
		if err != nil {
			log.Fatal("error while changing directory:", err)
		}
		bin, binErr := exec.LookPath("go")
		if binErr != nil {
			log.Fatal(binErr)
		}
		args := []string{"go", "run", appPath}
		execErr := syscall.Exec(bin, args, os.Environ())
		if execErr != nil {
			log.Fatal("error while running sketch: ", execErr)
		}
		err = os.Chdir(cwd)
		if err != nil {
			log.Fatal("error while changing directory:", err)
		}
	default:
		usage()
	}
}

func usage() {
	fmt.Println("Usage: sketchy command prefix")
	fmt.Println("Commands:")
	fmt.Println("\tinit - create new project with name 'prefix'")
	fmt.Println("\trun - run project with name 'prefix'")
	fmt.Println("\tversion  - print Sketchy version")
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

func copyFilesFromEmbedFS(fs *embed.FS, dir, dest string) {
	files, err := fs.ReadDir(dir)
	if err != nil {
		log.Fatal("error while reading template directory: ", err)
	}
	for _, file := range files {
		if !file.IsDir() {
			fname := path.Join(dir, file.Name())
			destPath := path.Join(dest, file.Name())
			destBytes, err := fs.ReadFile(fname)
			if err != nil {
				log.Fatal("error while reading template file: ", err)
			}
			err = os.WriteFile(destPath, destBytes, 0644)
			if err != nil {
				log.Fatal("error while writing template file: ", err)
			}
		}
	}
}
