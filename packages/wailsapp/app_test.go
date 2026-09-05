package wailsapp

import (
	"context"
	"reflect"
	"testing"

	"devctx/packages/application"
	"devctx/packages/core/project"
)

func TestAppDelegatesApplicationMethodsToService(t *testing.T) {
	service := &fakeService{
		validatedProject: application.ProjectState{Name: "api", Path: "/work/api"},
		launchState: application.LaunchState{
			Project: application.ProjectState{Name: "api", Path: "/work/api"},
		},
		homeDashboard: application.HomeDashboardState{
			Project: application.ProjectState{Name: "api", Path: "/work/api"},
		},
		recentProjects: application.RecentProjectsState{Projects: []application.RecentProjectState{{
			Project: application.ProjectState{Name: "api", Path: "/work/api"}, ContextID: "personal",
		}}},
		contexts: application.ContextListState{Contexts: []application.ContextListItem{{
			Context: application.ContextState{ID: "personal", Name: "Personal"}, ProjectCount: 1,
		}}},
		contextDetails: application.ContextDetailsState{
			Context: application.ContextState{ID: "personal", Name: "Personal"}, Location: "/contexts/personal",
		},
		trustCenter: application.TrustCenterState{CredentialSync: application.TrustCredentialSyncProtection{Enabled: false}},
		launchResult: application.LaunchProjectResult{
			Project: application.ProjectState{Name: "api", Path: "/work/api"},
			Context: application.ContextState{ID: "personal", Name: "Personal"},
		},
		preflightResult: application.PreflightLaunchProjectResult{
			Project:    application.ProjectState{Name: "api", Path: "/work/api"},
			Context:    application.ContextState{ID: "personal", Name: "Personal"},
			Confidence: application.LaunchConfidenceState{ContextID: "personal", Status: application.LaunchConfidenceReady},
		},
		bindResult: application.ProjectBindingState{
			ProjectPath: "/work/api",
			Bound:       true,
			ContextID:   "personal",
		},
		unbindResult: application.ProjectBindingState{
			ProjectPath: "/work/api",
			Bound:       false,
		},
		createContextResult: application.CreateContextResult{
			Context: application.ContextState{ID: "personal", Name: "Personal"},
		},
		contextTemplates: application.ContextTemplatesState{Templates: []application.ContextTemplateState{{ID: "personal", Name: "Personal"}}},
		duplicateContextResult: application.DuplicateContextResult{
			Context: application.ContextState{ID: "personal-copy", Name: "Personal copy"},
		},
		contextMetadataExport: application.ContextMetadataExport{Version: application.ContextTransferVersion, Context: application.ContextTransferMetadata{Name: "Personal"}},
		importContextMetadataResult: application.ImportContextMetadataResult{
			Context: application.ContextState{ID: "imported", Name: "Personal"},
		},
		diagnostics:         application.DiagnosticsState{Groups: []application.DiagnosticGroup{}},
		repairActions:       application.RepairActionsState{Actions: []application.RepairAction{}},
		repairResult:        application.RunRepairActionResult{ActionID: "recheck-provider-files"},
		history:             application.HistoryState{Entries: []application.HistoryEntry{}},
		runningEnvironments: application.RunningEnvironmentsState{Environments: []application.RunningEnvironmentState{}},
	}
	app := New(service, ManagementMode())
	app.Startup(context.Background())

	if app.ctx == nil {
		t.Fatal("startup context was not stored")
	}

	validationRequest := application.ValidateProjectDirectoryRequest{ProjectPath: "/work/api"}
	if project := app.ValidateProjectDirectory(validationRequest); !reflect.DeepEqual(project, service.validatedProject) {
		t.Fatalf("validated project = %#v, want %#v", project, service.validatedProject)
	}
	if service.validationRequest != validationRequest {
		t.Fatalf("validation request = %#v, want %#v", service.validationRequest, validationRequest)
	}

	stateRequest := application.GetLaunchStateRequest{ProjectPath: "/work/api"}
	state := app.GetLaunchState(stateRequest)
	if !reflect.DeepEqual(state, service.launchState) {
		t.Fatalf("launch state = %#v, want %#v", state, service.launchState)
	}
	if service.launchStateRequest != stateRequest {
		t.Fatalf("launch state request = %#v, want %#v", service.launchStateRequest, stateRequest)
	}

	dashboardRequest := application.GetHomeDashboardRequest{ProjectPath: "/work/api"}
	dashboard := app.GetHomeDashboard(dashboardRequest)
	if !reflect.DeepEqual(dashboard, service.homeDashboard) {
		t.Fatalf("home dashboard = %#v, want %#v", dashboard, service.homeDashboard)
	}
	if service.homeDashboardRequest != dashboardRequest {
		t.Fatalf("home dashboard request = %#v, want %#v", service.homeDashboardRequest, dashboardRequest)
	}

	recentProjects := app.GetRecentProjects()
	if !reflect.DeepEqual(recentProjects, service.recentProjects) {
		t.Fatalf("recent projects = %#v, want %#v", recentProjects, service.recentProjects)
	}

	contexts := app.GetContexts()
	if !reflect.DeepEqual(contexts, service.contexts) {
		t.Fatalf("contexts = %#v, want %#v", contexts, service.contexts)
	}

	detailsRequest := application.GetContextDetailsRequest{ContextID: "personal"}
	details := app.GetContextDetails(detailsRequest)
	if !reflect.DeepEqual(details, service.contextDetails) {
		t.Fatalf("context details = %#v, want %#v", details, service.contextDetails)
	}
	if service.contextDetailsRequest != detailsRequest {
		t.Fatalf("context details request = %#v, want %#v", service.contextDetailsRequest, detailsRequest)
	}
	updateDetailsRequest := application.UpdateContextDetailsRequest{ContextID: "personal", Name: "Personal work", Purpose: "Work"}
	if updated := app.UpdateContextDetails(updateDetailsRequest); !reflect.DeepEqual(updated, service.updateContextDetailsResult) {
		t.Fatalf("updated context details = %#v, want %#v", updated, service.updateContextDetailsResult)
	}
	if service.updateContextDetailsRequest != updateDetailsRequest {
		t.Fatalf("update details request = %#v, want %#v", service.updateContextDetailsRequest, updateDetailsRequest)
	}
	updateAppearanceRequest := application.UpdateContextAppearanceRequest{ContextID: "personal", Icon: "building", Accent: "amber"}
	if updated := app.UpdateContextAppearance(updateAppearanceRequest); !reflect.DeepEqual(updated, service.updateContextAppearanceResult) {
		t.Fatalf("updated context appearance = %#v, want %#v", updated, service.updateContextAppearanceResult)
	}
	if service.updateContextAppearanceRequest != updateAppearanceRequest {
		t.Fatalf("update appearance request = %#v, want %#v", service.updateContextAppearanceRequest, updateAppearanceRequest)
	}
	if trustCenter := app.GetTrustCenter(); !reflect.DeepEqual(trustCenter, service.trustCenter) {
		t.Fatalf("trust center = %#v, want %#v", trustCenter, service.trustCenter)
	}

	preflightRequest := application.PreflightLaunchProjectRequest{ProjectPath: "/work/api", ContextID: "personal"}
	preflight := app.PreflightLaunchProject(preflightRequest)
	if !reflect.DeepEqual(preflight, service.preflightResult) {
		t.Fatalf("preflight result = %#v, want %#v", preflight, service.preflightResult)
	}
	if service.preflightRequest != preflightRequest {
		t.Fatalf("preflight request = %#v, want %#v", service.preflightRequest, preflightRequest)
	}

	launchRequest := application.LaunchProjectRequest{ProjectPath: "/work/api", ContextID: "personal"}
	launch := app.LaunchProject(launchRequest)
	if !reflect.DeepEqual(launch, service.launchResult) {
		t.Fatalf("launch result = %#v, want %#v", launch, service.launchResult)
	}
	if service.launchRequest != launchRequest {
		t.Fatalf("launch request = %#v, want %#v", service.launchRequest, launchRequest)
	}

	bindRequest := application.BindProjectRequest{ProjectPath: "/work/api", ContextID: "personal"}
	binding := app.BindProject(bindRequest)
	if !reflect.DeepEqual(binding, service.bindResult) {
		t.Fatalf("binding = %#v, want %#v", binding, service.bindResult)
	}
	if service.bindRequest != bindRequest {
		t.Fatalf("bind request = %#v, want %#v", service.bindRequest, bindRequest)
	}

	unbindRequest := application.UnbindProjectRequest{ProjectPath: "/work/api"}
	unbound := app.UnbindProject(unbindRequest)
	if !reflect.DeepEqual(unbound, service.unbindResult) {
		t.Fatalf("unbound state = %#v, want %#v", unbound, service.unbindResult)
	}
	if service.unbindRequest != unbindRequest {
		t.Fatalf("unbind request = %#v, want %#v", service.unbindRequest, unbindRequest)
	}

	createContextRequest := application.CreateContextRequest{ContextID: "personal"}
	contextResult := app.CreateContext(createContextRequest)
	if !reflect.DeepEqual(contextResult, service.createContextResult) {
		t.Fatalf("create context result = %#v, want %#v", contextResult, service.createContextResult)
	}
	if !reflect.DeepEqual(service.createContextRequest, createContextRequest) {
		t.Fatalf("create context request = %#v, want %#v", service.createContextRequest, createContextRequest)
	}
	if templates := app.GetContextTemplates(); !reflect.DeepEqual(templates, service.contextTemplates) {
		t.Fatalf("context templates = %#v, want %#v", templates, service.contextTemplates)
	}
	duplicateRequest := application.DuplicateContextRequest{SourceContextID: "personal", ContextID: "personal-copy"}
	if duplicate := app.DuplicateContext(duplicateRequest); !reflect.DeepEqual(duplicate, service.duplicateContextResult) {
		t.Fatalf("duplicate context = %#v, want %#v", duplicate, service.duplicateContextResult)
	}
	if service.duplicateContextRequest != duplicateRequest {
		t.Fatalf("duplicate context request = %#v, want %#v", service.duplicateContextRequest, duplicateRequest)
	}
	exportRequest := application.ExportContextMetadataRequest{ContextID: "personal"}
	if exported := app.ExportContextMetadata(exportRequest); !reflect.DeepEqual(exported, service.contextMetadataExport) {
		t.Fatalf("context metadata export = %#v, want %#v", exported, service.contextMetadataExport)
	}
	if service.contextMetadataExportRequest != exportRequest {
		t.Fatalf("context metadata export request = %#v, want %#v", service.contextMetadataExportRequest, exportRequest)
	}
	importRequest := application.ImportContextMetadataRequest{ContextID: "imported", Export: service.contextMetadataExport}
	if imported := app.ImportContextMetadata(importRequest); !reflect.DeepEqual(imported, service.importContextMetadataResult) {
		t.Fatalf("context metadata import = %#v, want %#v", imported, service.importContextMetadataResult)
	}
	if !reflect.DeepEqual(service.importContextMetadataRequest, importRequest) {
		t.Fatalf("context metadata import request = %#v, want %#v", service.importContextMetadataRequest, importRequest)
	}

	diagnosticsRequest := application.GetDiagnosticsRequest{ContextID: "personal"}
	diagnostics := app.GetDiagnostics(diagnosticsRequest)
	if !reflect.DeepEqual(diagnostics, service.diagnostics) {
		t.Fatalf("diagnostics = %#v, want %#v", diagnostics, service.diagnostics)
	}
	if service.diagnosticsRequest != diagnosticsRequest {
		t.Fatalf("diagnostics request = %#v, want %#v", service.diagnosticsRequest, diagnosticsRequest)
	}

	repairActionsRequest := application.GetRepairActionsRequest{ContextID: "personal"}
	repairActions := app.GetRepairActions(repairActionsRequest)
	if !reflect.DeepEqual(repairActions, service.repairActions) {
		t.Fatalf("repair actions = %#v, want %#v", repairActions, service.repairActions)
	}
	if service.repairActionsRequest != repairActionsRequest {
		t.Fatalf("repair actions request = %#v, want %#v", service.repairActionsRequest, repairActionsRequest)
	}

	runRepairRequest := application.RunRepairActionRequest{ContextID: "personal", ActionID: "recheck-provider-files"}
	repairResult := app.RunRepairAction(runRepairRequest)
	if !reflect.DeepEqual(repairResult, service.repairResult) {
		t.Fatalf("repair result = %#v, want %#v", repairResult, service.repairResult)
	}
	if service.runRepairRequest != runRepairRequest {
		t.Fatalf("run repair request = %#v, want %#v", service.runRepairRequest, runRepairRequest)
	}

	history := app.GetHistory()
	if !reflect.DeepEqual(history, service.history) {
		t.Fatalf("history = %#v, want %#v", history, service.history)
	}
	runningEnvironments := app.GetRunningEnvironments()
	if !reflect.DeepEqual(runningEnvironments, service.runningEnvironments) {
		t.Fatalf("running environments = %#v, want %#v", runningEnvironments, service.runningEnvironments)
	}
}

