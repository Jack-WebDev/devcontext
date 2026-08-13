package context_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/editor"
	"devctx/packages/core/provider"
)

func TestRepositoryWriteAndGetContext(t *testing.T) {
	contextsDir := t.TempDir()
	repository := devcontext.NewRepository(contextsDir)
	ctx := storedContext("client-a", "Client A")
	createContextDir(t, contextsDir, ctx.ID)

	if err := repository.Write(ctx); err != nil {
		t.Fatalf("write context: %v", err)
	}

	got, err := repository.Get(ctx.ID)
	if err != nil {
		t.Fatalf("get context: %v", err)
	}
	if !reflect.DeepEqual(got, ctx) {
		t.Fatalf("stored context = %#v, want %#v", got, ctx)
	}
}

func TestRepositoryWriteReportsPermissionDeniedPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose Unix directory permissions consistently")
	}

	contextsDir := filepath.Join(t.TempDir(), "contexts")
	repository := devcontext.NewRepository(contextsDir)
	contextID := devcontext.MustID("personal")
	contextDir := filepath.Join(contextsDir, contextID.String())
	if err := os.MkdirAll(contextDir, 0o700); err != nil {
		t.Fatalf("create context dir: %v", err)
	}
	if err := os.Chmod(contextDir, 0o000); err != nil {
		t.Fatalf("chmod context dir: %v", err)
	}
	defer os.Chmod(contextDir, 0o700)

	err := repository.Write(storedContext("personal", "Personal"))
	if err == nil {
		t.Skip("current user can still write files below mode 000 directories")
	}
	var permissionErr *devcontext.StoragePermissionError
	if !errors.As(err, &permissionErr) {
		t.Fatalf("error = %T %v, want *devcontext.StoragePermissionError", err, err)
	}
	if permissionErr.StorageOperation() != "create temporary file" {
		t.Fatalf("operation = %q, want create temporary file", permissionErr.StorageOperation())
	}
	if permissionErr.StoragePath() != contextDir {
		t.Fatalf("path = %q, want %q", permissionErr.StoragePath(), contextDir)
	}
}

func TestRepositoryListReturnsEmptyForNoContexts(t *testing.T) {
	repository := devcontext.NewRepository(t.TempDir())

	contexts, err := repository.List()
	if err != nil {
		t.Fatalf("list contexts: %v", err)
	}
	if len(contexts) != 0 {
		t.Fatalf("context count = %d, want 0", len(contexts))
	}
}

func TestRepositoryListReturnsContextsInDeterministicOrder(t *testing.T) {
	contextsDir := t.TempDir()
	repository := devcontext.NewRepository(contextsDir)
	company := storedContext("company", "Company")
	personal := storedContext("personal", "Personal")
	client := storedContext("client-a", "Client A")

	writeStoredContext(t, contextsDir, repository, personal)
	writeStoredContext(t, contextsDir, repository, company)
	writeStoredContext(t, contextsDir, repository, client)

	contexts, err := repository.List()
	if err != nil {
		t.Fatalf("list contexts: %v", err)
	}

	want := []devcontext.Context{client, company, personal}
	if !reflect.DeepEqual(contexts, want) {
		t.Fatalf("contexts = %#v, want %#v", contexts, want)
	}
}

func TestRepositoryPersistenceRoundTripsMultipleContexts(t *testing.T) {
	contextsDir := t.TempDir()
	repository := devcontext.NewRepository(contextsDir)
	contexts := []devcontext.Context{
		storedContext("personal", "Personal"),
		storedContext("company", "Company"),
		storedContext("client-a", "Client A"),
	}

	for _, ctx := range contexts {
		writeStoredContext(t, contextsDir, repository, ctx)
	}

	restartedRepository := devcontext.NewRepository(contextsDir)
	got, err := restartedRepository.List()
	if err != nil {
		t.Fatalf("list contexts after restart: %v", err)
	}

	want := []devcontext.Context{contexts[2], contexts[1], contexts[0]}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("contexts = %#v, want %#v", got, want)
	}
	for _, wantContext := range contexts {
		gotContext, err := restartedRepository.Get(wantContext.ID)
		if err != nil {
			t.Fatalf("get context %q after restart: %v", wantContext.ID, err)
		}
		if !reflect.DeepEqual(gotContext, wantContext) {
			t.Fatalf("context %q = %#v, want %#v", wantContext.ID, gotContext, wantContext)
		}
	}
}

func TestConcurrentContextWritesLeaveParseableContext(t *testing.T) {
	contextsDir := t.TempDir()
	repository := devcontext.NewRepository(contextsDir)
	contextID := devcontext.MustID("personal")
	createContextDir(t, contextsDir, contextID)

	values := []devcontext.Context{
		storedContext("personal", "Personal"),
		storedContext("personal", "Personal Updated"),
	}
	var wg sync.WaitGroup
	errs := make(chan error, 40)
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			errs <- repository.Write(values[index%len(values)])
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("write context: %v", err)
		}
	}

	got, err := repository.Get(contextID)
	if err != nil {
		t.Fatalf("get final context: %v", err)
	}
	if !reflect.DeepEqual(got, values[0]) && !reflect.DeepEqual(got, values[1]) {
		t.Fatalf("context = %#v, want one complete written value", got)
	}
}

