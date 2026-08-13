package context

import (
	"time"

	"devctx/packages/core/editor"
	"devctx/packages/core/provider"
)

const (
	personalContextID = "personal"
	companyContextID  = "company"
)

// DefaultPersonalContext returns the built-in Personal context seed.
func DefaultPersonalContext(createdAt time.Time) Context {
	return defaultContextSeed(MustID(personalContextID), "Personal", createdAt)
}

// DefaultCompanyContext returns the built-in Company context seed.
func DefaultCompanyContext(createdAt time.Time) Context {
	return defaultContextSeed(MustID(companyContextID), "Company", createdAt)
}

func defaultContextSeed(id ID, name string, createdAt time.Time) Context {
	return Context{
		ID:     id,
		Name:   name,
		Editor: editor.DefaultConfig(),
		Providers: provider.Configs{
			provider.ClaudeID: {Enabled: true},
			provider.CodexID:  {Enabled: true},
		},
		CreatedAt: createdAt.UTC(),
	}
}