func TestAppReturnsApplicationErrorsAsSingleValues(t *testing.T) {
	want := application.NewError(project.ErrProjectDirectoryNotFound)
	app := New(&fakeService{launchStateErr: want}, ManagementMode())

	got := app.GetLaunchState(application.GetLaunchStateRequest{})
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("error value = %#v, want %#v", got, want)
	}
}

func TestNewRetainsHostSelectedApplicationMode(t *testing.T) {
	service := &fakeService{}
	tests := []struct {
		name string
		mode ApplicationMode
	}{
		{
			name: "management",
			mode: ManagementMode(),
		},
		{
			name: "launcher",
			mode: LauncherMode("/work/api"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New(service, tt.mode)
			if got := app.GetApplicationMode(); !reflect.DeepEqual(got, tt.mode) {
				t.Fatalf("application mode = %#v, want %#v", got, tt.mode)
			}
		})
	}
}

type fakeService struct {
	validationRequest   application.ValidateProjectDirectoryRequest
	validatedProject    application.ProjectState
	validatedProjectErr *application.Error

	settings           application.SettingsState
	launchStateRequest application.GetLaunchStateRequest
	launchState        application.LaunchState
	launchStateErr     *application.Error

	homeDashboardRequest application.GetHomeDashboardRequest
	homeDashboard        application.HomeDashboardState
	homeDashboardErr     *application.Error

	recentProjects    application.RecentProjectsState
	recentProjectsErr *application.Error

	contexts    application.ContextListState
	contextsErr *application.Error

	contextDetailsRequest          application.GetContextDetailsRequest
	contextDetails                 application.ContextDetailsState
	contextDetailsErr              *application.Error
	updateContextDetailsRequest    application.UpdateContextDetailsRequest
	updateContextDetailsResult     application.ContextState
	updateContextDetailsErr        *application.Error
	updateContextAppearanceRequest application.UpdateContextAppearanceRequest
	updateContextAppearanceResult  application.ContextState
	updateContextAppearanceErr     *application.Error

	trustCenter    application.TrustCenterState
	trustCenterErr *application.Error

	projects    application.ProjectsState
	projectsErr *application.Error

	preflightRequest application.PreflightLaunchProjectRequest
	preflightResult  application.PreflightLaunchProjectResult
	preflightErr     *application.Error

	launchRequest application.LaunchProjectRequest
	launchResult  application.LaunchProjectResult
	launchErr     *application.Error

	bindRequest application.BindProjectRequest
	bindResult  application.ProjectBindingState
	bindErr     *application.Error

	unbindRequest application.UnbindProjectRequest
	unbindResult  application.ProjectBindingState
	unbindErr     *application.Error

	createContextRequest application.CreateContextRequest
	createContextResult  application.CreateContextResult
	createContextErr     *application.Error

	contextTemplates        application.ContextTemplatesState
	duplicateContextRequest application.DuplicateContextRequest
	duplicateContextResult  application.DuplicateContextResult
	duplicateContextErr     *application.Error

	contextMetadataExportRequest application.ExportContextMetadataRequest
	contextMetadataExport        application.ContextMetadataExport
	contextMetadataExportErr     *application.Error
	importContextMetadataRequest application.ImportContextMetadataRequest
	importContextMetadataResult  application.ImportContextMetadataResult
	importContextMetadataErr     *application.Error

	diagnosticsRequest application.GetDiagnosticsRequest
	diagnostics        application.DiagnosticsState
	diagnosticsErr     *application.Error

	repairActionsRequest application.GetRepairActionsRequest
	repairActions        application.RepairActionsState
	repairActionsErr     *application.Error

	runRepairRequest application.RunRepairActionRequest
	repairResult     application.RunRepairActionResult
	repairErr        *application.Error

	history    application.HistoryState
	historyErr *application.Error

	runningEnvironments    application.RunningEnvironmentsState
	runningEnvironmentsErr *application.Error
}

