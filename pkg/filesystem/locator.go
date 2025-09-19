package filesystem

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	GoModName = "go.mod"
)

var (
	ErrFailToGetCallerID  = errors.New("failed to get caller info")
	ErrFailToFindRootDir  = errors.New("failed to find root dir")
	ErrModulePathNotFound = errors.New("module path not found in go.mod")
)

type Locator struct{}

func NewLocator() *Locator {
	return &Locator{}
}

func (l *Locator) FindRootDirWithGoMod(skipCaller int) (string, error) {
	result, err := l.FindRootDir(GoModName, skipCaller)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (l *Locator) FindRootDir(file string, skipCaller int) (string, error) {
	_, currentFilepath, _, ok := runtime.Caller(skipCaller)
	if !ok {
		return "", fmt.Errorf("%w", ErrFailToGetCallerID)
	}

	dir := l.findRootDir(currentFilepath, file)
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

func (l *Locator) findRootDir(from string, file string) string {
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

func (*Locator) ReadModulePath(root string) (string, error) {
	goModFilePath := filepath.Join(root, GoModName)

	fileContentBytes, err := os.ReadFile(goModFilePath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(fileContentBytes), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			mod := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			// strip quotes if any
			mod = strings.Trim(mod, "\"`")
			// drop trailing .git if present
			mod = strings.TrimSuffix(mod, ".git")

			return mod, nil
		}
	}
	return "", fmt.Errorf("%w", ErrModulePathNotFound)
}

// RelativePackagePath returns the package path relative to the module root.
// modRoot is a path from root dir to this project like: `/Users/username/project`
// fullPath is a full path to package  like `/Users/username/project/pkg/destination`
// returns relative path to package project/pkg/destination
func (*Locator) RelativePackagePath(modRoot string, fullPath string) (string, error) {
	slog.Debug("RelativePackagePath", slog.String("modRoot", modRoot), slog.String("fullPath", fullPath))
	result, err := filepath.Rel(modRoot, fullPath)
	if err != nil {
		return "", err
	}
	slog.Debug(
		"RelativePackagePath",
		slog.String("result", result),
	)

	return filepath.Dir(result), nil
}
