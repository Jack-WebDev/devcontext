package application

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/provider"
)

const (
	repairActionRecheckProviderFiles       = "recheck-provider-files"
	repairActionRecreateMissingDirectories = "recreate-missing-directories"
	repairActionResetProviderPrefix        = "reset-provider:"
	repairActionResetToolPrefix            = "reset-tool:"
)

func (s *Service) getRepairActions(request GetRepairActionsRequest) (RepairActionsState, error) {
	ctx, paths, err := s.repairContext(request.ContextID)
	if err != nil {
		return RepairActionsState{}, err
	}

	actions := []RepairAction{
		{
			ID: repairActionRecheckProviderFiles, Label: "Re-check provider files",
			Description: "Re-run provider file and identity checks without changing stored files.",
			Targets:     []RepairTarget{},
		},
		{
			ID: repairActionRecreateMissingDirectories, Label: "Recreate missing directories",
			Description: "Create missing context, provider, and selected-tool storage directories without deleting credentials.",
			Targets:     s.missingStorageTargets(ctx, paths),
		},
	}

	for _, providerID := range enabledProviderIDs(ctx) {
		integration, ok := s.dependencies.ProviderRegistry.Get(providerID)
		if !ok {
			continue
		}
		storageDir := paths.ProviderStorageDir(providerID)
		actions = append(actions, destructiveRepairAction(
			repairActionResetProviderPrefix+string(providerID),
			"Reset "+integration.DisplayName()+" storage",
			"Remove only files and folders owned by "+integration.DisplayName()+" in this context.",
			integration.DisplayName()+" storage",
			storageDir,
		))
	}

	toolID := ctx.Tool.DefaultTool
	if registered, ok := s.dependencies.ToolRegistry.Lookup(toolID); ok {
		actions = append(actions, destructiveRepairAction(
			repairActionResetToolPrefix+string(toolID),
			"Reset "+registered.DisplayName+" storage",
			"Remove only files and folders owned by "+registered.DisplayName+" in this context.",
			registered.DisplayName+" storage",
			paths.ToolStorageDir(toolID),
		))
	}

	return RepairActionsState{Actions: actions}, nil
}

func (s *Service) runRepairAction(request RunRepairActionRequest) (RunRepairActionResult, error) {
	ctx, paths, err := s.repairContext(request.ContextID)
	if err != nil {
		return RunRepairActionResult{}, err
	}
	actions, err := s.getRepairActions(GetRepairActionsRequest{ContextID: request.ContextID})
	if err != nil {
		return RunRepairActionResult{}, err
	}
	action, ok := repairActionByID(actions.Actions, request.ActionID)
	if !ok {
		return RunRepairActionResult{}, fmt.Errorf("unknown repair action %q", request.ActionID)
	}
	if action.Destructive && !request.ConfirmDestructive {
		return RunRepairActionResult{}, fmt.Errorf("repair action %q requires confirmation", request.ActionID)
	}

	switch request.ActionID {
	case repairActionRecheckProviderFiles:
		// Diagnostics are recomputed below without changing any files.
	case repairActionRecreateMissingDirectories:
		if err := s.recreateMissingDirectories(ctx, paths); err != nil {
			return RunRepairActionResult{}, err
		}
	default:
		if err := s.resetRepairStorage(ctx, paths, request.ActionID); err != nil {
			return RunRepairActionResult{}, err
		}
	}

	diagnostics, err := s.getDiagnostics(GetDiagnosticsRequest{ContextID: request.ContextID})
	if err != nil {
		return RunRepairActionResult{}, err
	}
	return RunRepairActionResult{ActionID: action.ID, Diagnostics: diagnostics}, nil
}

func (s *Service) repairContext(rawContextID string) (devcontext.Context, filesystem.ContextPaths, error) {
	contextID, err := devcontext.NewID(rawContextID)
	if err != nil {
		return devcontext.Context{}, filesystem.ContextPaths{}, err
	}
	ctx, err := s.dependencies.Contexts.Get(contextID)
	if err != nil {
		return devcontext.Context{}, filesystem.ContextPaths{}, err
	}
	paths, err := filesystem.DeriveContextPaths(s.dependencies.Paths, contextID)
	if err != nil {
		return devcontext.Context{}, filesystem.ContextPaths{}, err
	}
	paths = paths.WithProviderStorageDirs(enabledProviderIDs(ctx)).WithToolStorageDirs([]codingtool.ID{ctx.Tool.DefaultTool})
	return ctx, paths, nil
}