func (s *fakeService) ValidateProjectDirectory(request application.ValidateProjectDirectoryRequest) (application.ProjectState, *application.Error) {
	s.validationRequest = request
	return s.validatedProject, s.validatedProjectErr
}

func (s *fakeService) GetSettings() (application.SettingsState, *application.Error) {
	return s.settings, nil
}
func (s *fakeService) UpdateSettings(request application.UpdateSettingsRequest) (application.SettingsState, *application.Error) {
	s.settings = application.SettingsState(request)
	return s.settings, nil
}
func (s *fakeService) GetTrayState() (application.TrayState, *application.Error) {
	return application.TrayState{}, nil
}

func (s *fakeService) GetLaunchState(request application.GetLaunchStateRequest) (application.LaunchState, *application.Error) {
	s.launchStateRequest = request
	return s.launchState, s.launchStateErr
}

func (s *fakeService) GetHomeDashboard(request application.GetHomeDashboardRequest) (application.HomeDashboardState, *application.Error) {
	s.homeDashboardRequest = request
	return s.homeDashboard, s.homeDashboardErr
}

func (s *fakeService) GetRecentProjects() (application.RecentProjectsState, *application.Error) {
	return s.recentProjects, s.recentProjectsErr
}

func (s *fakeService) GetContexts() (application.ContextListState, *application.Error) {
	return s.contexts, s.contextsErr
}

