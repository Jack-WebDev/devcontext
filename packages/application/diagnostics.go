package application

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/provider"
)

func (s *Service) getDiagnostics(request GetDiagnosticsRequest) (DiagnosticsState, error) {
	contextID, err := devcontext.NewID(request.ContextID)
	if err != nil {
		return DiagnosticsState{}, err
	}
	ctx, err := s.dependencies.Contexts.Get(contextID)
	if err != nil {
		return DiagnosticsState{}, err
	}
	paths, err := filesystem.DeriveContextPaths(s.dependencies.Paths, ctx.ID)
	if err != nil {
		return DiagnosticsState{}, err
	}
	paths = paths.WithProviderStorageDirs(enabledProviderIDs(ctx)).WithToolStorageDirs([]codingtool.ID{ctx.Tool.DefaultTool})
	entries := s.providerStateEntries(ctx)

	return DiagnosticsState{Groups: []DiagnosticGroup{
		s.contextFilesystemDiagnostics(ctx, paths),
		s.providerDiagnostics(ctx, paths, entries),
		s.toolDiagnostics(ctx, paths),
	}}, nil
}

func (s *Service) contextFilesystemDiagnostics(ctx devcontext.Context, paths filesystem.ContextPaths) DiagnosticGroup {
	checks := []DiagnosticCheck{
		directoryDiagnostic("context-directory", "Context directory", paths.RootDir),
		permissionDiagnostic(paths.RootDir),
		storageCompletenessDiagnostic(ctx, paths, s.dependencies.ProviderRegistry, s.dependencies.ToolRegistry),
	}
	return DiagnosticGroup{ID: "context-filesystem", Label: "Context filesystem", Checks: checks}
}

func (s *Service) providerDiagnostics(ctx devcontext.Context, paths filesystem.ContextPaths, entries []providerStateEntry) DiagnosticGroup {
	checks := make([]DiagnosticCheck, 0)
	for _, entry := range entries {
		if !entry.state.Enabled {
			continue
		}
		prefix := "provider-" + entry.state.ID + "-"
		runtimeContext := providerRuntimeContext(ctx, ctx.Providers[entry.providerID], paths, entry.providerID)
		checks = append(checks,
			providerReadinessDiagnostic(prefix+"readiness", entry),
			directoryDiagnostic(prefix+"storage", entry.state.Name+" storage", runtimeContext.Paths.StorageDir),
		)
		if diagnosticsProvider, ok := entry.provider.(provider.CredentialDiagnosticsProvider); ok {
			for index, credentialFile := range diagnosticsProvider.CredentialDiagnosticFiles(runtimeContext) {
				checks = append(checks, credentialFileDiagnostic(fmt.Sprintf("%scredential-%d", prefix, index), entry.state.Name+" "+credentialFile.Label, credentialFile.Path))
			}
		}
		checks = append(checks, providerIdentityDiagnostic(prefix+"identity", entry.state.Name, entry.state.Identity))
	}
	return DiagnosticGroup{ID: "providers", Label: "Providers", Checks: checks}
}

func (s *Service) toolDiagnostics(ctx devcontext.Context, paths filesystem.ContextPaths) DiagnosticGroup {
	toolID := ctx.Tool.DefaultTool
	toolName := s.dependencies.ToolRegistry.DisplayName(toolID)
	checks := []DiagnosticCheck{
		directoryDiagnostic("tool-storage", toolName+" storage", paths.ToolStorageDir(toolID)),
	}
	registered, ok := s.dependencies.ToolRegistry.Lookup(toolID)
	if !ok {
		checks = append(checks, DiagnosticCheck{
			ID: "tool-executable", Severity: DiagnosticSeverityBlocked, Label: "Selected coding tool",
			Message: "The selected coding tool is not registered.",
		})
		return DiagnosticGroup{ID: "coding-tool", Label: "Selected coding tool", Checks: checks}
	}
	executable, err := registered.Integration.DetectExecutable(ctx.Tool.ConfigFor(toolID))
	if err != nil {
		checks = append(checks, DiagnosticCheck{
			ID: "tool-executable", Severity: DiagnosticSeverityBlocked, Label: toolName + " executable",
			Message: toolName + " could not be found or used for launch.",
		})
	} else {
		checks = append(checks, DiagnosticCheck{
			ID: "tool-executable", Severity: DiagnosticSeverityReady, Label: toolName + " executable",
			Message: toolName + " is available for launch.",
			Details: []DiagnosticDetail{{Label: "Executable", Value: string(executable), IsPath: true}},
		})
	}
	return DiagnosticGroup{ID: "coding-tool", Label: "Selected coding tool", Checks: checks}
}

func directoryDiagnostic(id, label, path string) DiagnosticCheck {
	info, err := os.Stat(path)
	if err != nil {
		return DiagnosticCheck{ID: id, Severity: DiagnosticSeverityBlocked, Label: label, Message: label + " is missing or inaccessible.", Details: pathDetail(path)}
	}
	if !info.IsDir() {
		return DiagnosticCheck{ID: id, Severity: DiagnosticSeverityBlocked, Label: label, Message: label + " is not a directory.", Details: pathDetail(path)}
	}
	return DiagnosticCheck{ID: id, Severity: DiagnosticSeverityReady, Label: label, Message: label + " is available.", Details: pathDetail(path)}
}

