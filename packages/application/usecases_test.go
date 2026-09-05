package application

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/launcher"
	devlog "devctx/packages/core/logging"
	"devctx/packages/core/project"
	"devctx/packages/core/provider"
	coreRunning "devctx/packages/core/running"
)

func metadataValueForTest(fields []ProviderMetadataField, label string) string {
	for _, field := range fields {
		if field.Label == label {
			return field.Value
		}
	}
	return ""
}

func TestGetLaunchStateReturnsBoundProjectState(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))
	fixture.writeBindings(t, project.Binding{
		ProjectPath: project.Path(fixture.projectDir),
		ContextID:   devcontext.MustID("personal"),
		CreatedAt:   fixture.now,
	})

	state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get launch state: %v", appErr)
	}

	if state.Project != (ProjectState{Name: "current", Path: fixture.projectDir}) {
		t.Fatalf("project = %#v, want current project", state.Project)
	}
	if !state.Binding.Bound || state.Binding.ContextID != "personal" {
		t.Fatalf("binding = %#v, want personal binding", state.Binding)
	}
	if state.SelectedContextID != "personal" {
		t.Fatalf("selected context = %q, want personal", state.SelectedContextID)
	}
	if state.Confidence == nil || state.Confidence.ContextID != "personal" {
		t.Fatalf("confidence = %#v, want personal confidence", state.Confidence)
	}
	if state.SelectionRequired {
		t.Fatal("selection required = true, want false")
	}
	if state.FirstRun {
		t.Fatal("first run = true, want false")
	}
	if state.ResolutionSource != string(launcher.ResolutionSourceProjectBinding) {
		t.Fatalf("resolution source = %q, want project binding", state.ResolutionSource)
	}
	if len(state.Contexts) != 1 {
		t.Fatalf("context count = %d, want 1", len(state.Contexts))
	}
	contextState := state.Contexts[0]
	if contextState.ID != "personal" ||
		contextState.Name != "Personal" ||
		contextState.Tool != (ToolState{ID: "fake-editor", Name: "Fake Tool", Status: LaunchConfidenceReady, Message: "Fake Tool is available for launch."}) ||
		!reflect.DeepEqual(contextState.AvailableTools, []ToolOption{{ID: "fake-editor", Name: "Fake Tool"}}) ||
		!reflect.DeepEqual(contextState.Providers, []ProviderState{
			{
				ID:      "fake",
				Name:    "Fake Provider",
				Enabled: true,
				State:   ProviderReadinessReady,
				SetupAction: &ProviderSetupAction{
					State:   ProviderSetupWaitingForSignIn,
					Label:   "Waiting for sign-in",
					Message: "Waiting for Fake Provider sign-in verification.",
				},
				Identity: ProviderIdentityState{
					Status:  ProviderIdentityUnavailable,
					Message: "Account identity unavailable.",
				},
			},
		}) ||
		!reflect.DeepEqual(contextState.Metadata, map[string]string{"accent": "blue"}) {
		t.Fatalf("context state = %#v, want personal context identity/readiness", contextState)
	}
	if contextState.Confidence.ContextID != "personal" {
		t.Fatalf("context confidence = %#v, want personal confidence", contextState.Confidence)
	}
}

func TestGetHomeDashboardReturnsCurrentProjectAndContextSummary(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))
	fixture.writeBindings(t, project.Binding{
		ProjectPath: project.Path(fixture.projectDir),
		ContextID:   devcontext.MustID("personal"),
		CreatedAt:   fixture.now,
	})

	dashboard, appErr := fixture.service().GetHomeDashboard(GetHomeDashboardRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get home dashboard: %v", appErr)
	}
	if dashboard.Project != (ProjectState{Name: "current", Path: fixture.projectDir}) {
		t.Fatalf("project = %#v", dashboard.Project)
	}
	if dashboard.CurrentContext == nil || dashboard.CurrentContext.ID != "personal" || dashboard.CurrentContext.Name != "Personal" {
		t.Fatalf("current context = %#v", dashboard.CurrentContext)
	}
	if dashboard.CurrentContext.Confidence.ContextID != "personal" {
		t.Fatalf("current context confidence = %#v", dashboard.CurrentContext.Confidence)
	}
	if len(dashboard.RecentProjects) != 0 || dashboard.Running.Count != 0 || dashboard.Activity.Count != 0 {
		t.Fatalf("future dashboard summaries = %#v", dashboard)
	}
}

func TestLaunchProjectRecordsRecentProjectAfterSuccessfulLaunch(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))

	_, appErr := fixture.service().LaunchProject(LaunchProjectRequest{ProjectPath: ".", ContextID: "personal"})
	if appErr != nil {
		t.Fatalf("launch project: %v", appErr)
	}

	recents, err := project.NewRecentRepository(fixture.recentsPath).List()
	if err != nil {
		t.Fatalf("list recent projects: %v", err)
	}
	if !reflect.DeepEqual(recents, []project.RecentProject{{
		ProjectPath:    project.Path(fixture.projectDir),
		ContextID:      devcontext.MustID("personal"),
		LastLaunchedAt: fixture.now,
	}}) {
		t.Fatalf("recent projects = %#v", recents)
	}
}

func TestGetRecentProjectsReturnsNewestFirstWithContextMetadata(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))
	fixture.writeContext(t, fixture.context("company", "Company"))
	older := fixture.now.Add(-time.Hour)
	newer := fixture.now.Add(-time.Minute)
	if err := project.WriteRecentProjectsFile(fixture.recentsPath, []project.RecentProject{
		{ProjectPath: project.Path(filepath.Join(fixture.root, "projects", "older")), ContextID: devcontext.MustID("personal"), LastLaunchedAt: older},
		{ProjectPath: project.Path(filepath.Join(fixture.root, "projects", "newer")), ContextID: devcontext.MustID("company"), LastLaunchedAt: newer},
		{ProjectPath: project.Path(filepath.Join(fixture.root, "projects", "removed")), ContextID: devcontext.MustID("removed"), LastLaunchedAt: fixture.now},
	}); err != nil {
		t.Fatalf("write recents: %v", err)
	}

	result, appErr := fixture.service().GetRecentProjects()
	if appErr != nil {
		t.Fatalf("get recent projects: %v", appErr)
	}
	want := []RecentProjectState{
		{Project: ProjectState{Name: "removed", Path: filepath.Join(fixture.root, "projects", "removed")}, ContextID: "removed", LastLaunchedAt: fixture.now},
		{Project: ProjectState{Name: "newer", Path: filepath.Join(fixture.root, "projects", "newer")}, ContextID: "company", ContextName: "Company", LastLaunchedAt: newer},
		{Project: ProjectState{Name: "older", Path: filepath.Join(fixture.root, "projects", "older")}, ContextID: "personal", ContextName: "Personal", LastLaunchedAt: older},
	}
	if !reflect.DeepEqual(result.Projects, want) {
		t.Fatalf("recent projects = %#v, want %#v", result.Projects, want)
	}
}

func TestGetContextsReturnsUsageAndReadinessSummaries(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))
	fixture.writeContext(t, fixture.context("company", "Company"))
	fixture.writeBindings(t,
		project.Binding{ProjectPath: project.Path(filepath.Join(fixture.root, "projects", "one")), ContextID: devcontext.MustID("personal"), CreatedAt: fixture.now},
		project.Binding{ProjectPath: project.Path(filepath.Join(fixture.root, "projects", "two")), ContextID: devcontext.MustID("personal"), CreatedAt: fixture.now},
		project.Binding{ProjectPath: project.Path(filepath.Join(fixture.root, "projects", "three")), ContextID: devcontext.MustID("removed"), CreatedAt: fixture.now},
	)
	older := fixture.now.Add(-time.Hour)
	newer := fixture.now.Add(-time.Minute)
	if err := project.WriteRecentProjectsFile(fixture.recentsPath, []project.RecentProject{
		{ProjectPath: project.Path(filepath.Join(fixture.root, "projects", "one")), ContextID: devcontext.MustID("personal"), LastLaunchedAt: older},
		{ProjectPath: project.Path(filepath.Join(fixture.root, "projects", "two")), ContextID: devcontext.MustID("personal"), LastLaunchedAt: newer},
		{ProjectPath: project.Path(filepath.Join(fixture.root, "projects", "three")), ContextID: devcontext.MustID("removed"), LastLaunchedAt: fixture.now},
	}); err != nil {
		t.Fatalf("write recents: %v", err)
	}

	result, appErr := fixture.service().GetContexts()
	if appErr != nil {
		t.Fatalf("get contexts: %v", appErr)
	}
	if len(result.Contexts) != 2 {
		t.Fatalf("context count = %d, want 2", len(result.Contexts))
	}
	personal := result.Contexts[1]
	if personal.Context.ID != "personal" || personal.Context.Tool.ID != "fake-editor" ||
		personal.Context.Confidence.ContextID != "personal" || personal.ProjectCount != 2 {
		t.Fatalf("personal context summary = %#v", personal)
	}
	if len(personal.EnabledProviders) != 1 || !personal.EnabledProviders[0].Enabled || personal.EnabledProviders[0].ID != "fake" {
		t.Fatalf("enabled providers = %#v, want fake provider", personal.EnabledProviders)
	}
	if personal.LastUsedAt == nil || !personal.LastUsedAt.Equal(newer) {
		t.Fatalf("personal last used = %#v, want %s", personal.LastUsedAt, newer)
	}
	company := result.Contexts[0]
	if company.Context.ID != "company" || company.ProjectCount != 0 || company.LastUsedAt != nil {
		t.Fatalf("company context summary = %#v", company)
	}
}