func (s *fakeService) GetContextDetails(request application.GetContextDetailsRequest) (application.ContextDetailsState, *application.Error) {
	s.contextDetailsRequest = request
	return s.contextDetails, s.contextDetailsErr
}

func (s *fakeService) UpdateContextDetails(request application.UpdateContextDetailsRequest) (application.ContextState, *application.Error) {
	s.updateContextDetailsRequest = request
	return s.updateContextDetailsResult, s.updateContextDetailsErr
}

func (s *fakeService) UpdateContextAppearance(request application.UpdateContextAppearanceRequest) (application.ContextState, *application.Error) {
	s.updateContextAppearanceRequest = request
	return s.updateContextAppearanceResult, s.updateContextAppearanceErr
}

func (s *fakeService) ArchiveContext(application.ArchiveContextRequest) (application.ContextState, *application.Error) {
	return application.ContextState{}, nil
}
func (s *fakeService) RestoreContext(application.RestoreContextRequest) (application.ContextState, *application.Error) {
	return application.ContextState{}, nil
}
func (s *fakeService) PreviewDeleteContext(application.DeleteContextPreviewRequest) (application.DeleteContextPreview, *application.Error) {
	return application.DeleteContextPreview{}, nil
}
func (s *fakeService) DeleteContext(application.DeleteContextRequest) (application.DeleteContextResult, *application.Error) {
	return application.DeleteContextResult{}, nil
}

