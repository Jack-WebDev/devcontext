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
	return WriteGlobalConfigFileWithPermissions(path, globalConfig, filesystem.NewDefaultStoragePermissions())
}

// WriteGlobalConfigFileWithPermissions writes global configuration using the
// supplied storage permission policy.
func WriteGlobalConfigFileWithPermissions(path string, globalConfig GlobalConfig, permissions filesystem.StoragePermissions) error {
	if permissions == nil {
		permissions = filesystem.NewDefaultStoragePermissions()
	}

	data, err := EncodeGlobalConfigTOML(globalConfig)
	if err != nil {
		return err
	}

	if err := writeFileAtomically(path, func(file *os.File) error {
		_, err := file.Write(data)
		return err
	}); err != nil {
		return err
	}

	return permissions.ApplyFile(path)
}

// ReadGlobalConfigFile reads and decodes global configuration from path.
func ReadGlobalConfigFile(path string) (GlobalConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if wrapped := filesystem.WrapStoragePermissionError("read file", path, err); wrapped != err {
			return GlobalConfig{}, fmt.Errorf("read global configuration %q: %w", path, wrapped)
		}
		return GlobalConfig{}, fmt.Errorf("read global configuration %q: %w", path, err)
	}
	return DecodeGlobalConfigFileTOML(path, data)
}

// InitializeDevContextHome creates the top-level storage layout and default
// global configuration when it does not already exist.
func InitializeDevContextHome(paths filesystem.PlatformPaths) (HomeLayout, error) {
	return InitializeDevContextHomeWithPermissions(paths, filesystem.NewDefaultStoragePermissions())
}

// InitializeDevContextHomeWithPermissions creates the top-level storage layout
// with the supplied storage permission policy.
func InitializeDevContextHomeWithPermissions(paths filesystem.PlatformPaths, permissions filesystem.StoragePermissions) (HomeLayout, error) {
	if permissions == nil {
		permissions = filesystem.NewDefaultStoragePermissions()
	}

	layout, err := DevContextHomeLayout(paths)
	if err != nil {
		return HomeLayout{}, err
	}

	for _, dir := range []string{layout.HomeDir, layout.ContextsDir, layout.LogsDir} {
		if err := os.MkdirAll(dir, permissions.DirectoryMode()); err != nil {
			if wrapped := filesystem.WrapStoragePermissionError("create directory", dir, err); wrapped != err {
				return HomeLayout{}, fmt.Errorf("create Dev Context directory %q: %w", dir, wrapped)
			}
			return HomeLayout{}, fmt.Errorf("create Dev Context directory %q: %w", dir, err)
		}
		if err := permissions.ApplyDirectory(dir); err != nil {
			return HomeLayout{}, err
		}
	}

	if err := ensureDefaultGlobalConfig(layout.ConfigPath, permissions); err != nil {
		return HomeLayout{}, err
	}

	return layout, nil
}

func ensureDefaultGlobalConfig(path string, permissions filesystem.StoragePermissions) error {
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return fmt.Errorf("global configuration path %q is a directory", path)
		}
		return permissions.ApplyFile(path)
	}
	if !os.IsNotExist(err) {
		if wrapped := filesystem.WrapStoragePermissionError("inspect file", path, err); wrapped != err {
			return fmt.Errorf("inspect global configuration %q: %w", path, wrapped)
		}
		return fmt.Errorf("inspect global configuration %q: %w", path, err)
	}

	if err := WriteGlobalConfigFileWithPermissions(path, DefaultGlobalConfig(), permissions); err != nil {
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
		if wrapped := filesystem.WrapStoragePermissionError("create temporary file", dir, err); wrapped != err {
			return fmt.Errorf("create temporary file for %q: %w", path, wrapped)
		}
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
		if wrapped := filesystem.WrapStoragePermissionError("write temporary file", tempPath, err); wrapped != err {
			return fmt.Errorf("write temporary file for %q: %w", path, wrapped)
		}
		return fmt.Errorf("write temporary file for %q: %w", path, err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		if wrapped := filesystem.WrapStoragePermissionError("sync temporary file", tempPath, err); wrapped != err {
			return fmt.Errorf("sync temporary file for %q: %w", path, wrapped)
		}
		return fmt.Errorf("sync temporary file for %q: %w", path, err)
	}
	if err := file.Close(); err != nil {
		if wrapped := filesystem.WrapStoragePermissionError("close temporary file", tempPath, err); wrapped != err {
			return fmt.Errorf("close temporary file for %q: %w", path, wrapped)
		}
		return fmt.Errorf("close temporary file for %q: %w", path, err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		if wrapped := filesystem.WrapStoragePermissionError("replace file", path, err); wrapped != err {
			return fmt.Errorf("replace %q atomically: %w", path, wrapped)
		}
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