func TestGetContextDetailsReturnsConfiguredContextMetadata(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))
	fixture.writeBindings(t, project.Binding{
		ProjectPath: project.Path(filepath.Join(fixture.root, "projects", "one")),
		ContextID:   devcontext.MustID("personal"),
		CreatedAt:   fixture.now,
	})
	lastUsed := fixture.now.Add(-time.Minute)
	if err := project.WriteRecentProjectsFile(fixture.recentsPath, []project.RecentProject{{
		ProjectPath:    project.Path(filepath.Join(fixture.root, "projects", "one")),
		ContextID:      devcontext.MustID("personal"),
		LastLaunchedAt: lastUsed,
	}}); err != nil {
		t.Fatalf("write recents: %v", err)
	}

	details, appErr := fixture.service().GetContextDetails(GetContextDetailsRequest{ContextID: "personal"})
	if appErr != nil {
		t.Fatalf("get context details: %v", appErr)
	}
	paths, err := filesystem.DeriveContextPaths(fixture.paths, devcontext.MustID("personal"))
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	if details.Context.ID != "personal" || details.Context.Tool.ID != "fake-editor" ||
		details.Location != paths.RootDir || !details.CreatedAt.Equal(fixture.now) || details.ProjectCount != 1 {
		t.Fatalf("context details = %#v", details)
	}
	if details.LastUsedAt == nil || !details.LastUsedAt.Equal(lastUsed) {
		t.Fatalf("last used = %#v, want %s", details.LastUsedAt, lastUsed)
	}
	if len(details.EnabledProviders) != 1 || details.EnabledProviders[0].ID != "fake" {
		t.Fatalf("enabled providers = %#v", details.EnabledProviders)
	}
}

func TestGetTrustCenterReportsActualLocalBoundaries(t *testing.T) {
	fixture := newApplicationFixture(t)
	ctx := fixture.context("personal", "Personal")
	ctx.Providers = provider.Configs{fixture.provider.ID(): {Enabled: true}}
	fixture.writeContext(t, ctx)
	fixture.writeBindings(t, project.Binding{ProjectPath: project.Path(fixture.projectDir), ContextID: ctx.ID, CreatedAt: fixture.now})

	state, appErr := fixture.service().GetTrustCenter()
	if appErr != nil {
		t.Fatalf("get trust center: %v", appErr)
	}
	if state.CredentialSync.Enabled || len(state.Contexts) != 1 || len(state.ProjectMappings) != 1 || len(state.IntegrationBoundaries) != 1 {
		t.Fatalf("trust center = %#v", state)
	}
	protection := state.Contexts[0]
	if protection.Name != "Personal" || protection.Tool.Isolation.Status != LaunchConfidenceReady || len(protection.Providers) != 1 || protection.Providers[0].Isolation.Status != LaunchConfidenceReady {
		t.Fatalf("context protection = %#v", protection)
	}
	if state.ProjectMappings[0].ContextName != "Personal" || state.ProjectMappings[0].Project.Path != fixture.projectDir {
		t.Fatalf("project mapping = %#v", state.ProjectMappings[0])
	}
	if state.IntegrationBoundaries[0].StatusDataAvailable {
		t.Fatalf("integration boundary = %#v, want no fake-tool status-data export", state.IntegrationBoundaries[0])
	}
}

func TestGetLaunchStateDerivesProviderSetupActions(t *testing.T) {
	tests := []struct {
		name         string
		status       provider.Status
		hasIdentity  bool
		wantState    ProviderSetupState
		wantLabel    string
		wantNoAction bool
	}{
		{
			name:      "not configured opens setup",
			status:    provider.NotConfiguredStatus("credentials are missing"),
			wantState: ProviderSetupOpenAndConfigure,
			wantLabel: "Open and configure",
		},
		{
			name:      "configured provider awaits sign-in verification",
			status:    provider.ConfiguredStatus(),
			wantState: ProviderSetupWaitingForSignIn,
			wantLabel: "Waiting for sign-in",
		},
		{
			name:        "configured verified provider is verified",
			status:      provider.ConfiguredStatus(),
			hasIdentity: true,
			wantState:   ProviderSetupVerified,
			wantLabel:   "Verified",
		},
		{
			name:         "unavailable provider has no setup action",
			status:       provider.UnavailableStatus("status unavailable"),
			wantNoAction: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newApplicationFixture(t)
			fixture.provider.statusByContext = map[string]provider.Status{"personal": tt.status}
			fixture.provider.hasIdentity = tt.hasIdentity
			fixture.provider.identity = provider.Identity{Fields: []provider.MetadataField{{Label: "Email", Value: "developer@example.com"}}}
			fixture.writeContext(t, fixture.context("personal", "Personal"))

			state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
			if appErr != nil {
				t.Fatalf("get launch state: %v", appErr)
			}
			action := state.Contexts[0].Providers[0].SetupAction
			if tt.wantNoAction {
				if action != nil {
					t.Fatalf("setup action = %#v, want nil", action)
				}
				return
			}
			if action == nil || action.State != tt.wantState || action.Label != tt.wantLabel || action.Message == "" {
				t.Fatalf("setup action = %#v", action)
			}
		})
	}
}

func TestGetLaunchStateExposesContextDescription(t *testing.T) {
	fixture := newApplicationFixture(t)
	ctx := fixture.context("personal", "Personal")
	ctx.Metadata["description"] = "Personal projects"
	fixture.writeContext(t, ctx)

	state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get launch state: %v", appErr)
	}
	if len(state.Contexts) != 1 || state.Contexts[0].Description != "Personal projects" {
		t.Fatalf("context descriptions = %#v, want Personal projects", state.Contexts)
	}
}

func TestSecondRegisteredToolWorksAcrossStateAndLaunch(t *testing.T) {
	fixture := newApplicationFixture(t)
	secondTool := &applicationSecondTool{}
	fixture.toolRegistry = codingtool.MustNewRegistry([]codingtool.RegisteredTool{
		{Integration: fixture.editor, DisplayName: "Fake Tool"},
		{Integration: secondTool, DisplayName: "Second Tool"},
	}, fixture.editor.ID())

	ctx := fixture.context("personal", "Personal")
	ctx.Tool = codingtool.LaunchTarget{
		DefaultTool: secondTool.ID(),
		Tools: map[codingtool.ID]codingtool.Config{
			secondTool.ID(): {Options: map[string]string{"mode": "isolated"}},
		},
	}
	fixture.writeContext(t, ctx)
	fixture.writeBindings(t, project.Binding{ProjectPath: project.Path(fixture.projectDir), ContextID: ctx.ID, CreatedAt: fixture.now})

	state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get launch state: %v", appErr)
	}
	contextState := state.Contexts[0]
	if contextState.Tool != (ToolState{ID: "second-tool", Name: "Second Tool", Status: LaunchConfidenceReady, Message: "Second Tool is available for launch."}) {
		t.Fatalf("tool state = %#v", contextState.Tool)
	}
	if !reflect.DeepEqual(contextState.AvailableTools, []ToolOption{{ID: "fake-editor", Name: "Fake Tool"}, {ID: "second-tool", Name: "Second Tool"}}) {
		t.Fatalf("available tools = %#v", contextState.AvailableTools)
	}
	assertConfidenceCheck(t, contextState.Confidence.Checks, LaunchConfidenceCheck{Component: LaunchConfidenceCheckTool, ToolID: "second-tool", Severity: LaunchConfidenceReady, Label: "Second Tool", Message: "Second Tool is available for launch."})
	assertConfidenceCheck(t, contextState.Confidence.Checks, LaunchConfidenceCheck{Component: LaunchConfidenceCheckIsolation, ToolID: "second-tool", Severity: LaunchConfidenceReady, Label: "Second Tool isolation", Message: "Second Tool isolation storage is ready."})

	_, appErr = fixture.service().LaunchProject(LaunchProjectRequest{ProjectPath: ".", ContextID: "personal"})
	if appErr != nil {
		t.Fatalf("launch project: %v", appErr)
	}
	if len(secondTool.requests) != 1 {
		t.Fatalf("second tool requests = %#v", secondTool.requests)
	}
	request := secondTool.requests[0]
	if request.Config.Options["mode"] != "isolated" || request.Paths.StorageDir == "" {
		t.Fatalf("second tool command request = %#v", request)
	}
	if len(fixture.process.requests) != 1 || fixture.process.requests[0].Tool != (launcher.Tool{ID: "second-tool", DisplayName: "Second Tool"}) {
		t.Fatalf("process requests = %#v", fixture.process.requests)
	}
}

func TestGetLaunchStateReturnsConfidenceForSelectedContext(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.provider = &applicationFakeProvider{
		id:          provider.CodexID,
		displayName: "Codex",
		statusByContext: map[string]provider.Status{
			"personal": provider.NotConfiguredStatus("Codex is not authenticated."),
		},
	}
	fixture.editor.executable = "/fixture/code"
	ctx := fixture.context("personal", "Personal")
	ctx.Providers = provider.Configs{
		provider.CodexID: {Enabled: true},
	}
	fixture.writeContext(t, ctx)
	contextPaths, err := filesystem.DeriveContextPaths(fixture.paths, ctx.ID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	removeAll(t, contextPaths.ToolStorageDir(ctx.Tool.DefaultTool))
	fixture.writeBindings(t, project.Binding{
		ProjectPath: project.Path(fixture.projectDir),
		ContextID:   devcontext.MustID("personal"),
		CreatedAt:   fixture.now,
	})

	state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get launch state: %v", appErr)
	}

	if state.Confidence == nil {
		t.Fatal("confidence = nil, want selected context confidence")
	}
	if state.Confidence.ContextID != "personal" {
		t.Fatalf("confidence context = %q, want personal", state.Confidence.ContextID)
	}
	if state.Confidence.Status != LaunchConfidenceBlocked {
		t.Fatalf("confidence status = %q, want blocked because selected tool storage is absent", state.Confidence.Status)
	}
	wantChecks := []LaunchConfidenceCheck{
		{
			Component:  LaunchConfidenceCheckProvider,
			ProviderID: "codex",
			Severity:   LaunchConfidenceNeedsAttention,
			Label:      "Codex",
			Message:    "Codex is not authenticated.",
			ActionHint: "Open and configure Codex for this context.",
		},
		{
			Component: LaunchConfidenceCheckTool,
			ToolID:    "fake-editor",
			Severity:  LaunchConfidenceReady,
			Label:     "Fake Tool",
			Message:   "Fake Tool is available for launch.",
		},
		{
			Component: LaunchConfidenceCheckIsolation,
			Severity:  LaunchConfidenceReady,
			Label:     "Context storage",
			Message:   "Context storage is ready.",
		},
		{
			Component:  LaunchConfidenceCheckIsolation,
			ProviderID: "codex",
			Severity:   LaunchConfidenceReady,
			Label:      "Codex isolation",
			Message:    "Codex isolation storage is ready.",
		},
		{
			Component:  LaunchConfidenceCheckIsolation,
			Severity:   LaunchConfidenceBlocked,
			ToolID:     "fake-editor",
			Label:      "Fake Tool isolation",
			Message:    "Fake Tool isolation storage is not ready.",
			ActionHint: "Run diagnostics to repair context storage.",
		},
	}
	if !reflect.DeepEqual(state.Confidence.Checks, wantChecks) {
		t.Fatalf("confidence checks = %#v, want %#v", state.Confidence.Checks, wantChecks)
	}
}

