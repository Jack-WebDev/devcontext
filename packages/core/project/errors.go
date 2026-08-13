package project

import "fmt"

// PathError describes a project path failure with the affected path and
// operation.
type PathError struct {
	Path      string
	Operation string
	Err       error
	Cause     error
}

func (e *PathError) Error() string {
	switch {
	case e == nil:
		return ""
	case e.Operation != "" && e.Path != "" && e.Cause != nil:
		return fmt.Sprintf("%s project path %q: %v: %v", e.Operation, e.Path, e.Err, e.Cause)
	case e.Operation != "" && e.Path != "":
		return fmt.Sprintf("%s project path %q: %v", e.Operation, e.Path, e.Err)
	case e.Path != "" && e.Cause != nil:
		return fmt.Sprintf("project path %q: %v: %v", e.Path, e.Err, e.Cause)
	case e.Path != "":
		return fmt.Sprintf("project path %q: %v", e.Path, e.Err)
	default:
		return fmt.Sprintf("project path: %v", e.Err)
	}
}

func (e *PathError) Unwrap() []error {
	if e == nil {
		return nil
	}
	if e.Cause == nil {
		return []error{e.Err}
	}
	return []error{e.Err, e.Cause}
}

func newPathError(path string, operation string, err error, cause error) error {
	return &PathError{
		Path:      path,
		Operation: operation,
		Err:       err,
		Cause:     cause,
	}
}
