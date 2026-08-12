package application

import "fmt"

// Service coordinates product use cases.
//
// The current starter UI only needs a greeting. Real launch, binding, and
// configuration use cases will be added here in later phases.
type Service struct{}

// NewService creates an application service.
func NewService() *Service {
	return &Service{}
}

// Greet returns a greeting for the current starter UI.
func (s *Service) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
