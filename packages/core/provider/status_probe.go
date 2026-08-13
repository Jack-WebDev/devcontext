package provider

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// StatusProbe reads only local process and filesystem readiness signals.
type StatusProbe interface {
	LookPath(file string) (string, error)
	ReadDir(path string) ([]os.DirEntry, error)
}

type defaultStatusProbe struct{}

func (defaultStatusProbe) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (defaultStatusProbe) ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}

func resolveStatusProbe(probe StatusProbe) StatusProbe {
	if probe != nil {
		return probe
	}
	return defaultStatusProbe{}
}

func detectLocalStatus(probe StatusProbe, command string, displayName string, directory string) (Status, error) {
	probe = resolveStatusProbe(probe)

	if _, err := probe.LookPath(command); err != nil {
		return UnavailableStatus(fmt.Sprintf("%s command %q was not found", displayName, command)), nil
	}
	if directory == "" {
		return DirectoryMissingStatus(fmt.Sprintf("%s context directory is not configured", displayName)), nil
	}

	entries, err := probe.ReadDir(directory)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DirectoryMissingStatus(fmt.Sprintf("%s context directory is missing", displayName)), nil
		}
		return UnavailableStatus(fmt.Sprintf("%s context directory could not be inspected", displayName)), nil
	}
	if len(entries) == 0 {
		return NotConfiguredStatus(fmt.Sprintf("%s context directory is empty", displayName)), nil
	}

	return ReadyStatus(), nil
}
