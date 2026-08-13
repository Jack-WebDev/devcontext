package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteProjectBindingsFile writes project bindings through a same-directory
// temporary file and atomic rename.
func WriteProjectBindingsFile(path string, bindings []Binding) error {
	data, err := EncodeProjectBindingsTOML(bindings)
	if err != nil {
		return err
	}

	return writeProjectBindingsFileAtomically(path, func(file *os.File) error {
		_, err := file.Write(data)
		return err
	})
}

type projectBindingsAtomicWriteFunc func(file *os.File) error

func writeProjectBindingsFileAtomically(path string, write projectBindingsAtomicWriteFunc) error {
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
	syncProjectBindingsDirectory(dir)
	return nil
}

func syncProjectBindingsDirectory(path string) {
	dir, err := os.Open(path)
	if err != nil {
		return
	}
	defer dir.Close()

	_ = dir.Sync()
}
