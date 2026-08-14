package provider

import (
	"errors"
	"fmt"
	"os"
)

// StatusProbe reads only local filesystem readiness signals.
type StatusProbe interface {
	ReadDir(path string) ([]os.DirEntry, error)
}

type defaultStatusProbe struct{}

func (defaultStatusProbe) ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}

func resolveStatusProbe(probe StatusProbe) StatusProbe {
	if probe != nil {
		return probe
	}
	return defaultStatusProbe{}
}

func detectLocalStatus(probe StatusProbe, displayName string, directory string) (Status, error) {
	probe = resolveStatusProbe(probe)

	if directory == "" {
		return NotConfiguredStatus(fmt.Sprintf("%s isolated provider state was not found", displayName)), nil
	}

	entries, err := probe.ReadDir(directory)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return NotConfiguredStatus(fmt.Sprintf("%s isolated provider state was not found", displayName)), nil
		}
		return UnavailableStatus(fmt.Sprintf("%s context directory could not be inspected", displayName)), nil
	}
	if len(entries) == 0 {
		return NotConfiguredStatus(fmt.Sprintf("%s isolated provider state was not found", displayName)), nil
	}

	return ConfiguredStatus(), nil
}
