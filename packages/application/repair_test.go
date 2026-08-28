package application

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"devctx/packages/core/filesystem"
	devlog "devctx/packages/core/logging"
)

func TestRepairActionsPreviewAndResetOnlyProviderOwnedStorage(t *testing.T) {
	fixture := newApplicationFixture(t)
	logger := &applicationRecordingLogger{}
	fixture.logger = logger
	ctx := fixture.context("personal", "Personal")
	fixture.writeContext(t, ctx)
	paths, err := filesystem.DeriveContextPaths(fixture.paths, ctx.ID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	providerStorage := paths.ProviderStorageDir(fixture.provider.id)
	credentialPath := filepath.Join(providerStorage, "credentials.json")
	writeFile(t, credentialPath, []byte("opaque credential"))

	actions, appErr := fixture.service().GetRepairActions(GetRepairActionsRequest{ContextID: ctx.ID.String()})
	if appErr != nil {
		t.Fatalf("get repair actions: %v", appErr)
	}
	reset := repairActionForTest(t, actions.Actions, repairActionResetProviderPrefix+string(fixture.provider.id))
	if !reset.Destructive || !reset.RequiresConfirmation || !repairTargetsPath(reset.Targets, credentialPath) {
		t.Fatalf("provider reset preview = %#v", reset)
	}

	if _, appErr := fixture.service().RunRepairAction(RunRepairActionRequest{ContextID: ctx.ID.String(), ActionID: reset.ID}); appErr == nil {
		t.Fatal("destructive repair succeeded without confirmation")
	}
	result, appErr := fixture.service().RunRepairAction(RunRepairActionRequest{ContextID: ctx.ID.String(), ActionID: reset.ID, ConfirmDestructive: true})
	if appErr != nil {
		t.Fatalf("run confirmed repair: %v", appErr)
	}
	if result.ActionID != reset.ID {
		t.Fatalf("action ID = %q, want %q", result.ActionID, reset.ID)
	}
	if _, err := os.Stat(credentialPath); !os.IsNotExist(err) {
		t.Fatalf("credential path still exists or could not be checked: %v", err)
	}
	if info, err := os.Stat(providerStorage); err != nil || !info.IsDir() {
		t.Fatalf("provider storage = %#v, %v; want preserved directory", info, err)
	}
	wantEvents := []devlog.EventName{devlog.EventProviderReset, devlog.EventRepairCompleted}
	if got := applicationEventNames(logger.events); !reflect.DeepEqual(got, wantEvents) {
		t.Fatalf("repair events = %#v, want %#v", got, wantEvents)
	}
}

func TestRecreateMissingDirectoriesDoesNotDeleteProviderFiles(t *testing.T) {
	fixture := newApplicationFixture(t)
	ctx := fixture.context("personal", "Personal")
	fixture.writeContext(t, ctx)
	paths, err := filesystem.DeriveContextPaths(fixture.paths, ctx.ID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	credentialPath := filepath.Join(paths.ProviderStorageDir(fixture.provider.id), "credentials.json")
	writeFile(t, credentialPath, []byte("opaque credential"))
	removeAll(t, paths.ToolStorageDir(ctx.Tool.DefaultTool))

	result, appErr := fixture.service().RunRepairAction(RunRepairActionRequest{ContextID: ctx.ID.String(), ActionID: repairActionRecreateMissingDirectories})
	if appErr != nil {
		t.Fatalf("recreate missing directories: %v", appErr)
	}
	if result.ActionID != repairActionRecreateMissingDirectories {
		t.Fatalf("action ID = %q", result.ActionID)
	}
	if info, err := os.Stat(paths.ToolStorageDir(ctx.Tool.DefaultTool)); err != nil || !info.IsDir() {
		t.Fatalf("tool storage = %#v, %v; want recreated directory", info, err)
	}
	if _, err := os.Stat(credentialPath); err != nil {
		t.Fatalf("credential file was changed: %v", err)
	}
}

func repairActionForTest(t *testing.T, actions []RepairAction, id string) RepairAction {
	t.Helper()
	action, ok := repairActionByID(actions, id)
	if !ok {
		t.Fatalf("repair action %q not found in %#v", id, actions)
	}
	return action
}

func repairTargetsPath(targets []RepairTarget, path string) bool {
	for _, target := range targets {
		if target.Path == path {
			return true
		}
	}
	return false
}
