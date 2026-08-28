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
		diagnostics:         application.DiagnosticsState{Groups: []application.DiagnosticGroup{}},
		repairActions:       application.RepairActionsState{Actions: []application.RepairAction{}},
		repairResult:        application.RunRepairActionResult{ActionID: "recheck-provider-files"},
		history:             application.HistoryState{Entries: []application.HistoryEntry{}},
		runningEnvironments: application.RunningEnvironmentsState{Environments: []application.RunningEnvironmentState{}},
	}
	app := New(service)
	app.Startup(context.Background())

	if app.ctx == nil {
		t.Fatal("startup context was not stored")
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
	app := New(&fakeService{launchStateErr: want})

	got := app.GetLaunchState(application.GetLaunchStateRequest{})
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("error value = %#v, want %#v", got, want)
	}
}

type fakeService struct {
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

	contextDetailsRequest application.GetContextDetailsRequest
	contextDetails        application.ContextDetailsState
	contextDetailsErr     *application.Error

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
