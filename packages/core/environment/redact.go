package environment

import "strings"

const (
	// RedactedValue replaces sensitive environment values in diagnostics.
	RedactedValue = "<redacted>"
)

var sensitiveKeyFragments = []string{
	"TOKEN",
	"SECRET",
	"PASSWORD",
	"PASSWD",
	"COOKIE",
	"AUTHORIZATION",
	"API_KEY",
	"ACCESS_KEY",
	"PRIVATE_KEY",
	"CREDENTIAL",
}

// Redacted returns a copy of the variables with sensitive values masked.
func (v Variables) Redacted() Variables {
	redacted := make(Variables, len(v))
	for key, value := range v {
		if IsSensitiveKey(key) {
			redacted[key] = RedactedValue
			continue
		}
		redacted[key] = value
	}
	return redacted
}

// RedactedEnviron returns deterministic KEY=value entries with sensitive values
// masked.
func (v Variables) RedactedEnviron() []string {
	return v.Redacted().Environ()
}

// IsSensitiveKey reports whether a variable name commonly carries credentials
// or authorization data.
func IsSensitiveKey(key string) bool {
	normalized := strings.ToUpper(key)
	for _, fragment := range sensitiveKeyFragments {
		if strings.Contains(normalized, fragment) {
			return true
		}
	}

	for _, segment := range splitKeySegments(normalized) {
		if segment == "AUTH" {
			return true
		}
	}
	return false
}

func splitKeySegments(key string) []string {
	return strings.FieldsFunc(key, func(r rune) bool {
		return r == '_' || r == '-' || r == '.'
	})
}
