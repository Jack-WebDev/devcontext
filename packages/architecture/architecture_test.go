package architecture_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

var frameworkFreeRoots = []string{
	filepath.Join("..", "application"),
	filepath.Join("..", "core"),
}

func TestCoreAndApplicationPackagesDoNotImportWails(t *testing.T) {
	for _, root := range frameworkFreeRoots {
		root := root

		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}

			file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}

			for _, imported := range file.Imports {
				importPath, err := strconv.Unquote(imported.Path.Value)
				if err != nil {
					return err
				}
				if importPath == "github.com/wailsapp/wails/v2" || strings.HasPrefix(importPath, "github.com/wailsapp/wails/v2/") {
					t.Fatalf("%s imports Wails through %q", path, importPath)
				}
			}

			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}
}
