package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"syscall"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) < 3 {
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
		err = copyTemplate(dirPath)
		if err != nil {
			log.Fatal("error while copying template files: ", err)
		}
		os.Chdir(dirPath)
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
		os.Chdir(cwd)
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
		os.Chdir(dirPath)
		bin, binErr := exec.LookPath("go")
		if binErr != nil {
			log.Fatal(binErr)
		}
		args := []string{"go", "run", appPath}
		execErr := syscall.Exec(bin, args, os.Environ())
		if execErr != nil {
			log.Fatal("error while running sketch: ", execErr)
		}
		os.Chdir(cwd)
	default:
		usage()
	}
}

func usage() {
	fmt.Println("Usage: sketchy command prefix")
	fmt.Println("Commands:")
	fmt.Println("\tinit - create new project with name 'prefix'")
	fmt.Println("\trun - run project with name 'prefix'")
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
