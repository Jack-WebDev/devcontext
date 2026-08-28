package launcher

import (
	"net/mail"
	"strings"
)

// AccountIdentityEvidence is verified, presentation-safe identity metadata
// from one enabled provider. It intentionally contains no credential data.
type AccountIdentityEvidence struct {
	ProviderID string
	Fields     []AccountIdentityField
}

// AccountIdentityField is one provider-supplied, presentation-safe identity
// field. Only fields with a documented stable meaning can be compared.
type AccountIdentityField struct {
	Label string
	Value string
}

// AccountIdentityMismatchConfidenceCheck reports a launchable warning only
// when two or more enabled providers have verified, valid Email fields with
// different values. Missing identity data, unsupported fields, a single
// provider, and matching emails are unknown or consistent evidence, not a
// mismatch.
func AccountIdentityMismatchConfidenceCheck(evidence []AccountIdentityEvidence) (ConfidenceCheck, bool) {
	providerEmails := make(map[string]map[string]bool)
	for _, providerEvidence := range evidence {
		if strings.TrimSpace(providerEvidence.ProviderID) == "" {
			continue
		}
		for _, field := range providerEvidence.Fields {
			if normalizedEmail, ok := verifiedEmailField(field); ok {
				if providerEmails[providerEvidence.ProviderID] == nil {
					providerEmails[providerEvidence.ProviderID] = make(map[string]bool)
				}
				providerEmails[providerEvidence.ProviderID][normalizedEmail] = true
			}
		}
	}

	emails := make(map[string]bool)
	providersWithOneEmail := 0
	for _, values := range providerEmails {
		if len(values) != 1 {
			continue
		}
		providersWithOneEmail++
		for value := range values {
			emails[value] = true
		}
	}
	if providersWithOneEmail < 2 || len(emails) < 2 {
		return ConfidenceCheck{}, false
	}

	return ConfidenceCheck{
		Component:  ConfidenceCheckIdentity,
		Severity:   ConfidenceNeedsAttention,
		Label:      "Account identity",
		Message:    "Verified provider email identities do not match for this context.",
		ActionHint: "Review provider account configuration before launch.",
	}, true
}

func verifiedEmailField(field AccountIdentityField) (string, bool) {
	if !strings.EqualFold(strings.TrimSpace(field.Label), "email") {
		return "", false
	}
	value := strings.TrimSpace(field.Value)
	parsed, err := mail.ParseAddress(value)
	if err != nil || parsed.Address != value {
		return "", false
	}
	return strings.ToLower(parsed.Address), true
}