func (s *Service) missingStorageTargets(ctx devcontext.Context, paths filesystem.ContextPaths) []RepairTarget {
	err := filesystem.ValidateContextDirectoryTreeWithRegistries(paths, ctx, s.dependencies.ProviderRegistry, s.dependencies.ToolRegistry)
	storageError := &filesystem.ContextStorageError{}
	if !errors.As(err, &storageError) {
		return []RepairTarget{}
	}
	targets := make([]RepairTarget, 0, len(storageError.Missing))
	for _, missing := range storageError.Missing {
		targets = append(targets, RepairTarget{Label: diagnosticMissingDirectoryLabel(missing), Path: missing.Path, Kind: "directory"})
	}
	return targets
}

func destructiveRepairAction(id, label, description, targetLabel, storageDir string) RepairAction {
	return RepairAction{
		ID: id, Label: label, Description: description,
		Destructive: true, RequiresConfirmation: true,
		Targets: repairStorageTargets(targetLabel, storageDir),
	}
}

func repairStorageTargets(label, storageDir string) []RepairTarget {
	targets := make([]RepairTarget, 0)
	_ = filepath.WalkDir(storageDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || path == storageDir {
			return nil
		}
		kind := "file"
		if entry.IsDir() {
			kind = "directory"
		}
		targets = append(targets, RepairTarget{Label: label, Path: path, Kind: kind})
		return nil
	})
	sort.Slice(targets, func(i, j int) bool { return targets[i].Path < targets[j].Path })
	return targets
}

func (s *Service) recreateMissingDirectories(ctx devcontext.Context, paths filesystem.ContextPaths) error {
	directories := []string{paths.RootDir, paths.ToolStorageRootDir, paths.ToolStorageDir(ctx.Tool.DefaultTool)}
	for _, providerID := range enabledProviderIDs(ctx) {
		directories = append(directories, paths.ProviderStorageDir(providerID))
	}
	for _, directory := range directories {
		if _, err := os.Stat(directory); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("inspect repair directory %q: %w", directory, err)
		}
		if err := os.MkdirAll(directory, s.dependencies.StoragePermissions.DirectoryMode()); err != nil {
			return fmt.Errorf("recreate directory %q: %w", directory, err)
		}
		if err := s.dependencies.StoragePermissions.ApplyDirectory(directory); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) resetRepairStorage(ctx devcontext.Context, paths filesystem.ContextPaths, actionID string) error {
	if providerID, ok := repairProviderID(actionID); ok {
		config, enabled := ctx.Providers[providerID]
		if !enabled || !config.Enabled {
			return fmt.Errorf("provider %q is not enabled for this context", providerID)
		}
		if _, registered := s.dependencies.ProviderRegistry.Get(providerID); !registered {
			return fmt.Errorf("provider %q is not registered", providerID)
		}
		return clearStorageDirectory(paths.ProviderStorageDir(providerID))
	}
	if toolID, ok := repairToolID(actionID); ok {
		if toolID != ctx.Tool.DefaultTool {
			return fmt.Errorf("coding tool %q is not selected for this context", toolID)
		}
		if _, registered := s.dependencies.ToolRegistry.Lookup(toolID); !registered {
			return fmt.Errorf("coding tool %q is not registered", toolID)
		}
		return clearStorageDirectory(paths.ToolStorageDir(toolID))
	}
	return fmt.Errorf("unknown repair action %q", actionID)
}

func clearStorageDirectory(storageDir string) error {
	entries, err := os.ReadDir(storageDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("list storage directory %q: %w", storageDir, err)
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(storageDir, entry.Name())); err != nil {
			return fmt.Errorf("reset storage entry %q: %w", entry.Name(), err)
		}
	}
	return nil
}

func repairActionByID(actions []RepairAction, id string) (RepairAction, bool) {
	for _, action := range actions {
		if action.ID == id {
			return action, true
		}
	}
	return RepairAction{}, false
}

func repairProviderID(actionID string) (provider.ID, bool) {
	if len(actionID) <= len(repairActionResetProviderPrefix) || actionID[:len(repairActionResetProviderPrefix)] != repairActionResetProviderPrefix {
		return "", false
	}
	return provider.ID(actionID[len(repairActionResetProviderPrefix):]), true
}

func repairToolID(actionID string) (codingtool.ID, bool) {
	if len(actionID) <= len(repairActionResetToolPrefix) || actionID[:len(repairActionResetToolPrefix)] != repairActionResetToolPrefix {
		return "", false
	}
	return codingtool.ID(actionID[len(repairActionResetToolPrefix):]), true
}