func TestGetLaunchStateReturnsPerContextConfidenceSummaries(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.provider = &applicationFakeProvider{
		id:          provider.CodexID,
		displayName: "Codex",
		statusByContext: map[string]provider.Status{
			"company":  provider.ConfiguredStatus(),
			"personal": provider.NotConfiguredStatus("Codex is not authenticated."),
		},
	}
	fixture.editor.executable = "/fixture/code"
	company := fixture.context("company", "Company")
	company.Providers = provider.Configs{
		provider.CodexID: {Enabled: true},
	}
	personal := fixture.context("personal", "Personal")
	personal.Providers = provider.Configs{
		provider.CodexID: {Enabled: true},
	}
	fixture.writeContext(t, company)
	fixture.writeContext(t, personal)
	companyPaths, err := filesystem.DeriveContextPaths(fixture.paths, company.ID)
	if err != nil {
		t.Fatalf("derive company paths: %v", err)
	}
	personalPaths, err := filesystem.DeriveContextPaths(fixture.paths, personal.ID)
	if err != nil {
		t.Fatalf("derive personal paths: %v", err)
	}
	removeAll(t, companyPaths.ToolStorageDir(company.Tool.DefaultTool))
	removeAll(t, personalPaths.ToolStorageDir(personal.Tool.DefaultTool))

	state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get launch state: %v", appErr)
	}
	if state.Confidence != nil {
		t.Fatalf("top-level confidence = %#v, want nil while selection is required", state.Confidence)
	}
	if len(state.Contexts) != 2 {
		t.Fatalf("context count = %d, want 2", len(state.Contexts))
	}

	confidenceByContext := map[string]LaunchConfidenceState{}
	for _, contextState := range state.Contexts {
		confidenceByContext[contextState.ID] = contextState.Confidence
	}

	if got := confidenceByContext["company"]; got.ContextID != "company" || got.Status != LaunchConfidenceBlocked {
		t.Fatalf("company confidence = %#v, want company blocked by missing selected tool storage", got)
	}
	if got := confidenceByContext["personal"]; got.ContextID != "personal" || got.Status != LaunchConfidenceBlocked {
		t.Fatalf("personal confidence = %#v, want personal blocked by missing selected tool storage", got)
	}
	assertConfidenceCheck(t, confidenceByContext["company"].Checks, LaunchConfidenceCheck{
		Component:  LaunchConfidenceCheckProvider,
		ProviderID: "codex",
		Severity:   LaunchConfidenceReady,
		Label:      "Codex",
		Message:    "Codex is ready for this context.",
	})
	assertConfidenceCheck(t, confidenceByContext["personal"].Checks, LaunchConfidenceCheck{
		Component:  LaunchConfidenceCheckProvider,
		ProviderID: "codex",
		Severity:   LaunchConfidenceNeedsAttention,
		Label:      "Codex",
		Message:    "Codex is not authenticated.",
		ActionHint: "Open and configure Codex for this context.",
	})
}

func TestGetLaunchStateTopLevelConfidenceMatchesSelectedContextSummary(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.provider = &applicationFakeProvider{
		id:          provider.CodexID,
		displayName: "Codex",
		statusByContext: map[string]provider.Status{
			"personal": provider.NotConfiguredStatus("Codex is not authenticated."),
		},
	}
	fixture.editor.executable = "/fixture/code"
	ctx := fixture.context("personal", "Personal")
	ctx.Providers = provider.Configs{
		provider.CodexID: {Enabled: true},
	}
	fixture.writeContext(t, ctx)
	fixture.writeBindings(t, project.Binding{
		ProjectPath: project.Path(fixture.projectDir),
		ContextID:   devcontext.MustID("personal"),
		CreatedAt:   fixture.now,
	})

	state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get launch state: %v", appErr)
	}
	if state.Confidence == nil {
		t.Fatal("top-level confidence = nil, want selected context confidence")
	}
	if !reflect.DeepEqual(*state.Confidence, state.Contexts[0].Confidence) {
		t.Fatalf("top-level confidence = %#v, want context confidence %#v", *state.Confidence, state.Contexts[0].Confidence)
	}
}

func TestPreflightLaunchProjectReturnsReadinessWithoutStartingProcess(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))

	result, appErr := fixture.service().PreflightLaunchProject(PreflightLaunchProjectRequest{
		ProjectPath: ".",
		ContextID:   "personal",
	})
	if appErr != nil {
		t.Fatalf("preflight launch project: %v", appErr)
	}

	if result.Project != (ProjectState{Name: "current", Path: fixture.projectDir}) {
		t.Fatalf("project = %#v, want current project", result.Project)
	}
	if result.Context.ID != "personal" || result.Confidence.ContextID != "personal" {
		t.Fatalf("preflight result = %#v, want personal context and confidence", result)
	}
	if !reflect.DeepEqual(result.Confidence, result.Context.Confidence) {
		t.Fatalf("preflight confidence = %#v, want context confidence %#v", result.Confidence, result.Context.Confidence)
	}
	if got, want := result.VerificationSteps, []LaunchVerificationStep{
		{ID: "prepare_environment", Label: "Prepare isolated environment", Status: LaunchVerificationStepReady, Message: "Prepare isolated environment is ready."},
		{ID: "check_providers", Label: "Check enabled providers", Status: LaunchVerificationStepReady, Message: "Check enabled providers is ready."},
		{ID: "prepare_tool", Label: "Prepare Fake Tool", Status: LaunchVerificationStepReady, Message: "Prepare Fake Tool is ready."},
		{ID: "start_tool", Label: "Start Fake Tool", Status: LaunchVerificationStepPending, Message: "Fake Tool will start after launch verification completes."},
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("verification steps = %#v, want %#v", got, want)
	}
	if len(fixture.process.requests) != 0 {
		t.Fatalf("process requests = %#v, want none for preflight", fixture.process.requests)
	}
}

func TestPreflightGroupsApplyProductGroupingAndBlockingPolicy(t *testing.T) {
	groups := preflightGroups(
		ContextState{
			Name: "Company",
			Tool: ToolState{ID: "fake", Name: "Fake Tool"},
			Confidence: LaunchConfidenceState{Checks: []LaunchConfidenceCheck{
				{Component: LaunchConfidenceCheckIdentity, Severity: LaunchConfidenceNeedsAttention, Label: "Account identity", Message: "Provider accounts differ."},
				{Component: LaunchConfidenceCheckIsolation, Severity: LaunchConfidenceReady, Label: "Context storage", Message: "Context storage is ready."},
				{Component: LaunchConfidenceCheckProvider, ProviderID: "codex", Severity: LaunchConfidenceNeedsAttention, Label: "Codex", Message: "Codex needs setup."},
				{Component: LaunchConfidenceCheckTool, ToolID: "fake", Severity: LaunchConfidenceBlocked, Label: "Fake Tool", Message: "Fake Tool is unavailable."},
			}},
		},
		[]ResolutionWarning{{Code: "context_mismatch", Message: "Selected context differs from the remembered context."}},
		&RunningEnvironmentConflict{Kind: "different_context"},
	)

	if len(groups) != 5 {
		t.Fatalf("group count = %d, want 5", len(groups))
	}
	byID := make(map[PreflightGroupID]PreflightGroup, len(groups))
	for _, group := range groups {
		byID[group.ID] = group
	}
	for _, id := range []PreflightGroupID{
		PreflightGroupProject,
		PreflightGroupContext,
		PreflightGroupIsolation,
		PreflightGroupTools,
		PreflightGroupWorkspace,
	} {
		if _, ok := byID[id]; !ok {
			t.Fatalf("missing %q group", id)
		}
	}
	if group := byID[PreflightGroupProject]; group.Status != LaunchConfidenceNeedsAttention || group.Blocking {
		t.Fatalf("project group = %#v, want non-blocking warning", group)
	}
	if group := byID[PreflightGroupContext]; group.Status != LaunchConfidenceNeedsAttention || group.Blocking {
		t.Fatalf("context group = %#v, want non-blocking identity warning", group)
	}
	if group := byID[PreflightGroupIsolation]; group.Status != LaunchConfidenceReady || group.Blocking {
		t.Fatalf("isolation group = %#v, want ready", group)
	}
	if group := byID[PreflightGroupTools]; group.Status != LaunchConfidenceBlocked || !group.Blocking || group.Message != "Fake Tool is unavailable." {
		t.Fatalf("tools group = %#v, want blocking tool failure", group)
	}
	if group := byID[PreflightGroupWorkspace]; group.Status != LaunchConfidenceNeedsAttention || group.Blocking {
		t.Fatalf("workspace group = %#v, want non-blocking conflict", group)
	}
}

func TestPreflightLaunchProjectRequiresMismatchConfirmation(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))
	fixture.writeContext(t, fixture.context("company", "Company"))
	fixture.writeBindings(t, project.Binding{
		ProjectPath: project.Path(fixture.projectDir),
		ContextID:   devcontext.MustID("company"),
		CreatedAt:   fixture.now,
	})

	_, appErr := fixture.service().PreflightLaunchProject(PreflightLaunchProjectRequest{
		ProjectPath: ".",
		ContextID:   "personal",
	})
	if appErr == nil {
		t.Fatal("preflight error = nil, want mismatch confirmation error")
	}
	if appErr.Code != ErrorCodeContextMismatch {
		t.Fatalf("error code = %q, want %q", appErr.Code, ErrorCodeContextMismatch)
	}
	if len(fixture.process.requests) != 0 {
		t.Fatalf("process requests = %#v, want none for preflight", fixture.process.requests)
	}
}

func TestGetLaunchStateReturnsUnboundSelectorState(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))
	fixture.writeContext(t, fixture.context("company", "Company"))

	state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get launch state: %v", appErr)
	}

	if state.Binding.Bound {
		t.Fatalf("binding = %#v, want unbound", state.Binding)
	}
	if !state.SelectionRequired {
		t.Fatal("selection required = false, want true")
	}
	if state.SelectedContextID != "" {
		t.Fatalf("selected context = %q, want empty", state.SelectedContextID)
	}
	if len(state.Contexts) != 2 {
		t.Fatalf("context count = %d, want 2", len(state.Contexts))
	}
	if state.Confidence != nil {
		t.Fatalf("confidence = %#v, want nil while user selection is required", state.Confidence)
	}
	if state.FirstRun {
		t.Fatal("first run = true, want false")
	}
}