func (s *fakeService) GetTrustCenter() (application.TrustCenterState, *application.Error) {
	return s.trustCenter, s.trustCenterErr
}

func (s *fakeService) GetProjects() (application.ProjectsState, *application.Error) {
	return s.projects, s.projectsErr
}

func (s *fakeService) GetDiagnostics(request application.GetDiagnosticsRequest) (application.DiagnosticsState, *application.Error) {
	s.diagnosticsRequest = request
	return s.diagnostics, s.diagnosticsErr
}

func (s *fakeService) GetRepairActions(request application.GetRepairActionsRequest) (application.RepairActionsState, *application.Error) {
	s.repairActionsRequest = request
	return s.repairActions, s.repairActionsErr
}

func (s *fakeService) RunRepairAction(request application.RunRepairActionRequest) (application.RunRepairActionResult, *application.Error) {
	s.runRepairRequest = request
	return s.repairResult, s.repairErr
}

func (s *fakeService) GetHistory() (application.HistoryState, *application.Error) {
	return s.history, s.historyErr
}

func (s *fakeService) GetRunningEnvironments() (application.RunningEnvironmentsState, *application.Error) {
	return s.runningEnvironments, s.runningEnvironmentsErr
}

func (s *fakeService) PreflightLaunchProject(request application.PreflightLaunchProjectRequest) (application.PreflightLaunchProjectResult, *application.Error) {
	s.preflightRequest = request
	return s.preflightResult, s.preflightErr
}

func (s *fakeService) LaunchProject(request application.LaunchProjectRequest) (application.LaunchProjectResult, *application.Error) {
	s.launchRequest = request
	return s.launchResult, s.launchErr
}

func (s *fakeService) BindProject(request application.BindProjectRequest) (application.ProjectBindingState, *application.Error) {
	s.bindRequest = request
	return s.bindResult, s.bindErr
}

func (s *fakeService) UnbindProject(request application.UnbindProjectRequest) (application.ProjectBindingState, *application.Error) {
	s.unbindRequest = request
	return s.unbindResult, s.unbindErr
}

func (s *fakeService) CreateContext(request application.CreateContextRequest) (application.CreateContextResult, *application.Error) {
	s.createContextRequest = request
	return s.createContextResult, s.createContextErr
}

func (s *fakeService) GetContextTemplates() application.ContextTemplatesState {
	return s.contextTemplates
}

func (s *fakeService) DuplicateContext(request application.DuplicateContextRequest) (application.DuplicateContextResult, *application.Error) {
	s.duplicateContextRequest = request
	return s.duplicateContextResult, s.duplicateContextErr
}

func (s *fakeService) ExportContextMetadata(request application.ExportContextMetadataRequest) (application.ContextMetadataExport, *application.Error) {
	s.contextMetadataExportRequest = request
	return s.contextMetadataExport, s.contextMetadataExportErr
}

func (s *fakeService) ImportContextMetadata(request application.ImportContextMetadataRequest) (application.ImportContextMetadataResult, *application.Error) {
	s.importContextMetadataRequest = request
	return s.importContextMetadataResult, s.importContextMetadataErr
}
