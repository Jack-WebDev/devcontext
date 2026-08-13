package logging

import (
	"regexp"
	"sort"
	"strings"

	"devctx/packages/core/environment"
)

const (
	// RedactedValue replaces sensitive values in persisted diagnostics.
	RedactedValue = environment.RedactedValue

	maxDiagnosticLength = 2048
)

var (
	sensitiveAssignmentPattern = regexp.MustCompile(`(?i)\b([A-Z0-9_.-]*(TOKEN|SECRET|PASSWORD|PASSWD|COOKIE|API[_-]?KEY|ACCESS[_-]?KEY|PRIVATE[_-]?KEY|CREDENTIAL)[A-Z0-9_.-]*\s*[:=]\s*)("[^"]*"|'[^']*'|[^\s,;&]+)`)
	authorizationPattern       = regexp.MustCompile(`(?i)\b(authorization\s*[:=]\s*)(bearer|basic)\s+[A-Za-z0-9._~+/=-]+`)
	cookiePattern              = regexp.MustCompile(`(?i)\b(cookie\s*[:=]\s*)[^\r\n]+`)
	oauthAssignmentPattern     = regexp.MustCompile(`(?i)\b((code|oauth_code|authorization_code)\s*[:=]\s*)("[^"]*"|'[^']*'|[^\s,;&]+)`)
	oauthQueryPattern          = regexp.MustCompile(`(?i)([?&](code|oauth_code|authorization_code|access_token|refresh_token|id_token)=)[^&\s]+`)
)

// SanitizeError removes likely credentials from error text.
func SanitizeError(err error, knownEnvironment []string) string {
	if err == nil {
		return ""
	}
	return SanitizeText(err.Error(), knownEnvironment)
}

// SanitizeText removes likely credentials from diagnostic text.
func SanitizeText(value string, knownEnvironment []string) string {
	sanitized := value
	for _, secret := range sensitiveEnvironmentValues(knownEnvironment) {
		sanitized = strings.ReplaceAll(sanitized, secret, RedactedValue)
	}

	sanitized = authorizationPattern.ReplaceAllString(sanitized, "${1}"+RedactedValue)
	sanitized = cookiePattern.ReplaceAllString(sanitized, "${1}"+RedactedValue)
	sanitized = oauthAssignmentPattern.ReplaceAllString(sanitized, "${1}"+RedactedValue)
	sanitized = oauthQueryPattern.ReplaceAllString(sanitized, "${1}"+RedactedValue)
	sanitized = sensitiveAssignmentPattern.ReplaceAllString(sanitized, "${1}"+RedactedValue)

	if len(sanitized) > maxDiagnosticLength {
		return sanitized[:maxDiagnosticLength] + "...<truncated>"
	}
	return sanitized
}

func sensitiveEnvironmentValues(entries []string) []string {
	valuesByContent := map[string]struct{}{}
	for _, entry := range entries {
		key, value, ok := strings.Cut(entry, "=")
		if !ok || key == "" || value == "" {
			continue
		}
		if environment.IsSensitiveKey(key) {
			valuesByContent[value] = struct{}{}
		}
	}

	values := make([]string, 0, len(valuesByContent))
	for value := range valuesByContent {
		values = append(values, value)
	}
	sort.Slice(values, func(i int, j int) bool {
		return len(values[i]) > len(values[j])
	})
	return values
}
