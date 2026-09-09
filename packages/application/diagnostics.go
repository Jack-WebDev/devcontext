package application

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	codingtool "devctx/packages/core/codingtool"
	"devctx/packages/core/config"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/provider"
)

func (s *Service) getDiagnostics(request GetDiagnosticsRequest) (DiagnosticsState, error) {
	if request.ContextID == "" {
		return s.systemDiagnostics()
	}
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
		s.contextFilesDiagnostics(ctx, paths),
		s.contextIsolationDiagnostics(ctx, paths),
		s.toolDiagnostics(ctx, paths),
		s.contextBindingsDiagnostics(ctx),
		s.contextEnvironmentDiagnostics(ctx, paths, entries),
	}}, nil
}

func (s *Service) systemDiagnostics() (DiagnosticsState, error) {
	layout, err := config.DevContextHomeLayout(s.dependencies.Paths)
	if err != nil {
		return DiagnosticsState{}, err
	}
	return DiagnosticsState{Groups: []DiagnosticGroup{
		{ID: "configuration", Label: "Configuration", Checks: []DiagnosticCheck{globalConfigDiagnostic(s.dependencies.ConfigPath)}},
		{ID: "storage", Label: "Storage", Checks: []DiagnosticCheck{
			directoryDiagnostic("app-home", "Dev Context storage", layout.HomeDir),
			directoryDiagnostic("contexts-directory", "Contexts storage", layout.ContextsDir),
			directoryDiagnostic("logs-directory", "Logs storage", layout.LogsDir),
		}},
		s.systemToolDiagnostics(),
		s.systemIntegrationDiagnostics(),
		{ID: "updates", Label: "Updates", Checks: []DiagnosticCheck{{
			ID: "update-service", Severity: DiagnosticSeverityReady, Label: "Update checks",
			Message:    "This build does not include an update service.",
			ActionHint: "Check the release channel used to install Dev Context for updates.",
		}}},
	}}, nil
}

func globalConfigDiagnostic(path string) DiagnosticCheck {
	if _, err := config.ReadGlobalConfigFile(path); err != nil {
		return DiagnosticCheck{ID: "global-config", Severity: DiagnosticSeverityBlocked, Label: "Global configuration", Message: "Global configuration is missing, unreadable, or invalid.", Details: pathDetail(path), ActionHint: "Restore the configuration file from a known-good copy or recreate it before changing settings."}
	}
	return DiagnosticCheck{ID: "global-config", Severity: DiagnosticSeverityReady, Label: "Global configuration", Message: "Global configuration is available."}
}

func (s *Service) systemToolDiagnostics() DiagnosticGroup {
	checks := make([]DiagnosticCheck, 0, len(s.dependencies.ToolRegistry.All()))
	for _, tool := range s.dependencies.ToolRegistry.All() {
		if _, err := tool.Integration.DetectExecutable(codingtool.Config{}); err != nil {
			checks = append(checks, DiagnosticCheck{ID: "tool-" + string(tool.Integration.ID()), Severity: DiagnosticSeverityNeedsAttention, Label: tool.DisplayName, Message: tool.DisplayName + " is registered but is not currently available.", ActionHint: "Install or configure " + tool.DisplayName + " before selecting it for a context."})
			continue
		}
		checks = append(checks, DiagnosticCheck{ID: "tool-" + string(tool.Integration.ID()), Severity: DiagnosticSeverityReady, Label: tool.DisplayName, Message: tool.DisplayName + " is available."})
	}
	return DiagnosticGroup{ID: "tools", Label: "Tools", Checks: checks}
}

func (s *Service) systemIntegrationDiagnostics() DiagnosticGroup {
	providers := s.dependencies.ProviderRegistry.All()
	checks := make([]DiagnosticCheck, 0, len(providers))
	for _, integration := range providers {
		checks = append(checks, DiagnosticCheck{ID: "integration-" + string(integration.ID()), Severity: DiagnosticSeverityReady, Label: integration.DisplayName(), Message: integration.DisplayName() + " is available to configure in a context."})
	}
	return DiagnosticGroup{ID: "integrations", Label: "Integrations", Checks: checks}
}

func (s *Service) contextFilesDiagnostics(ctx devcontext.Context, paths filesystem.ContextPaths) DiagnosticGroup {
	checks := []DiagnosticCheck{
		directoryDiagnostic("context-directory", "Context directory", paths.RootDir),
		fileDiagnostic("context-config", "Context configuration", paths.ConfigPath),
	}
	return DiagnosticGroup{ID: "context-files", Label: "Context files", Checks: checks}
}