func TestGetLaunchStateDetectsProviderCredentialSessionsForContextCreation(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.providerRegistry = provider.MustNewRegistry([]provider.Provider{
		fixture.provider,
		provider.ClaudeProvider{},
	}, fixture.provider.ID())
	fixture.writeContext(t, fixture.context("personal", "Personal"))
	writeApplicationJSONFixture(t, filepath.Join(fixture.homeDir, ".claude", ".credentials.json"), map[string]string{
		"subscriptionType": "Pro",
		"organizationUuid": "e783",
		"organizationName": "Jishin Labs",
		"accessToken":      "not-presented",
	})

	state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get launch state: %v", appErr)
	}

	if len(state.ProviderCredentialSessions) != 1 {
		t.Fatalf("provider credential sessions = %#v, want one Claude session", state.ProviderCredentialSessions)
	}
	session := state.ProviderCredentialSessions[0]
	if session.ProviderID != "claude" || session.Name != "Claude" || !session.MetadataAvailable {
		t.Fatalf("provider credential session = %#v, want available Claude session", session)
	}
	if metadataValueForTest(session.Fields, "Subscription") != "Pro" ||
		metadataValueForTest(session.Fields, "Organization UUID") != "e783" ||
		metadataValueForTest(session.Fields, "Organization") != "Jishin Labs" {
		t.Fatalf("claude metadata = %#v", session.Fields)
	}
	if rendered := fmt.Sprintf("%#v", state.ProviderCredentialSessions); strings.Contains(rendered, "not-presented") {
		t.Fatalf("provider credential sessions exposed credential value: %#v", state.ProviderCredentialSessions)
	}
}

func TestGetLaunchStateIncludesRegisteredProviderCredentialMetadata(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.provider = &applicationFakeProvider{
		id:               "future",
		displayName:      "Future Provider",
		hasGlobalSession: true,
		globalSession: provider.CredentialSession{
			MetadataAvailable: true,
			Fields:            []provider.MetadataField{{Label: "Workspace", Value: "Example"}},
		},
	}
	fixture.writeContext(t, fixture.context("personal", "Personal"))

	state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get launch state: %v", appErr)
	}
	if len(state.ProviderCredentialSessions) != 1 {
		t.Fatalf("provider credential sessions = %#v", state.ProviderCredentialSessions)
	}
	session := state.ProviderCredentialSessions[0]
	if session.ProviderID != "future" || session.Name != "Future Provider" || metadataValueForTest(session.Fields, "Workspace") != "Example" {
		t.Fatalf("provider credential session = %#v", session)
	}
}

func TestGetLaunchStateIncludesRegisteredProviderIdentityMetadata(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.provider = &applicationFakeProvider{
		id:          "future",
		displayName: "Future Provider",
		statusByContext: map[string]provider.Status{
			"personal": provider.ConfiguredStatus(),
		},
		hasIdentity: true,
		identity:    provider.Identity{Fields: []provider.MetadataField{{Label: "Workspace", Value: "Example"}}},
	}
	ctx := fixture.context("personal", "Personal")
	ctx.Providers = provider.Configs{"future": {Enabled: true}}
	fixture.writeContext(t, ctx)

	state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get launch state: %v", appErr)
	}
	identity := state.Contexts[0].Providers[0].Identity
	if identity.Status != ProviderIdentityVerified || metadataValueForTest(identity.Fields, "Workspace") != "Example" {
		t.Fatalf("provider identity = %#v", identity)
	}
}

func TestGetLaunchStateReportsMissingProviderStatus(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.provider.statusByContext = map[string]provider.Status{
		"personal": provider.UnavailableStatus("Fake Provider command was not found"),
	}
	fixture.writeContext(t, fixture.context("personal", "Personal"))

	state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get launch state: %v", appErr)
	}

	got := state.Contexts[0].Providers[0]
	want := ProviderState{
		ID:          "fake",
		Name:        "Fake Provider",
		Enabled:     true,
		State:       ProviderReadinessUnavailable,
		Explanation: "Fake Provider command was not found",
		Identity: ProviderIdentityState{
			Status:  ProviderIdentityUnavailable,
			Message: "Account identity unavailable.",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("provider status = %#v, want %#v", got, want)
	}
}

func TestGetLaunchStateNormalizesProviderReadinessForUI(t *testing.T) {
	tests := []struct {
		name   string
		status provider.Status
		want   ProviderReadinessState
	}{
		{
			name:   "configured maps to ready",
			status: provider.ConfiguredStatus(),
			want:   ProviderReadinessReady,
		},
		{
			name:   "not configured",
			status: provider.NotConfiguredStatus("missing"),
			want:   ProviderReadinessNotConfigured,
		},
		{
			name:   "directory missing",
			status: provider.DirectoryMissingStatus("missing directory"),
			want:   ProviderReadinessDirectoryMissing,
		},
		{
			name:   "unavailable",
			status: provider.UnavailableStatus("unavailable"),
			want:   ProviderReadinessUnavailable,
		},
		{
			name:   "unknown falls back to unavailable",
			status: provider.Status{State: "configured_but_unknown"},
			want:   ProviderReadinessUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newApplicationFixture(t)
			fixture.provider.statusByContext = map[string]provider.Status{
				"personal": tt.status,
			}
			fixture.writeContext(t, fixture.context("personal", "Personal"))

			state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
			if appErr != nil {
				t.Fatalf("get launch state: %v", appErr)
			}

			if got := state.Contexts[0].Providers[0].State; got != tt.want {
				t.Fatalf("provider state = %q, want %q", got, tt.want)
			}
			if !state.Contexts[0].Providers[0].State.Valid() {
				t.Fatalf("provider state is invalid: %q", state.Contexts[0].Providers[0].State)
			}
		})
	}
}

func TestGetLaunchStateReturnsProviderIdentityContract(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		status  provider.Status
		want    ProviderIdentityState
	}{
		{
			name:    "configured provider identity unavailable until verified",
			enabled: true,
			status:  provider.ConfiguredStatus(),
			want: ProviderIdentityState{
				Status:  ProviderIdentityUnavailable,
				Message: "Account identity unavailable.",
			},
		},
		{
			name:    "not configured provider has no identity",
			enabled: true,
			status:  provider.NotConfiguredStatus("missing"),
			want:    ProviderIdentityState{Status: ProviderIdentityNone},
		},
		{
			name:    "disabled provider has no identity",
			enabled: false,
			status:  provider.ConfiguredStatus(),
			want:    ProviderIdentityState{Status: ProviderIdentityNone},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newApplicationFixture(t)
			fixture.provider.statusByContext = map[string]provider.Status{
				"personal": tt.status,
			}
			ctx := fixture.context("personal", "Personal")
			ctx.Providers["fake"] = provider.Config{Enabled: tt.enabled}
			fixture.writeContext(t, ctx)

			state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
			if appErr != nil {
				t.Fatalf("get launch state: %v", appErr)
			}

			if got := state.Contexts[0].Providers[0].Identity; !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("provider identity = %#v, want %#v", got, tt.want)
			}
			if !state.Contexts[0].Providers[0].Identity.Status.Valid() {
				t.Fatalf("provider identity status is invalid: %q", state.Contexts[0].Providers[0].Identity.Status)
			}
		})
	}
}

func TestGetLaunchStateReturnsVerifiedProviderIdentitiesForIsolatedContexts(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.providerRegistry = provider.MustNewRegistry([]provider.Provider{
		applicationFakeProvider{
			id:          provider.CodexID,
			displayName: "Codex",
			statusByContext: map[string]provider.Status{
				"personal": provider.ConfiguredStatus(),
			},
		},
		applicationFakeProvider{
			id:          provider.ClaudeID,
			displayName: "Claude",
			statusByContext: map[string]provider.Status{
				"personal": provider.ConfiguredStatus(),
			},
		},
	})

	ctx := fixture.context("personal", "Personal")
	ctx.Providers = provider.Configs{
		provider.CodexID:  {Enabled: true},
		provider.ClaudeID: {Enabled: true},
	}
	fixture.writeContext(t, ctx)

	contextPaths, err := filesystem.DeriveContextPaths(fixture.paths, ctx.ID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	codexToken := applicationTestJWT(t, map[string]string{
		"email":               "user@company.com",
		"chatgpt_plan_type":   "business",
		"chatgpt_account_id":  "acct-123",
		"ignored_token_claim": "jwt-secret-claim",
	})
	writeApplicationJSONFixture(t, filepath.Join(contextPaths.ProviderStorageDir(provider.CodexID), "auth.json"), map[string]any{
		"tokens": map[string]string{
			"id_token":     codexToken,
			"access_token": "codex-access-token",
		},
	})
	writeApplicationJSONFixture(t, filepath.Join(contextPaths.ProviderStorageDir(provider.ClaudeID), ".credentials.json"), map[string]string{
		"subscriptionType": "Pro",
		"organizationUuid": "e783-organization",
		"organizationName": "Jishin Labs",
		"accessToken":      "claude-access-token",
		"refreshToken":     "claude-refresh-token",
	})

	state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get launch state: %v", appErr)
	}

	providersByID := map[string]ProviderState{}
	for _, providerState := range state.Contexts[0].Providers {
		providersByID[providerState.ID] = providerState
	}

	codex := providersByID["codex"].Identity
	if codex.Status != ProviderIdentityVerified {
		t.Fatalf("codex identity = %#v, want verified codex identity", codex)
	}
	if metadataValueForTest(codex.Fields, "Email") != "user@company.com" ||
		metadataValueForTest(codex.Fields, "ChatGPT plan") != "business" ||
		metadataValueForTest(codex.Fields, "ChatGPT account ID") != "acct-123" {
		t.Fatalf("codex identity metadata = %#v", codex.Fields)
	}

	claude := providersByID["claude"].Identity
	if claude.Status != ProviderIdentityVerified {
		t.Fatalf("claude identity = %#v, want verified claude identity", claude)
	}
	if metadataValueForTest(claude.Fields, "Subscription") != "Pro" ||
		metadataValueForTest(claude.Fields, "Organization UUID") != "e783-organization" ||
		metadataValueForTest(claude.Fields, "Organization") != "Jishin Labs" {
		t.Fatalf("claude identity metadata = %#v", claude.Fields)
	}

	rendered := fmt.Sprintf("%#v", state.Contexts[0].Providers)
	for _, secret := range []string{
		codexToken,
		"codex-access-token",
		"jwt-secret-claim",
		"claude-access-token",
		"claude-refresh-token",
	} {
		if strings.Contains(rendered, secret) {
			t.Fatalf("provider state exposed credential value %q: %s", secret, rendered)
		}
	}
}