func TestRepositoryListSkipsMalformedEntries(t *testing.T) {
	contextsDir := t.TempDir()
	repository := devcontext.NewRepository(contextsDir)
	valid := storedContext("personal", "Personal")

	writeStoredContext(t, contextsDir, repository, valid)
	writeFile(t, filepath.Join(contextsDir, "not-a-context"), []byte("ignored"))
	createContextDir(t, contextsDir, devcontext.MustID("broken"))
	writeFile(t, filepath.Join(contextsDir, "broken", "context.toml"), []byte("id = "))
	createContextDir(t, contextsDir, devcontext.MustID("missing-config"))
	createContextDir(t, contextsDir, devcontext.MustID("mismatch"))
	writeFile(t, filepath.Join(contextsDir, "mismatch", "context.toml"), []byte(`
id = "company"
name = "Company"
created_at = 2026-08-13T12:30:00Z

[editor]
type = "vscode"
`))
	if err := os.Mkdir(filepath.Join(contextsDir, "Invalid"), 0o700); err != nil {
		t.Fatalf("create invalid ID dir: %v", err)
	}

	contexts, err := repository.List()
	if err != nil {
		t.Fatalf("list contexts: %v", err)
	}

	want := []devcontext.Context{valid}
	if !reflect.DeepEqual(contexts, want) {
		t.Fatalf("contexts = %#v, want %#v", contexts, want)
	}
}

func TestRepositoryGetDistinguishesErrorCategories(t *testing.T) {
	contextsDir := t.TempDir()
	repository := devcontext.NewRepository(contextsDir)

	_, err := repository.Get(devcontext.ID{})
	if !errors.Is(err, devcontext.ErrInvalidID) {
		t.Fatalf("invalid ID error = %v, want %v", err, devcontext.ErrInvalidID)
	}

	_, err = repository.Get(devcontext.MustID("missing"))
	if !errors.Is(err, devcontext.ErrContextNotFound) {
		t.Fatalf("not found error = %v, want %v", err, devcontext.ErrContextNotFound)
	}

	unreadableID := devcontext.MustID("unreadable")
	createContextDir(t, contextsDir, unreadableID)
	if err := os.Mkdir(filepath.Join(contextsDir, unreadableID.String(), "context.toml"), 0o700); err != nil {
		t.Fatalf("create unreadable context config: %v", err)
	}
	_, err = repository.Get(unreadableID)
	if !errors.Is(err, devcontext.ErrUnreadableContextConfig) {
		t.Fatalf("unreadable config error = %v, want %v", err, devcontext.ErrUnreadableContextConfig)
	}
}

func TestRepositoryGetMissingContextReportsAvailableIDs(t *testing.T) {
	contextsDir := t.TempDir()
	repository := devcontext.NewRepository(contextsDir)
	writeStoredContext(t, contextsDir, repository, storedContext("personal", "Personal"))
	writeStoredContext(t, contextsDir, repository, storedContext("company", "Company"))

	_, err := repository.Get(devcontext.MustID("client-a"))
	if !errors.Is(err, devcontext.ErrContextNotFound) {
		t.Fatalf("error = %v, want %v", err, devcontext.ErrContextNotFound)
	}

	var missingErr *devcontext.MissingContextError
	if !errors.As(err, &missingErr) {
		t.Fatalf("error = %T, want *devcontext.MissingContextError", err)
	}
	if missingErr.ContextID != devcontext.MustID("client-a") {
		t.Fatalf("context ID = %q, want client-a", missingErr.ContextID.String())
	}
	wantAvailable := []devcontext.ID{
		devcontext.MustID("company"),
		devcontext.MustID("personal"),
	}
	if !reflect.DeepEqual(missingErr.AvailableIDs, wantAvailable) {
		t.Fatalf("available IDs = %#v, want %#v", missingErr.AvailableIDs, wantAvailable)
	}
}

func TestRepositoryGetWrapsMalformedConfigAsUnreadable(t *testing.T) {
	contextsDir := t.TempDir()
	repository := devcontext.NewRepository(contextsDir)
	contextID := devcontext.MustID("broken")
	createContextDir(t, contextsDir, contextID)
	writeFile(t, filepath.Join(contextsDir, contextID.String(), "context.toml"), []byte("id = "))

	_, err := repository.Get(contextID)
	if !errors.Is(err, devcontext.ErrUnreadableContextConfig) {
		t.Fatalf("error = %v, want %v", err, devcontext.ErrUnreadableContextConfig)
	}
	if !errors.Is(err, devcontext.ErrInvalidContextConfig) {
		t.Fatalf("error = %v, want %v", err, devcontext.ErrInvalidContextConfig)
	}
}

func writeStoredContext(t *testing.T, contextsDir string, repository devcontext.Repository, ctx devcontext.Context) {
	t.Helper()

	createContextDir(t, contextsDir, ctx.ID)
	if err := repository.Write(ctx); err != nil {
		t.Fatalf("write context %q: %v", ctx.ID, err)
	}
}

func createContextDir(t *testing.T, contextsDir string, contextID devcontext.ID) {
	t.Helper()

	if err := os.MkdirAll(filepath.Join(contextsDir, contextID.String()), 0o700); err != nil {
		t.Fatalf("create context dir %q: %v", contextID, err)
	}
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func storedContext(id string, name string) devcontext.Context {
	return devcontext.Context{
		ID:     devcontext.MustID(id),
		Name:   name,
		Editor: editor.DefaultConfig(),
		Providers: provider.Configs{
			"claude": {Enabled: true},
			"codex":  {Enabled: true},
		},
		Metadata: devcontext.Metadata{
			"kind": "test",
		},
		CreatedAt: time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC),
	}
}