func (s *Service) contextIsolationDiagnostics(ctx devcontext.Context, paths filesystem.ContextPaths) DiagnosticGroup {
	return DiagnosticGroup{ID: "isolation", Label: "Isolation", Checks: []DiagnosticCheck{
		permissionDiagnostic(paths.RootDir),
		storageCompletenessDiagnostic(ctx, paths, s.dependencies.ProviderRegistry, s.dependencies.ToolRegistry),
	}}
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
			Message: toolName + " could not be found or used for launch.", ActionHint: "Install or configure " + toolName + " before launching this context.",
		})
	} else {
		checks = append(checks, DiagnosticCheck{
			ID: "tool-executable", Severity: DiagnosticSeverityReady, Label: toolName + " executable",
			Message: toolName + " is available for launch.",
			Details: []DiagnosticDetail{{Label: "Executable", Value: string(executable), IsPath: true}},
		})
	}
	return DiagnosticGroup{ID: "tools", Label: "Tools", Checks: checks}
}

func (s *Service) contextBindingsDiagnostics(ctx devcontext.Context) DiagnosticGroup {
	bindings, err := s.dependencies.Projects.List()
	if err != nil {
		return DiagnosticGroup{ID: "bindings", Label: "Project bindings", Checks: []DiagnosticCheck{{ID: "project-bindings", Severity: DiagnosticSeverityBlocked, Label: "Project bindings", Message: "Project bindings could not be read.", ActionHint: "Review the project bindings file before launching remembered projects."}}}
	}
	count := 0
	for _, binding := range bindings {
		if binding.ContextID == ctx.ID {
			count++
		}
	}
	message := "No projects are remembered for this context."
	if count == 1 {
		message = "One project is remembered for this context."
	}
	if count > 1 {
		message = fmt.Sprintf("%d projects are remembered for this context.", count)
	}
	return DiagnosticGroup{ID: "bindings", Label: "Project bindings", Checks: []DiagnosticCheck{{ID: "project-bindings", Severity: DiagnosticSeverityReady, Label: "Project bindings", Message: message}}}
}

func (s *Service) contextEnvironmentDiagnostics(ctx devcontext.Context, paths filesystem.ContextPaths, entries []providerStateEntry) DiagnosticGroup {
	checks := make([]DiagnosticCheck, 0, len(entries)*4)
	for _, entry := range entries {
		if !entry.state.Enabled {
			continue
		}
		prefix := "provider-" + entry.state.ID + "-"
		check := providerReadinessDiagnostic(prefix+"readiness", entry)
		check.ActionHint = "Configure or re-check " + entry.state.Name + " in this context."
		runtimeContext := providerRuntimeContext(ctx, ctx.Providers[entry.providerID], paths, entry.providerID)
		checks = append(checks, check, directoryDiagnostic(prefix+"storage", entry.state.Name+" storage", runtimeContext.Paths.StorageDir))
		if diagnosticsProvider, ok := entry.provider.(provider.CredentialDiagnosticsProvider); ok {
			for index, credentialFile := range diagnosticsProvider.CredentialDiagnosticFiles(runtimeContext) {
				checks = append(checks, credentialFileDiagnostic(fmt.Sprintf("%scredential-%d", prefix, index), entry.state.Name+" "+credentialFile.Label, credentialFile.Path))
			}
		}
		checks = append(checks, providerIdentityDiagnostic(prefix+"identity", entry.state.Name, entry.state.Identity))
	}
	if len(checks) == 0 {
		checks = append(checks, DiagnosticCheck{ID: "launch-environment", Severity: DiagnosticSeverityReady, Label: "Launch environment", Message: "No provider environment is configured for this context."})
	}
	return DiagnosticGroup{ID: "environment", Label: "Environment", Checks: checks}
}

func directoryDiagnostic(id, label, path string) DiagnosticCheck {
	info, err := os.Stat(path)
	if err != nil {
		return DiagnosticCheck{ID: id, Severity: DiagnosticSeverityBlocked, Label: label, Message: label + " is missing or inaccessible.", Details: pathDetail(path), ActionHint: "Restore or recreate this local storage location before continuing."}
	}
	if !info.IsDir() {
		return DiagnosticCheck{ID: id, Severity: DiagnosticSeverityBlocked, Label: label, Message: label + " is not a directory.", Details: pathDetail(path), ActionHint: "Restore this local storage location as a directory before continuing."}
	}
	return DiagnosticCheck{ID: id, Severity: DiagnosticSeverityReady, Label: label, Message: label + " is available.", Details: pathDetail(path)}
}

