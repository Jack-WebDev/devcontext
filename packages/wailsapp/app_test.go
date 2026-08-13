package wailsapp

import (
	"context"
	"testing"
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

type fakeService struct {
	greetedName string
}

func (s *fakeService) Greet(name string) string {
	s.greetedName = name
	return "fake greeting for " + name
}