func permissionDiagnostic(path string) DiagnosticCheck {
	if runtime.GOOS == "windows" {
		return DiagnosticCheck{ID: "context-permissions", Severity: DiagnosticSeverityReady, Label: "Context permissions", Message: "Context permissions are managed by the operating system.", Details: pathDetail(path)}
	}
	info, err := os.Stat(path)
	if err != nil {
		return DiagnosticCheck{ID: "context-permissions", Severity: DiagnosticSeverityBlocked, Label: "Context permissions", Message: "Context permissions could not be inspected.", Details: pathDetail(path)}
	}
	if info.Mode().Perm()&0o077 != 0 {
		return DiagnosticCheck{ID: "context-permissions", Severity: DiagnosticSeverityNeedsAttention, Label: "Context permissions", Message: "Context storage is accessible beyond the current user.", Details: append(pathDetail(path), DiagnosticDetail{Label: "Mode", Value: info.Mode().Perm().String()})}
	}
	return DiagnosticCheck{ID: "context-permissions", Severity: DiagnosticSeverityReady, Label: "Context permissions", Message: "Context storage is restricted to the current user.", Details: append(pathDetail(path), DiagnosticDetail{Label: "Mode", Value: info.Mode().Perm().String()})}
}

func storageCompletenessDiagnostic(ctx devcontext.Context, paths filesystem.ContextPaths, providers provider.Registry, tools codingtool.Registry) DiagnosticCheck {
	err := filesystem.ValidateContextDirectoryTreeWithRegistries(paths, ctx, providers, tools)
	if err == nil {
		return DiagnosticCheck{ID: "context-storage-completeness", Severity: DiagnosticSeverityReady, Label: "Context storage completeness", Message: "All required context storage is available."}
	}
	storageError := &filesystem.ContextStorageError{}
	if !errors.As(err, &storageError) {
		return DiagnosticCheck{ID: "context-storage-completeness", Severity: DiagnosticSeverityBlocked, Label: "Context storage completeness", Message: "Context storage could not be validated."}
	}
	details := make([]DiagnosticDetail, 0, len(storageError.Missing))
	for _, missing := range storageError.Missing {
		details = append(details, DiagnosticDetail{Label: diagnosticMissingDirectoryLabel(missing), Value: missing.Path, IsPath: true})
	}
	return DiagnosticCheck{ID: "context-storage-completeness", Severity: DiagnosticSeverityBlocked, Label: "Context storage completeness", Message: "Required context storage is missing or incomplete.", Details: details}
}

func providerReadinessDiagnostic(id string, entry providerStateEntry) DiagnosticCheck {
	severity := DiagnosticSeverityNeedsAttention
	message := entry.state.Explanation
	if message == "" {
		message = entry.state.Name + " needs configuration."
	}
	if entry.status.State == provider.StatusConfigured {
		severity = DiagnosticSeverityReady
		message = entry.state.Name + " is configured for this context."
	}
	if entry.status.State == provider.StatusUnavailable {
		severity = DiagnosticSeverityBlocked
		message = entry.state.Name + " readiness could not be determined."
	}
	return DiagnosticCheck{ID: id, Severity: severity, Label: entry.state.Name + " readiness", Message: message}
}

func credentialFileDiagnostic(id, label, path string) DiagnosticCheck {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DiagnosticCheck{ID: id, Severity: DiagnosticSeverityNeedsAttention, Label: label, Message: label + " file is not present.", Details: pathDetail(path)}
		}
		return DiagnosticCheck{ID: id, Severity: DiagnosticSeverityBlocked, Label: label, Message: label + " file could not be read.", Details: pathDetail(path)}
	}
	_ = file.Close()
	return DiagnosticCheck{ID: id, Severity: DiagnosticSeverityReady, Label: label, Message: label + " file is present and readable.", Details: pathDetail(path)}
}

func providerIdentityDiagnostic(id, providerName string, identity ProviderIdentityState) DiagnosticCheck {
	if identity.Status == ProviderIdentityVerified {
		return DiagnosticCheck{ID: id, Severity: DiagnosticSeverityReady, Label: providerName + " account identity", Message: providerName + " account identity is verified.", Details: providerIdentityDetails(identity.Fields)}
	}
	if identity.Status == ProviderIdentityUnavailable {
		return DiagnosticCheck{ID: id, Severity: DiagnosticSeverityNeedsAttention, Label: providerName + " account identity", Message: "Account identity is unavailable."}
	}
	return DiagnosticCheck{ID: id, Severity: DiagnosticSeverityNeedsAttention, Label: providerName + " account identity", Message: "Account identity has not been verified."}
}

func providerIdentityDetails(fields []ProviderMetadataField) []DiagnosticDetail {
	details := make([]DiagnosticDetail, 0, len(fields))
	for _, field := range fields {
		details = append(details, DiagnosticDetail{Label: field.Label, Value: field.Value})
	}
	return details
}

func pathDetail(path string) []DiagnosticDetail {
	return []DiagnosticDetail{{Label: "Location", Value: path, IsPath: true}}
}

func diagnosticMissingDirectoryLabel(missing filesystem.MissingContextDirectory) string {
	if missing.ProviderDisplayName != "" {
		return missing.ProviderDisplayName + " storage"
	}
	if missing.ToolDisplayName != "" {
		return missing.ToolDisplayName + " storage"
	}
	return "Context storage"
}