func TestGetLaunchStateDoesNotInferIdentityMismatchEvidenceFromContextName(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.providerRegistry = provider.MustNewRegistry([]provider.Provider{
		applicationFakeProvider{
			id:          provider.CodexID,
			displayName: "Codex",
			statusByContext: map[string]provider.Status{
				"company": provider.ConfiguredStatus(),
			},
		},
	})

	ctx := fixture.context("company", "Company")
	ctx.Providers = provider.Configs{
		provider.CodexID: {Enabled: true},
	}
	fixture.writeContext(t, ctx)

	contextPaths, err := filesystem.DeriveContextPaths(fixture.paths, ctx.ID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	writeApplicationJSONFixture(t, filepath.Join(contextPaths.ProviderStorageDir(provider.CodexID), "auth.json"), map[string]string{
		"id_token": applicationTestJWT(t, map[string]string{
			"email":              "user@gmail.com",
			"chatgpt_plan_type":  "plus",
			"chatgpt_account_id": "acct-personal",
		}),
	})

	state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get launch state: %v", appErr)
	}

	identity := state.Contexts[0].Providers[0].Identity
	if identity.Status != ProviderIdentityVerified {
		t.Fatalf("identity status = %q, want verified without inferred mismatch evidence", identity.Status)
	}
}

func TestGetLaunchStateReportsOnlyMeaningfulProviderEmailMismatchEvidence(t *testing.T) {
	fixture := newApplicationFixture(t)
	first := applicationFakeProvider{
		id: "first", displayName: "First", hasIdentity: true,
		identity:        provider.Identity{Fields: []provider.MetadataField{{Label: "Email", Value: "personal@example.com"}}},
		statusByContext: map[string]provider.Status{"company": provider.ConfiguredStatus()},
	}
	second := applicationFakeProvider{
		id: "second", displayName: "Second", hasIdentity: true,
		identity:        provider.Identity{Fields: []provider.MetadataField{{Label: "Email", Value: "work@example.com"}}},
		statusByContext: map[string]provider.Status{"company": provider.ConfiguredStatus()},
	}
	fixture.providerRegistry = provider.MustNewRegistry([]provider.Provider{first, second}, first.ID())
	ctx := fixture.context("company", "Company")
	ctx.Providers = provider.Configs{first.ID(): {Enabled: true}, second.ID(): {Enabled: true}}
	fixture.writeContext(t, ctx)

	state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get launch state: %v", appErr)
	}
	var mismatch *LaunchConfidenceCheck
	for index := range state.Contexts[0].Confidence.Checks {
		check := &state.Contexts[0].Confidence.Checks[index]
		if check.Component == LaunchConfidenceCheckIdentity {
			mismatch = check
		}
	}
	if mismatch == nil || mismatch.Severity != LaunchConfidenceNeedsAttention {
		t.Fatalf("mismatch check = %#v, want needs-attention identity check", mismatch)
	}
	if state.Contexts[0].Confidence.Status != LaunchConfidenceNeedsAttention {
		t.Fatalf("confidence status = %q, want needs_attention", state.Contexts[0].Confidence.Status)
	}
}

func TestGetLaunchStateReportsDanglingBindingWarning(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))
	fixture.writeBindings(t, project.Binding{
		ProjectPath: project.Path(fixture.projectDir),
		ContextID:   devcontext.MustID("company"),
		CreatedAt:   fixture.now,
	})

	state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get launch state: %v", appErr)
	}

	if !state.Binding.Dangling || state.Binding.MissingContextID != "company" {
		t.Fatalf("binding = %#v, want dangling company binding", state.Binding)
	}
	if len(state.Warnings) != 1 || state.Warnings[0].Code != string(launcher.WarningDanglingProjectBinding) {
		t.Fatalf("warnings = %#v, want dangling binding warning", state.Warnings)
	}
}

func TestGetLaunchStateDetectsFirstRunState(t *testing.T) {
	t.Run("absent initialization", func(t *testing.T) {
		fixture := newApplicationFixture(t)
		removeAll(t, fixture.contextsDir)

		state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
		if appErr != nil {
			t.Fatalf("get launch state: %v", appErr)
		}

		assertFirstRunState(t, state, fixture.projectDir)
	})

	t.Run("initialized empty home", func(t *testing.T) {
		fixture := newApplicationFixture(t)

		state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
		if appErr != nil {
			t.Fatalf("get launch state: %v", appErr)
		}

		assertFirstRunState(t, state, fixture.projectDir)
	})

	t.Run("populated home", func(t *testing.T) {
		fixture := newApplicationFixture(t)
		fixture.writeContext(t, fixture.context("personal", "Personal"))

		state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
		if appErr != nil {
			t.Fatalf("get launch state: %v", appErr)
		}

		if state.FirstRun {
			t.Fatal("first run = true, want false")
		}
		if len(state.Contexts) != 1 {
			t.Fatalf("context count = %d, want 1", len(state.Contexts))
		}
	})
}

func TestGetLaunchStateReportsContextStorageErrors(t *testing.T) {
	fixture := newApplicationFixture(t)
	removeAll(t, fixture.contextsDir)
	writeFile(t, fixture.contextsDir, []byte("not a directory"))

	_, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
	if appErr == nil {
		t.Fatal("get launch state error = nil, want context storage error")
	}
	if appErr.Code != ErrorCodeInternal {
		t.Fatalf("error code = %q, want %q", appErr.Code, ErrorCodeInternal)
	}
}

func TestCreateContextCreatesDefaultPersonalAndCompanyContexts(t *testing.T) {
	defaultRegistry := provider.MustNewRegistry([]provider.Provider{
		applicationFakeProvider{id: "fake"},
	}, "fake")
	now := time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC)
	tests := []struct {
		name      string
		contextID string
		want      devcontext.Context
	}{
		{
			name:      "personal",
			contextID: "personal",
			want:      devcontext.Context{ID: devcontext.MustID("personal"), Name: "Personal", Tool: codingtool.LaunchTarget{DefaultTool: "fake-editor", Tools: map[codingtool.ID]codingtool.Config{"fake-editor": {}}}, Providers: defaultRegistry.DefaultConfigs(), CreatedAt: now},
		},
		{
			name:      "company",
			contextID: "company",
			want:      devcontext.Context{ID: devcontext.MustID("company"), Name: "Company", Tool: codingtool.LaunchTarget{DefaultTool: "fake-editor", Tools: map[codingtool.ID]codingtool.Config{"fake-editor": {}}}, Providers: defaultRegistry.DefaultConfigs(), CreatedAt: now},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newApplicationFixture(t)

			result, appErr := fixture.service().CreateContext(CreateContextRequest{ContextID: tt.contextID})
			if appErr != nil {
				t.Fatalf("create context: %v", appErr)
			}

			if result.Context.ID != tt.want.ID.String() || result.Context.Name != tt.want.Name {
				t.Fatalf("context state = %#v, want %s %s", result.Context, tt.want.ID.String(), tt.want.Name)
			}

			stored, err := devcontext.NewRepository(fixture.contextsDir).Get(tt.want.ID)
			if err != nil {
				t.Fatalf("get stored context: %v", err)
			}
			if !reflect.DeepEqual(stored, tt.want) {
				t.Fatalf("stored context = %#v, want %#v", stored, tt.want)
			}

			state, stateErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
			if stateErr != nil {
				t.Fatalf("get launch state: %v", stateErr)
			}
			if state.FirstRun {
				t.Fatal("first run = true, want false")
			}
			if len(state.Contexts) != 1 || state.Contexts[0].ID != tt.contextID {
				t.Fatalf("contexts = %#v, want created context", state.Contexts)
			}
		})
	}
}

func TestCreateContextReportsDuplicateDefaultContext(t *testing.T) {
	fixture := newApplicationFixture(t)
	if _, appErr := fixture.service().CreateContext(CreateContextRequest{ContextID: "personal"}); appErr != nil {
		t.Fatalf("create original context: %v", appErr)
	}

	_, appErr := fixture.service().CreateContext(CreateContextRequest{ContextID: "personal"})
	if appErr == nil {
		t.Fatal("create duplicate error = nil, want validation error")
	}
	if appErr.Code != ErrorCodeValidation {
		t.Fatalf("error code = %q, want %q", appErr.Code, ErrorCodeValidation)
	}
}

func TestCreateContextGeneratesUniqueIDFromName(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("my-work", "Existing work"))

	result, appErr := fixture.service().CreateContext(CreateContextRequest{Name: "My Work"})
	if appErr != nil {
		t.Fatalf("create context: %v", appErr)
	}
	if result.Context.ID != "my-work-2" {
		t.Fatalf("generated context ID = %q, want my-work-2", result.Context.ID)
	}
	if result.Context.Name != "My Work" {
		t.Fatalf("context name = %q, want My Work", result.Context.Name)
	}
}

func TestCreateContextRecordsHistoryEvent(t *testing.T) {
	fixture := newApplicationFixture(t)
	logger := &applicationRecordingLogger{}
	fixture.logger = logger

	if _, appErr := fixture.service().CreateContext(CreateContextRequest{ContextID: "personal"}); appErr != nil {
		t.Fatalf("create context: %v", appErr)
	}

	if got := applicationEventNames(logger.events); !reflect.DeepEqual(got, []devlog.EventName{devlog.EventContextCreated}) {
		t.Fatalf("event names = %#v, want context-created event", got)
	}
	if event := logger.events[0]; event.ContextID != "personal" || event.ToolID != "fake-editor" {
		t.Fatalf("context-created event = %#v", event)
	}
}

func TestCreateContextReportsPermissionFailure(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.storagePermissions = filesystem.NewStoragePermissions(true, func(string, os.FileMode) error {
		return os.ErrPermission
	})

	_, appErr := fixture.service().CreateContext(CreateContextRequest{ContextID: "personal"})
	if appErr == nil {
		t.Fatal("create context error = nil, want permission error")
	}
	if appErr.Code != ErrorCodeValidation {
		t.Fatalf("error code = %q, want %q", appErr.Code, ErrorCodeValidation)
	}
}

