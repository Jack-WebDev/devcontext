package wailsapp

import (
	"context"
	"reflect"
	"testing"

	"devctx/packages/application"
)

func TestAppDelegatesStarterMethodToService(t *testing.T) {
	service := &fakeService{}
	app := New(service)

	app.Startup(context.Background())
	got := app.Greet("Alex")

	if got != "fake greeting for Alex" {
		t.Fatalf("greeting = %q, want delegated greeting", got)
	}
	if service.greetedName != "Alex" {
		t.Fatalf("service greeted name = %q, want Alex", service.greetedName)
	}
	if app.ctx == nil {
		t.Fatal("startup context was not stored")
	}
}

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
	}
	app := New(service)

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
}

type fakeService struct {
	greetedName string

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
}

func (s *fakeService) Greet(name string) string {
	s.greetedName = name
	return "fake greeting for " + name
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
