package context_test

import (
	"errors"
	"strings"
	"testing"

	devcontext "devctx/packages/core/context"
)

func TestNewIDAcceptsFilesystemSafeLowercaseValues(t *testing.T) {
	tests := []string{
		"personal",
		"company",
		"client-a",
		"client-123",
		"client--a",
		"a1",
	}

	for _, value := range tests {
		t.Run(value, func(t *testing.T) {
			id, err := devcontext.NewID(value)
			if err != nil {
				t.Fatalf("new ID: %v", err)
			}
			if id.String() != value {
				t.Fatalf("ID string = %q, want %q", id.String(), value)
			}
		})
	}
}

func TestNewIDRejectsUnsafeOrAmbiguousValues(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		wantMessage string
	}{
		{
			name:        "empty",
			value:       "",
			wantMessage: "cannot be empty",
		},
		{
			name:        "leading whitespace",
			value:       " personal",
			wantMessage: "must contain only lowercase letters, digits, and hyphens",
		},
		{
			name:        "trailing whitespace",
			value:       "personal ",
			wantMessage: "must contain only lowercase letters, digits, and hyphens",
		},
		{
			name:        "internal whitespace",
			value:       "client a",
			wantMessage: "must contain only lowercase letters, digits, and hyphens",
		},
		{
			name:        "uppercase",
			value:       "Personal",
			wantMessage: "must contain only lowercase letters, digits, and hyphens",
		},
		{
			name:        "parent traversal",
			value:       "..",
			wantMessage: "must contain only lowercase letters, digits, and hyphens",
		},
		{
			name:        "nested path",
			value:       "client/a",
			wantMessage: "must contain only lowercase letters, digits, and hyphens",
		},
		{
			name:        "windows separator",
			value:       `client\a`,
			wantMessage: "must contain only lowercase letters, digits, and hyphens",
		},
		{
			name:        "leading hyphen",
			value:       "-client",
			wantMessage: "must start and end with a lowercase letter or digit",
		},
		{
			name:        "trailing hyphen",
			value:       "client-",
			wantMessage: "must start and end with a lowercase letter or digit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := devcontext.NewID(tt.value)
			if !errors.Is(err, devcontext.ErrInvalidID) {
				t.Fatalf("error = %v, want %v", err, devcontext.ErrInvalidID)
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Fatalf("error = %q, want message containing %q", err.Error(), tt.wantMessage)
			}
		})
	}
}

func TestMustIDPanicsForInvalidValue(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("MustID did not panic")
		}
	}()

	devcontext.MustID("Personal")
}