func TestCreateContextReportsStorageWriteFailure(t *testing.T) {
	fixture := newApplicationFixture(t)
	removeAll(t, fixture.contextsDir)
	writeFile(t, fixture.contextsDir, []byte("not a directory"))

	_, appErr := fixture.service().CreateContext(CreateContextRequest{ContextID: "personal"})
	if appErr == nil {
		t.Fatal("create context error = nil, want write failure")
	}
	if appErr.Code != ErrorCodeInternal {
		t.Fatalf("error code = %q, want %q", appErr.Code, ErrorCodeInternal)
	}
}

func TestGetContextTemplatesReturnsBuiltInSafeDefaults(t *testing.T) {
	templates := newApplicationFixture(t).service().GetContextTemplates().Templates
	if got, want := len(templates), 6; got != want {
		t.Fatalf("template count = %d, want %d", got, want)
	}
	if templates[0].ID != "personal" || templates[5].ID != "custom" {
		t.Fatalf("templates = %#v", templates)
	}
}

func TestCreateContextAppliesTemplateDefaults(t *testing.T) {
	fixture := newApplicationFixture(t)
	_, appErr := fixture.service().CreateContext(CreateContextRequest{ContextID: "freelance", TemplateID: "freelance"})
	if appErr != nil {
		t.Fatalf("create context: %v", appErr)
	}
	stored, err := devcontext.NewRepository(fixture.contextsDir).Get(devcontext.MustID("freelance"))
	if err != nil {
		t.Fatalf("get stored context: %v", err)
	}
	if stored.Name != "Freelance" || stored.Metadata["accent"] != "amber" {
		t.Fatalf("stored context = %#v, want template defaults", stored)
	}
}

func TestDuplicateContextCopiesSafeConfigurationWithoutCredentials(t *testing.T) {
	fixture := newApplicationFixture(t)
	source := fixture.context("personal", "Personal")
	source.Tool.Tools[source.Tool.DefaultTool] = codingtool.Config{ExecutableOverride: "/custom/tool", Options: map[string]string{"profile": "personal"}}
	source.Providers["fake"] = provider.Config{Enabled: true, Options: map[string]string{"region": "south"}}
	fixture.writeContext(t, source)
	sourcePaths, err := filesystem.DeriveContextPaths(fixture.paths, source.ID)
	if err != nil {
		t.Fatalf("derive source paths: %v", err)
	}
	writeFile(t, filepath.Join(sourcePaths.ProviderStorageDir("fake"), "credential.json"), []byte(`{"token":"secret"}`))

	result, appErr := fixture.service().DuplicateContext(DuplicateContextRequest{SourceContextID: "personal", ContextID: "personal-copy"})
	if appErr != nil {
		t.Fatalf("duplicate context: %v", appErr)
	}
	if result.Context.Name != "Personal copy" {
		t.Fatalf("name = %q, want Personal copy", result.Context.Name)
	}
	target, err := devcontext.NewRepository(fixture.contextsDir).Get(devcontext.MustID("personal-copy"))
	if err != nil {
		t.Fatalf("get duplicate: %v", err)
	}
	if !reflect.DeepEqual(target.Tool, source.Tool) || !reflect.DeepEqual(target.Providers, source.Providers) || !reflect.DeepEqual(target.Metadata, source.Metadata) {
		t.Fatalf("duplicate safe configuration = %#v, want source %#v", target, source)
	}
	targetPaths, err := filesystem.DeriveContextPaths(fixture.paths, target.ID)
	if err != nil {
		t.Fatalf("derive target paths: %v", err)
	}
	if _, err := os.Stat(filepath.Join(targetPaths.ProviderStorageDir("fake"), "credential.json")); !os.IsNotExist(err) {
		t.Fatalf("duplicate credential file error = %v, want not exist", err)
	}
}

func TestExportContextMetadataIncludesOnlyPortableSafeConfiguration(t *testing.T) {
	fixture := newApplicationFixture(t)
	source := fixture.context("personal", "Personal")
	source.Metadata = devcontext.Metadata{"accent": "sage"}
	source.Tool.Tools[source.Tool.DefaultTool] = codingtool.Config{ExecutableOverride: "/private/bin/code", Options: map[string]string{"profile": "personal"}}
	source.Providers["fake"] = provider.Config{Enabled: true, Options: map[string]string{"region": "south"}}
	fixture.writeContext(t, source)
	sourcePaths, err := filesystem.DeriveContextPaths(fixture.paths, source.ID)
	if err != nil {
		t.Fatalf("derive source paths: %v", err)
	}
	writeFile(t, filepath.Join(sourcePaths.ProviderStorageDir("fake"), "credential.json"), []byte(`{"token":"secret"}`))

	exported, appErr := fixture.service().ExportContextMetadata(ExportContextMetadataRequest{ContextID: "personal"})
	if appErr != nil {
		t.Fatalf("export context metadata: %v", appErr)
	}
	want := ContextMetadataExport{
		Version: ContextTransferVersion,
		Context: ContextTransferMetadata{
			Name: "Personal", Metadata: map[string]string{"accent": "sage"},
			Providers:    []ContextTransferProvider{{ID: "fake", Enabled: true, Options: map[string]string{"region": "south"}}},
			LaunchTarget: ContextTransferLaunchTarget{DefaultTool: "fake-editor", Tools: []ContextTransferTool{{ID: "fake-editor", Options: map[string]string{"profile": "personal"}}}},
		},
	}
	if !reflect.DeepEqual(exported, want) {
		t.Fatalf("export = %#v, want %#v", exported, want)
	}
}

func TestImportContextMetadataCreatesFreshStorageWithoutCredentials(t *testing.T) {
	fixture := newApplicationFixture(t)
	exported := ContextMetadataExport{
		Version: ContextTransferVersion,
		Context: ContextTransferMetadata{
			Name: "Imported Personal", Metadata: map[string]string{"accent": "sage"},
			Providers:    []ContextTransferProvider{{ID: "fake", Enabled: true, Options: map[string]string{"region": "south"}}},
			LaunchTarget: ContextTransferLaunchTarget{DefaultTool: "fake-editor", Tools: []ContextTransferTool{{ID: "fake-editor", Options: map[string]string{"profile": "personal"}}}},
		},
	}

	result, appErr := fixture.service().ImportContextMetadata(ImportContextMetadataRequest{ContextID: "imported", Export: exported})
	if appErr != nil {
		t.Fatalf("import context metadata: %v", appErr)
	}
	if result.Context.ID != "imported" || result.Context.Name != "Imported Personal" {
		t.Fatalf("import result = %#v", result)
	}
	stored, err := devcontext.NewRepository(fixture.contextsDir).Get(devcontext.MustID("imported"))
	if err != nil {
		t.Fatalf("get imported context: %v", err)
	}
	if stored.CreatedAt.IsZero() || stored.Tool.Tools["fake-editor"].ExecutableOverride != "" {
		t.Fatalf("imported context = %#v, want fresh creation and no executable override", stored)
	}
	paths, err := filesystem.DeriveContextPaths(fixture.paths, stored.ID)
	if err != nil {
		t.Fatalf("derive imported paths: %v", err)
	}
	if _, err := os.Stat(filepath.Join(paths.ProviderStorageDir("fake"), "credential.json")); !os.IsNotExist(err) {
		t.Fatalf("imported credential file error = %v, want not exist", err)
	}
}

func TestImportContextMetadataRejectsUnsupportedVersionAndUnknownIntegration(t *testing.T) {
	fixture := newApplicationFixture(t)
	base := ContextMetadataExport{Version: ContextTransferVersion, Context: ContextTransferMetadata{Name: "Imported", LaunchTarget: ContextTransferLaunchTarget{DefaultTool: "fake-editor"}}}
	for _, testCase := range []struct {
		name   string
		export ContextMetadataExport
	}{
		{name: "version", export: ContextMetadataExport{Version: ContextTransferVersion + 1, Context: base.Context}},
		{name: "tool", export: ContextMetadataExport{Version: ContextTransferVersion, Context: ContextTransferMetadata{Name: "Imported", LaunchTarget: ContextTransferLaunchTarget{DefaultTool: "unknown"}}}},
		{name: "provider", export: ContextMetadataExport{Version: ContextTransferVersion, Context: ContextTransferMetadata{Name: "Imported", Providers: []ContextTransferProvider{{ID: "unknown", Enabled: true}}, LaunchTarget: ContextTransferLaunchTarget{DefaultTool: "fake-editor"}}}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, appErr := fixture.service().ImportContextMetadata(ImportContextMetadataRequest{ContextID: "imported-" + testCase.name, Export: testCase.export})
			if appErr == nil || appErr.Code != ErrorCodeValidation {
				t.Fatalf("import error = %#v, want validation error", appErr)
			}
		})
	}
}

func TestLaunchProjectBuildsPlanAndStartsProcess(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))

	result, appErr := fixture.service().LaunchProject(LaunchProjectRequest{
		ProjectPath: fixture.projectDir,
		ContextID:   "personal",
	})
	if appErr != nil {
		t.Fatalf("launch project: %v", appErr)
	}

	if result.Project.Path != fixture.projectDir || result.Context.ID != "personal" {
		t.Fatalf("launch result = %#v, want personal current project", result)
	}
	if len(fixture.process.requests) != 1 {
		t.Fatalf("process request count = %d, want 1", len(fixture.process.requests))
	}
	request := fixture.process.requests[0]
	if request.DetachMode != launcher.DetachModeAttached {
		t.Fatalf("detach mode = %q, want %q", request.DetachMode, launcher.DetachModeAttached)
	}
	if request.Environment["DEVCTX_CONTEXT"] != "personal" {
		t.Fatalf("DEVCTX_CONTEXT = %q, want personal", request.Environment["DEVCTX_CONTEXT"])
	}
	if request.Environment["FAKE_CONTEXT"] != "personal" {
		t.Fatalf("FAKE_CONTEXT = %q, want personal", request.Environment["FAKE_CONTEXT"])
	}
	if !reflect.DeepEqual(request.Arguments[len(request.Arguments)-1:], launcher.Arguments{fixture.projectDir}) {
		t.Fatalf("arguments = %#v, want project path as final argument", request.Arguments)
	}
}

