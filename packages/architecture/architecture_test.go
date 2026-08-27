package architecture_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
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

func TestProviderSpecificSwitchesStayInsideProviderImplementations(t *testing.T) {
	goRoots := []string{
		filepath.Join("..", "application"),
		filepath.Join("..", "core", "launcher"),
		filepath.Join("..", "core", "filesystem"),
	}
	for _, root := range goRoots {
		assertNoProviderSpecificSwitches(t, root, ".go")
	}

	assertNoProviderSpecificSwitches(t, filepath.Join("..", "..", "frontend", "src", "lib"), ".ts")
}

func assertNoProviderSpecificSwitches(t *testing.T, root string, extension string) {
	t.Helper()

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, extension) || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, forbidden := range []string{
			"case provider.CodexID:",
			"case provider.ClaudeID:",
			`case "codex":`,
			`case "claude":`,
		} {
			if strings.Contains(string(contents), forbidden) {
				t.Errorf("%s contains provider-specific switch case %q", path, forbidden)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
}
