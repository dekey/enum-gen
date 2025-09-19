package filesystem

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

var (
	ErrFailToGetCallerID = errors.New("failed to get caller info")
	ErrFailToFindRootDir = errors.New("failed to find root dir")
)

func FindRootDir(file string, skipCaller int) (string, error) {
	_, currentFilepath, _, ok := runtime.Caller(skipCaller)
	if !ok {
		return "", fmt.Errorf("%w", ErrFailToGetCallerID)
	}

	dir := findRootDir(currentFilepath, file)
	if dir == "" {
		return "", fmt.Errorf(
			"cannot find root dir for file [%s] in filepath [%s] %w",
			file,
			currentFilepath,
			ErrFailToFindRootDir,
		)
	}

	return dir, nil
}

func findRootDir(from string, file string) string {
	dir := filepath.Dir(from)
	gopath := filepath.Clean(os.Getenv("GOPATH"))
	for dir != "/" && dir != gopath {
		envFile := filepath.Join(dir, file)
		if _, err := os.Stat(envFile); os.IsNotExist(err) {
			dir = filepath.Dir(dir)
			continue
		}
		return dir
	}
	return ""
}
