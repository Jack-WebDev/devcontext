package wailsapp

import (
	"context"
	"reflect"
	"testing"

	"devctx/packages/application"
)

func TestAppDelegatesApplicationMethodsToService(t *testing.T) {
	service := &fakeService{
		launchState: application.LaunchState{
			Project: application.ProjectState{Name: "api", Path: "/work/api"},
		},
		launchResult: application.LaunchProjectResult{
			Project: application.ProjectState{Name: "api", Path: "/work/api"},
			Context: application.ContextState{ID: "personal", Name: "Personal"},
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
	}
	app := New(service)
	app.Startup(context.Background())

	if app.ctx == nil {
		t.Fatal("startup context was not stored")
	}

	stateRequest := application.GetLaunchStateRequest{ProjectPath: "/work/api"}
	state, stateErr := app.GetLaunchState(stateRequest)
	if stateErr != nil {
		t.Fatalf("get launch state error = %v, want nil", stateErr)
	}
	if !reflect.DeepEqual(state, service.launchState) {
		t.Fatalf("launch state = %#v, want %#v", state, service.launchState)
	}
	if service.launchStateRequest != stateRequest {
		t.Fatalf("launch state request = %#v, want %#v", service.launchStateRequest, stateRequest)
	}

	launchRequest := application.LaunchProjectRequest{ProjectPath: "/work/api", ContextID: "personal"}
	launch, launchErr := app.LaunchProject(launchRequest)
	if launchErr != nil {
		t.Fatalf("launch project error = %v, want nil", launchErr)
	}
	if !reflect.DeepEqual(launch, service.launchResult) {
		t.Fatalf("launch result = %#v, want %#v", launch, service.launchResult)
	}
	if service.launchRequest != launchRequest {
		t.Fatalf("launch request = %#v, want %#v", service.launchRequest, launchRequest)
	}

	bindRequest := application.BindProjectRequest{ProjectPath: "/work/api", ContextID: "personal"}
	binding, bindErr := app.BindProject(bindRequest)
	if bindErr != nil {
		t.Fatalf("bind project error = %v, want nil", bindErr)
	}
	if !reflect.DeepEqual(binding, service.bindResult) {
		t.Fatalf("binding = %#v, want %#v", binding, service.bindResult)
	}
	if service.bindRequest != bindRequest {
		t.Fatalf("bind request = %#v, want %#v", service.bindRequest, bindRequest)
	}

	unbindRequest := application.UnbindProjectRequest{ProjectPath: "/work/api"}
	unbound, unbindErr := app.UnbindProject(unbindRequest)
	if unbindErr != nil {
		t.Fatalf("unbind project error = %v, want nil", unbindErr)
	}
	if !reflect.DeepEqual(unbound, service.unbindResult) {
		t.Fatalf("unbound state = %#v, want %#v", unbound, service.unbindResult)
	}
	if service.unbindRequest != unbindRequest {
		t.Fatalf("unbind request = %#v, want %#v", service.unbindRequest, unbindRequest)
	}

	createContextRequest := application.CreateContextRequest{ContextID: "personal"}
	contextResult, createContextErr := app.CreateContext(createContextRequest)
	if createContextErr != nil {
		t.Fatalf("create context error = %v, want nil", createContextErr)
	}
	if !reflect.DeepEqual(contextResult, service.createContextResult) {
		t.Fatalf("create context result = %#v, want %#v", contextResult, service.createContextResult)
	}
	if service.createContextRequest != createContextRequest {
		t.Fatalf("create context request = %#v, want %#v", service.createContextRequest, createContextRequest)
	}
}

type fakeService struct {
	launchStateRequest application.GetLaunchStateRequest
	launchState        application.LaunchState
	launchStateErr     *application.Error

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
}

func (s *fakeService) GetLaunchState(request application.GetLaunchStateRequest) (application.LaunchState, *application.Error) {
	s.launchStateRequest = request
	return s.launchState, s.launchStateErr
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
