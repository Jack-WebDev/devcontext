package codingtool_test

import (
	"path/filepath"
	"testing"

	codingtool "devctx/packages/core/codingtool"
)

type statusConsumer struct {
	fileName string
}

func (c statusConsumer) StatusDataFileName() string {
	return c.fileName
}

func TestStatusDataPathUsesToolOwnedStorage(t *testing.T) {
	path, err := codingtool.StatusDataPath(codingtool.ContextPaths{
		RootDir:    "/home/alex/.devctx/contexts/company",
		StorageDir: "/home/alex/.devctx/contexts/company/tools/future-tool",
	}, statusConsumer{fileName: "status.json"})
	if err != nil {
		t.Fatalf("status data path: %v", err)
	}
	if want := filepath.Join("/home/alex/.devctx/contexts/company/tools/future-tool", "status.json"); path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}
}

func TestStatusDataPathRejectsPathsOutsideToolStorage(t *testing.T) {
	_, err := codingtool.StatusDataPath(codingtool.ContextPaths{StorageDir: "/tool"}, statusConsumer{fileName: "../status.json"})
	if err == nil {
		t.Fatal("status data path error = nil, want invalid file name")
	}
}

func TestVSCodeEditorProvidesStatusDataFileName(t *testing.T) {
	if got, want := (codingtool.VSCodeEditor{}).StatusDataFileName(), "devctx-status.json"; got != want {
		t.Fatalf("status data file name = %q, want %q", got, want)
	}
}