func fileDiagnostic(id, label, path string) DiagnosticCheck {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return DiagnosticCheck{ID: id, Severity: DiagnosticSeverityBlocked, Label: label, Message: label + " is missing or inaccessible.", Details: pathDetail(path), ActionHint: "Restore the context configuration before launching this context."}
	}
	return DiagnosticCheck{ID: id, Severity: DiagnosticSeverityReady, Label: label, Message: label + " is available."}
}

func permissionDiagnostic(path string) DiagnosticCheck {
	if runtime.GOOS == "windows" {
		return DiagnosticCheck{ID: "context-permissions", Severity: DiagnosticSeverityReady, Label: "Context permissions", Message: "Context permissions are managed by the operating system.", Details: pathDetail(path)}
	}
	info, err := os.Stat(path)
	if err != nil {
		return DiagnosticCheck{ID: "context-permissions", Severity: DiagnosticSeverityBlocked, Label: "Context permissions", Message: "Context permissions could not be inspected.", Details: pathDetail(path), ActionHint: "Check that the current user can access this context's local storage."}
	}
	if info.Mode().Perm()&0o077 != 0 {
		return DiagnosticCheck{ID: "context-permissions", Severity: DiagnosticSeverityNeedsAttention, Label: "Context permissions", Message: "Context storage is accessible beyond the current user.", Details: append(pathDetail(path), DiagnosticDetail{Label: "Mode", Value: info.Mode().Perm().String()}), ActionHint: "Restrict this context's local storage to the current user if shared access is unintended."}
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
		return DiagnosticCheck{ID: "context-storage-completeness", Severity: DiagnosticSeverityBlocked, Label: "Context storage completeness", Message: "Context storage could not be validated.", ActionHint: "Use Recreate missing directories or inspect the context's local storage."}
	}
	details := make([]DiagnosticDetail, 0, len(storageError.Missing))
	for _, missing := range storageError.Missing {
		details = append(details, DiagnosticDetail{Label: diagnosticMissingDirectoryLabel(missing), Value: missing.Path, IsPath: true})
	}
	return DiagnosticCheck{ID: "context-storage-completeness", Severity: DiagnosticSeverityBlocked, Label: "Context storage completeness", Message: "Required context storage is missing or incomplete.", Details: details, ActionHint: "Use Recreate missing directories to restore the required empty storage directories."}
}

func providerReadinessDiagnostic(id string, entry providerStateEntry) DiagnosticCheck {
	severity := DiagnosticSeverityNeedsAttention
	message := entry.state.Explanation
	if message == "" {
		message = entry.state.Name + " needs configuration."
	}
	if entry.status.State == provider.StatusConfigured {
		message = entry.state.Name + " local credential state is present, but usable authentication has not been verified."
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
			return DiagnosticCheck{ID: id, Severity: DiagnosticSeverityNeedsAttention, Label: label, Message: label + " file is not present.", Details: pathDetail(path), ActionHint: "Sign in or configure this integration in the selected context."}
		}
		return DiagnosticCheck{ID: id, Severity: DiagnosticSeverityBlocked, Label: label, Message: label + " file could not be read.", Details: pathDetail(path), ActionHint: "Check access to this context's integration storage."}
	}
	_ = file.Close()
	return DiagnosticCheck{ID: id, Severity: DiagnosticSeverityReady, Label: label, Message: label + " file is present and readable.", Details: pathDetail(path)}
}

func providerIdentityDiagnostic(id, providerName string, identity ProviderIdentityState) DiagnosticCheck {
	if identity.Status == ProviderIdentityObserved {
		return DiagnosticCheck{ID: id, Severity: DiagnosticSeverityReady, Label: providerName + " account identity", Message: providerName + " account metadata was observed locally.", Details: providerIdentityDetails(identity.Fields)}
	}
	if identity.Status == ProviderIdentityUnavailable {
		return DiagnosticCheck{ID: id, Severity: DiagnosticSeverityNeedsAttention, Label: providerName + " account identity", Message: "Account identity is unavailable.", ActionHint: "Re-check the integration when its account information is available."}
	}
	return DiagnosticCheck{ID: id, Severity: DiagnosticSeverityNeedsAttention, Label: providerName + " account identity", Message: "Account identity has not been observed.", ActionHint: "Sign in or re-check this integration in the selected context."}
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
