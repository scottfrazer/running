package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func CreateDirectoryIfNotExists(dirName string) error {
	src, err := os.Stat(dirName)

	if os.IsNotExist(err) {
		err := os.MkdirAll(dirName, 0755)
		if err != nil {
			return err
		}
		return nil
	}

	if src.Mode().IsRegular() {
		return fmt.Errorf("Directory %s is a regular file", dirName)
	}

	return nil
}

// open opens the specified URL in the default browser of the user.
func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
