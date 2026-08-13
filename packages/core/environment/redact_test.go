package environment_test

import (
	"reflect"
	"strings"
	"testing"

	"devctx/packages/core/environment"
)

func TestRedactedMasksSensitiveEnvironmentValues(t *testing.T) {
	variables := environment.Variables{
		"GITHUB_TOKEN":      "raw-token",
		"CLIENT_SECRET":     "raw-secret",
		"DATABASE_PASSWORD": "raw-password",
		"SESSION_COOKIE":    "raw-cookie",
		"AUTHORIZATION":     "Bearer raw-authorization",
		"API_KEY":           "raw-api-key",
		"BASIC_AUTH":        "raw-basic-auth",
		"PATH":              "/usr/bin",
	}

	redacted := variables.Redacted()

	for _, key := range []string{
		"GITHUB_TOKEN",
		"CLIENT_SECRET",
		"DATABASE_PASSWORD",
		"SESSION_COOKIE",
		"AUTHORIZATION",
		"API_KEY",
		"BASIC_AUTH",
	} {
		if redacted[key] != environment.RedactedValue {
			t.Fatalf("redacted[%q] = %q, want %q", key, redacted[key], environment.RedactedValue)
		}
	}
	if redacted["PATH"] != "/usr/bin" {
		t.Fatalf("redacted PATH = %q, want %q", redacted["PATH"], "/usr/bin")
	}
	if variables["GITHUB_TOKEN"] != "raw-token" {
		t.Fatalf("original value changed to %q", variables["GITHUB_TOKEN"])
	}

	output := strings.Join(redacted.Environ(), "\n")
	for _, rawValue := range []string{
		"raw-token",
		"raw-secret",
		"raw-password",
		"raw-cookie",
		"raw-authorization",
		"raw-api-key",
		"raw-basic-auth",
	} {
		if strings.Contains(output, rawValue) {
			t.Fatalf("redacted output contains raw value %q: %s", rawValue, output)
		}
	}
}

func TestRedactedEnvironReturnsDeterministicMaskedEntries(t *testing.T) {
	variables := environment.Variables{
		"SECRET_TOKEN": "raw-token",
		"PATH":         "/usr/bin",
	}

	got := variables.RedactedEnviron()
	want := []string{
		"PATH=/usr/bin",
		"SECRET_TOKEN=<redacted>",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("redacted environ = %#v, want %#v", got, want)
	}
}