func TestLaunchProjectExportsSafeStatusForStatusAwareTool(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.provider = &applicationFakeProvider{
		id:          "fake",
		displayName: "Fake Provider",
		hasIdentity: true,
		identity: provider.Identity{Fields: []provider.MetadataField{
			{Label: "Account", Value: "developer@example.com"},
		}},
	}
	statusTool := applicationStatusDataEditor{applicationFakeEditor: fixture.editor, fileName: "integration-status.json"}
	fixture.toolRegistry = codingtool.MustNewRegistry([]codingtool.RegisteredTool{{Integration: statusTool, DisplayName: "Status Tool"}}, fixture.editor.ID())
	fixture.writeContext(t, fixture.context("personal", "Personal"))

	if _, appErr := fixture.service().LaunchProject(LaunchProjectRequest{ProjectPath: fixture.projectDir, ContextID: "personal"}); appErr != nil {
		t.Fatalf("launch project: %v", appErr)
	}

	statusPath := filepath.Join(fixture.contextsDir, "personal", "tools", string(fixture.editor.ID()), "integration-status.json")
	data, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("read status data: %v", err)
	}
	var status CodingToolStatusData
	if err := json.Unmarshal(data, &status); err != nil {
		t.Fatalf("decode status data: %v", err)
	}
	if status.SchemaVersion != codingToolStatusSchemaVersion || status.Project != (ProjectState{Name: "current", Path: fixture.projectDir}) {
		t.Fatalf("status header = %#v", status)
	}
	if status.Context.ID != "personal" || status.Context.Name != "Personal" || status.Context.Tool != (ToolOption{ID: "fake-editor", Name: "Status Tool"}) {
		t.Fatalf("status context = %#v", status.Context)
	}
	if !reflect.DeepEqual(status.Providers, []CodingToolStatusProvider{{
		ID: "fake", Name: "Fake Provider", Identity: ProviderIdentityState{
			Status: ProviderIdentityVerified,
			Fields: []ProviderMetadataField{{Label: "Account", Value: "developer@example.com"}},
		},
	}}) {
		t.Fatalf("status providers = %#v", status.Providers)
	}
	if status.Isolation.Status != LaunchConfidenceReady || status.Isolation.Message != "Context isolation is ready." {
		t.Fatalf("status isolation = %#v", status.Isolation)
	}
	if strings.Contains(string(data), "FAKE_CONTEXT") || strings.Contains(string(data), fixture.contextsDir) {
		t.Fatalf("status data contains private runtime information: %s", data)
	}
	if len(fixture.process.requests) != 1 {
		t.Fatalf("process requests = %#v, want launch after export", fixture.process.requests)
	}
}

func TestLaunchProjectRecordsLifecycleEvents(t *testing.T) {
	fixture := newApplicationFixture(t)
	logger := &applicationRecordingLogger{}
	fixture.logger = logger
	fixture.writeContext(t, fixture.context("personal", "Personal"))

	_, appErr := fixture.service().LaunchProject(LaunchProjectRequest{
		ProjectPath: fixture.projectDir,
		ContextID:   "personal",
	})
	if appErr != nil {
		t.Fatalf("launch project: %v", appErr)
	}

	wantNames := []devlog.EventName{
		devlog.EventContextResolution,
		devlog.EventLaunchSucceeded,
	}
	if got := applicationEventNames(logger.events); !reflect.DeepEqual(got, wantNames) {
		t.Fatalf("event names = %#v, want %#v", got, wantNames)
	}
	for _, event := range logger.events {
		if event.ProjectPath != fixture.projectDir {
			t.Fatalf("event project path = %q, want %q", event.ProjectPath, fixture.projectDir)
		}
		if event.ContextID != "personal" {
			t.Fatalf("event context ID = %q, want personal", event.ContextID)
		}
	}
}

func TestLaunchProjectCreatesAndUpdatesRunningEnvironment(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))

	launch := func(contextID string) {
		t.Helper()
		if _, appErr := fixture.service().LaunchProject(LaunchProjectRequest{ProjectPath: fixture.projectDir, ContextID: contextID}); appErr != nil {
			t.Fatalf("launch project: %v", appErr)
		}
	}

	launch("personal")
	repository := coreRunning.NewRepository(fixture.runningPath)
	environments, err := repository.List()
	if err != nil {
		t.Fatalf("list running environments: %v", err)
	}
	if len(environments) != 1 {
		t.Fatalf("running environments = %#v, want one", environments)
	}
	first := environments[0]
	if first.ID == "" || first.Project.Path != project.Path(fixture.projectDir) || first.Context.ID.String() != "personal" || first.Tool.Name != "Fake Tool" {
		t.Fatalf("running environment = %#v", first)
	}
	if first.Process.State != coreRunning.ProcessStateRunning || first.Session.State != coreRunning.SessionStateUnknown {
		t.Fatalf("running environment state = %#v", first)
	}

	fixture.now = fixture.now.Add(time.Minute)
	launch("personal")
	environments, err = repository.List()
	if err != nil {
		t.Fatalf("list updated running environments: %v", err)
	}
	if len(environments) != 1 || environments[0].ID != first.ID || !environments[0].StartedAt.Equal(fixture.now) {
		t.Fatalf("updated running environments = %#v", environments)
	}

	fixture.writeContext(t, fixture.context("company", "Company"))
	launch("company")
	environments, err = repository.List()
	if err != nil {
		t.Fatalf("list context-specific running environments: %v", err)
	}
	if len(environments) != 2 {
		t.Fatalf("running environments = %#v, want separate records by context", environments)
	}
}

func TestLaunchProjectRequiresMismatchConfirmation(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))
	fixture.writeContext(t, fixture.context("company", "Company"))
	fixture.writeBindings(t, project.Binding{
		ProjectPath: project.Path(fixture.projectDir),
		ContextID:   devcontext.MustID("company"),
		CreatedAt:   fixture.now,
	})

	_, appErr := fixture.service().LaunchProject(LaunchProjectRequest{
		ProjectPath: fixture.projectDir,
		ContextID:   "personal",
	})
	if appErr == nil {
		t.Fatal("launch error = nil, want mismatch confirmation error")
	}
	if appErr.Code != ErrorCodeContextMismatch {
		t.Fatalf("error code = %q, want %q", appErr.Code, ErrorCodeContextMismatch)
	}
	if appErr.ContextMismatch == nil ||
		appErr.ContextMismatch.BoundContextID != "company" ||
		appErr.ContextMismatch.RequestedContextID != "personal" {
		t.Fatalf("context mismatch = %#v, want company/personal details", appErr.ContextMismatch)
	}
	if len(fixture.process.requests) != 0 {
		t.Fatalf("process requests = %#v, want none", fixture.process.requests)
	}
}

func TestLaunchProjectAcceptsConfirmedMismatch(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))
	fixture.writeContext(t, fixture.context("company", "Company"))
	fixture.writeBindings(t, project.Binding{
		ProjectPath: project.Path(fixture.projectDir),
		ContextID:   devcontext.MustID("company"),
		CreatedAt:   fixture.now,
	})

	result, appErr := fixture.service().LaunchProject(LaunchProjectRequest{
		ProjectPath:            fixture.projectDir,
		ContextID:              "personal",
		ConfirmContextMismatch: true,
	})
	if appErr != nil {
		t.Fatalf("launch project: %v", appErr)
	}
	if len(result.Warnings) != 1 || result.Warnings[0].Code != string(launcher.WarningContextMismatch) {
		t.Fatalf("warnings = %#v, want mismatch warning", result.Warnings)
	}
	if len(fixture.process.requests) != 1 {
		t.Fatalf("process request count = %d, want 1", len(fixture.process.requests))
	}

	bindings, err := project.ReadProjectBindingsFile(fixture.bindingsPath)
	if err != nil {
		t.Fatalf("read project bindings: %v", err)
	}
	want := []project.Binding{{
		ProjectPath: project.Path(fixture.projectDir),
		ContextID:   devcontext.MustID("company"),
		CreatedAt:   fixture.now,
	}}
	if !reflect.DeepEqual(bindings, want) {
		t.Fatalf("bindings = %#v, want unchanged company binding", bindings)
	}
}

func TestLaunchProjectReturnsPresentationSafeLaunchFailure(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))
	fixture.process.err = launcher.ErrProcessStartFailed

	_, appErr := fixture.service().LaunchProject(LaunchProjectRequest{
		ProjectPath: fixture.projectDir,
		ContextID:   "personal",
	})
	if appErr == nil {
		t.Fatal("launch error = nil, want launch failure")
	}
	if appErr.Code != ErrorCodeLaunch {
		t.Fatalf("error code = %q, want %q", appErr.Code, ErrorCodeLaunch)
	}
	if len(fixture.process.requests) != 1 {
		t.Fatalf("process request count = %d, want 1", len(fixture.process.requests))
	}
}

func TestBindProjectPersistsCanonicalAssociation(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))

	state, appErr := fixture.service().BindProject(BindProjectRequest{
		ProjectPath: ".",
		ContextID:   "personal",
	})
	if appErr != nil {
		t.Fatalf("bind project: %v", appErr)
	}
	if !state.Bound || state.ContextID != "personal" || state.ProjectPath != fixture.projectDir {
		t.Fatalf("binding state = %#v, want personal binding", state)
	}

	stored, err := project.ReadProjectBindingsFile(fixture.bindingsPath)
	if err != nil {
		t.Fatalf("read project bindings: %v", err)
	}
	if !reflect.DeepEqual(stored, []project.Binding{
		{
			ProjectPath: project.Path(fixture.projectDir),
			ContextID:   devcontext.MustID("personal"),
			CreatedAt:   fixture.now,
		},
	}) {
		t.Fatalf("stored bindings = %#v, want personal binding", stored)
	}
}

