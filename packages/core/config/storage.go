package config

import (
	"fmt"
	"os"
	"path/filepath"

	"devctx/packages/core/filesystem"
)

const (
	globalConfigFileName = "config.toml"
	contextsDirName      = "contexts"
	logsDirName          = "logs"
	directoryMode        = 0o755
)

// HomeLayout describes the top-level Dev Context storage paths.
type HomeLayout struct {
	HomeDir     string
	ConfigPath  string
	ContextsDir string
	LogsDir     string
}

// DevContextHomeLayout derives the top-level Dev Context storage layout.
func DevContextHomeLayout(paths filesystem.PlatformPaths) (HomeLayout, error) {
	homeDir, err := paths.DevContextHomeDir()
	if err != nil {
		return HomeLayout{}, err
	}

	return HomeLayout{
		HomeDir:     homeDir,
		ConfigPath:  filepath.Join(homeDir, globalConfigFileName),
		ContextsDir: filepath.Join(homeDir, contextsDirName),
		LogsDir:     filepath.Join(homeDir, logsDirName),
	}, nil
}

// WriteGlobalConfigFile writes global configuration through a same-directory
// temporary file and atomic rename.
func WriteGlobalConfigFile(path string, globalConfig GlobalConfig) error {
	data, err := EncodeGlobalConfigTOML(globalConfig)
	if err != nil {
		return err
	}

	return writeFileAtomically(path, func(file *os.File) error {
		_, err := file.Write(data)
		return err
	})
}

// InitializeDevContextHome creates the top-level storage layout and default
// global configuration when it does not already exist.
func InitializeDevContextHome(paths filesystem.PlatformPaths) (HomeLayout, error) {
	layout, err := DevContextHomeLayout(paths)
	if err != nil {
		return HomeLayout{}, err
	}

	for _, dir := range []string{layout.HomeDir, layout.ContextsDir, layout.LogsDir} {
		if err := os.MkdirAll(dir, directoryMode); err != nil {
			return HomeLayout{}, fmt.Errorf("create Dev Context directory %q: %w", dir, err)
		}
	}

	if err := ensureDefaultGlobalConfig(layout.ConfigPath); err != nil {
		return HomeLayout{}, err
	}

	return layout, nil
}

func ensureDefaultGlobalConfig(path string) error {
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return fmt.Errorf("global configuration path %q is a directory", path)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("inspect global configuration %q: %w", path, err)
	}

	if err := WriteGlobalConfigFile(path, DefaultGlobalConfig()); err != nil {
		return fmt.Errorf("write default global configuration %q: %w", path, err)
	}
	return nil
}

type atomicWriteFunc func(file *os.File) error

func writeFileAtomically(path string, write atomicWriteFunc) error {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	file, err := os.CreateTemp(dir, "."+base+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary file for %q: %w", path, err)
	}

	tempPath := file.Name()
	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.Remove(tempPath)
		}
	}()

	if err := write(file); err != nil {
		_ = file.Close()
		return fmt.Errorf("write temporary file for %q: %w", path, err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync temporary file for %q: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close temporary file for %q: %w", path, err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace %q atomically: %w", path, err)
	}

	removeTemp = false
	syncDirectory(dir)
	return nil
}

func syncDirectory(path string) {
	dir, err := os.Open(path)
	if err != nil {
		return
	}
	defer dir.Close()

	_ = dir.Sync()
}
