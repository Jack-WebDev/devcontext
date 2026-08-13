package context

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidID identifies context IDs that are unsafe or ambiguous for
	// storage paths and commands.
	ErrInvalidID = errors.New("invalid context ID")
)

// ID identifies a development context.
type ID struct {
	value string
}

// NewID validates and returns a context ID.
func NewID(value string) (ID, error) {
	if value == "" {
		return ID{}, fmt.Errorf("%w: cannot be empty", ErrInvalidID)
	}
	if value[0] == '-' || value[len(value)-1] == '-' {
		return ID{}, fmt.Errorf("%w: %q must start and end with a lowercase letter or digit", ErrInvalidID, value)
	}

	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '-':
		default:
			return ID{}, fmt.Errorf("%w: %q must contain only lowercase letters, digits, and hyphens", ErrInvalidID, value)
		}
	}

	return ID{value: value}, nil
}

// MustID validates and returns a context ID, panicking if the value is invalid.
//
// This helper is intended for tests and static seed data. Runtime input should
// use NewID and handle the returned error.
func MustID(value string) ID {
	id, err := NewID(value)
	if err != nil {
		panic(err)
	}
	return id
}

// String returns the ID value used in paths, commands, and persisted config.
func (id ID) String() string {
	return id.value
}