func TestUnbindProjectReturnsRefreshedBindingState(t *testing.T) {
	t.Run("bound", func(t *testing.T) {
		fixture := newApplicationFixture(t)
		fixture.writeContext(t, fixture.context("personal", "Personal"))
		fixture.writeBindings(t, project.Binding{
			ProjectPath: project.Path(fixture.projectDir),
			ContextID:   devcontext.MustID("personal"),
			CreatedAt:   fixture.now,
		})

		state, appErr := fixture.service().UnbindProject(UnbindProjectRequest{ProjectPath: "."})
		if appErr != nil {
			t.Fatalf("unbind project: %v", appErr)
		}
		if state.Bound || state.ProjectPath != fixture.projectDir {
			t.Fatalf("binding state = %#v, want unbound project", state)
		}

		stored, err := project.ReadProjectBindingsFile(fixture.bindingsPath)
		if err != nil {
			t.Fatalf("read project bindings: %v", err)
		}
		if len(stored) != 0 {
			t.Fatalf("stored bindings = %#v, want empty", stored)
		}
	})

	t.Run("already unbound", func(t *testing.T) {
		fixture := newApplicationFixture(t)

		state, appErr := fixture.service().UnbindProject(UnbindProjectRequest{ProjectPath: "."})
		if appErr != nil {
			t.Fatalf("unbind project: %v", appErr)
		}
		if state.Bound || state.ProjectPath != fixture.projectDir {
			t.Fatalf("binding state = %#v, want unbound project", state)
		}
	})
}

func TestProjectBindingChangesRecordHistoryEvents(t *testing.T) {
	fixture := newApplicationFixture(t)
	logger := &applicationRecordingLogger{}
	fixture.logger = logger
	fixture.writeContext(t, fixture.context("personal", "Personal"))

	if _, appErr := fixture.service().BindProject(BindProjectRequest{ProjectPath: ".", ContextID: "personal"}); appErr != nil {
		t.Fatalf("bind project: %v", appErr)
	}
	if _, appErr := fixture.service().UnbindProject(UnbindProjectRequest{ProjectPath: "."}); appErr != nil {
		t.Fatalf("unbind project: %v", appErr)
	}
	if _, appErr := fixture.service().UnbindProject(UnbindProjectRequest{ProjectPath: "."}); appErr != nil {
		t.Fatalf("repeat unbind project: %v", appErr)
	}

	want := []devlog.EventName{devlog.EventProjectBindingChanged, devlog.EventProjectBindingChanged}
	if got := applicationEventNames(logger.events); !reflect.DeepEqual(got, want) {
		t.Fatalf("event names = %#v, want %#v", got, want)
	}
	for _, event := range logger.events {
		if event.ProjectPath != fixture.projectDir || event.ContextID != "personal" {
			t.Fatalf("binding event = %#v, want project and context", event)
		}
	}
}

func assertContextStates(t *testing.T, got []ContextState, want []ContextState) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("contexts = %#v, want %#v", got, want)
	}
}

func assertConfidenceCheck(t *testing.T, checks []LaunchConfidenceCheck, want LaunchConfidenceCheck) {
	t.Helper()

	for _, check := range checks {
		if check == want {
			return
		}
	}
	t.Fatalf("checks = %#v, want containing %#v", checks, want)
}

func assertFirstRunState(t *testing.T, state LaunchState, projectDir string) {
	t.Helper()

	if !state.FirstRun {
		t.Fatal("first run = false, want true")
	}
	if len(state.Contexts) != 0 {
		t.Fatalf("contexts = %#v, want none", state.Contexts)
	}
	if state.Project != (ProjectState{Name: "current", Path: projectDir}) {
		t.Fatalf("project = %#v, want current project", state.Project)
	}
	if state.Binding.Bound || state.Binding.ProjectPath != projectDir {
		t.Fatalf("binding = %#v, want unbound current project", state.Binding)
	}
	if !state.SelectionRequired {
		t.Fatal("selection required = false, want true")
	}
	if state.ResolutionSource != string(launcher.ResolutionSourceUserSelection) {
		t.Fatalf("resolution source = %q, want user selection", state.ResolutionSource)
	}
	if state.SelectedContextID != "" {
		t.Fatalf("selected context = %q, want empty", state.SelectedContextID)
	}
	if len(state.Warnings) != 0 {
		t.Fatalf("warnings = %#v, want none", state.Warnings)
	}
	if state.Confidence != nil {
		t.Fatalf("confidence = %#v, want nil", state.Confidence)
	}
}

type applicationFixture struct {
	root               string
	homeDir            string
	contextsDir        string
	projectDir         string
	bindingsPath       string
	recentsPath        string
	runningPath        string
	paths              filesystem.PlatformPaths
	now                time.Time
	provider           *applicationFakeProvider
	providerRegistry   provider.Registry
	editor             *applicationFakeEditor
	toolRegistry       codingtool.Registry
	process            *applicationFakeProcessLauncher
	storagePermissions filesystem.StoragePermissions
	logger             devlog.Logger
}

func newApplicationFixture(t *testing.T) applicationFixture {
	t.Helper()

	root := t.TempDir()
	fixture := applicationFixture{
		root:         root,
		homeDir:      filepath.Join(root, "home"),
		contextsDir:  filepath.Join(root, "contexts"),
		projectDir:   filepath.Join(root, "projects", "current"),
		bindingsPath: filepath.Join(root, "projects.toml"),
		recentsPath:  filepath.Join(root, "recents.toml"),
		runningPath:  filepath.Join(root, "running.toml"),
		now:          time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC),
		provider:     &applicationFakeProvider{id: "fake"},
		editor:       &applicationFakeEditor{},
		process:      &applicationFakeProcessLauncher{},
	}
	fixture.paths = filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return fixture.homeDir, nil
	})
	devContextHomeDir, err := fixture.paths.DevContextHomeDir()
	if err != nil {
		t.Fatalf("dev context home: %v", err)
	}
	fixture.contextsDir = filepath.Join(devContextHomeDir, "contexts")
	mkdir(t, fixture.homeDir)
	mkdir(t, fixture.contextsDir)
	mkdir(t, fixture.projectDir)
	return fixture
}

func (f applicationFixture) service() *Service {
	registry := f.providerRegistry
	if registry.IsZero() {
		registry = provider.MustNewRegistry([]provider.Provider{f.provider}, f.provider.ID())
	}

	toolRegistry := f.toolRegistry
	if toolRegistry.IsZero() {
		toolRegistry = codingtool.MustNewRegistry([]codingtool.RegisteredTool{{Integration: f.editor, DisplayName: "Fake Tool"}}, f.editor.ID())
	}

	return NewServiceWithDependencies(Dependencies{
		Contexts:            devcontext.NewRepository(f.contextsDir),
		Projects:            project.NewRepository(f.bindingsPath, f.paths),
		RecentProjects:      project.NewRecentRepository(f.recentsPath),
		RunningEnvironments: coreRunning.NewRepository(f.runningPath),
		Paths:               f.paths,
		ProviderRegistry:    registry,
		ToolRegistry:        toolRegistry,
		ProcessLauncher:     f.process,
		StoragePermissions:  f.storagePermissions,
		ParentEnvironment:   []string{"PATH=/fixture/bin"},
		WorkingDirectory:    f.projectDir,
		DetachMode:          launcher.DetachModeAttached,
		Logger:              f.logger,
		Now: func() time.Time {
			return f.now
		},
	})
}

func (f applicationFixture) context(id string, name string) devcontext.Context {
	return devcontext.Context{
		ID:   devcontext.MustID(id),
		Name: name,
		Tool: codingtool.LaunchTarget{DefaultTool: f.editor.ID(), Tools: map[codingtool.ID]codingtool.Config{f.editor.ID(): {}}},
		Providers: provider.Configs{
			"fake": {Enabled: true},
		},
		Metadata: devcontext.Metadata{
			"accent": "blue",
		},
		CreatedAt: f.now,
	}
}

func (f applicationFixture) writeContext(t *testing.T, ctx devcontext.Context) {
	t.Helper()

	contextPaths, err := filesystem.DeriveContextPaths(f.paths, ctx.ID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	contextPaths = contextPaths.WithProviderStorageDirs(enabledProviderIDs(ctx))
	mkdir(t, contextPaths.RootDir)
	mkdir(t, contextPaths.ToolStorageRootDir)
	mkdir(t, contextPaths.ToolStorageDir(ctx.Tool.DefaultTool))
	for _, dir := range contextPaths.ProviderStorageDirs {
		mkdir(t, dir)
	}
	if err := devcontext.NewRepository(f.contextsDir).Write(ctx); err != nil {
		t.Fatalf("write context %q: %v", ctx.ID.String(), err)
	}
}

func (f applicationFixture) writeBindings(t *testing.T, bindings ...project.Binding) {
	t.Helper()

	if err := project.WriteProjectBindingsFile(f.bindingsPath, bindings); err != nil {
		t.Fatalf("write project bindings: %v", err)
	}
}

func mkdir(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatalf("create directory %q: %v", path, err)
	}
}

func removeAll(t *testing.T, path string) {
	t.Helper()

	if err := os.RemoveAll(path); err != nil {
		t.Fatalf("remove %q: %v", path, err)
	}
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write file %q: %v", path, err)
	}
}

func writeApplicationJSONFixture(t *testing.T, path string, value any) {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json fixture: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create fixture directory %q: %v", filepath.Dir(path), err)
	}
	writeFile(t, path, data)
}

type applicationSecondTool struct {
	requests []codingtool.CommandRequest
}

func (applicationSecondTool) ID() codingtool.ID {
	return "second-tool"
}

func (t *applicationSecondTool) DetectExecutable(codingtool.Config) (codingtool.Executable, error) {
	return "/fixture/second-tool", nil
}

func (t *applicationSecondTool) BuildLaunchCommand(request codingtool.CommandRequest) (codingtool.Command, error) {
	t.requests = append(t.requests, request)
	return codingtool.Command{
		Executable: request.Executable,
		Arguments: codingtool.Arguments{
			"--state-dir", request.Paths.StorageDir,
			"--mode", request.Config.Options["mode"],
			request.ProjectPath,
		},
	}, nil
}

func applicationTestJWT(t *testing.T, claims map[string]string) string {
	t.Helper()

	header, err := json.Marshal(map[string]string{"alg": "none"})
	if err != nil {
		t.Fatalf("marshal jwt header: %v", err)
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal jwt payload: %v", err)
	}
	return strings.Join([]string{
		base64.RawURLEncoding.EncodeToString(header),
		base64.RawURLEncoding.EncodeToString(payload),
		"signature",
	}, ".")
}

type applicationRecordingLogger struct {
	events []devlog.Event
}

func (l *applicationRecordingLogger) Record(event devlog.Event) error {
	l.events = append(l.events, event)
	return nil
}

func applicationEventNames(events []devlog.Event) []devlog.EventName {
	names := make([]devlog.EventName, len(events))
	for i, event := range events {
		names[i] = event.Name
	}
	return names
}
