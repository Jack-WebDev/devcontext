package context

import (
	"fmt"
	"time"

	codingtool "devctx/packages/core/codingtool"
	"devctx/packages/core/provider"
)

const (
	personalContextID = "personal"
	companyContextID  = "company"
)

// DefaultPersonalContext returns the built-in Personal context seed.
func DefaultPersonalContext(createdAt time.Time) Context {
	return DefaultPersonalContextWithProviderRegistry(createdAt, provider.BuiltInRegistry())
}

// DefaultCompanyContext returns the built-in Company context seed.
func DefaultCompanyContext(createdAt time.Time) Context {
	return DefaultCompanyContextWithProviderRegistry(createdAt, provider.BuiltInRegistry())
}

// DefaultPersonalContextWithProviderRegistry returns the built-in Personal
// context seed using the registry's default-enabled providers.
func DefaultPersonalContextWithProviderRegistry(createdAt time.Time, registry provider.Registry) Context {
	return defaultContextSeed(MustID(personalContextID), "Personal", createdAt, registry)
}

// DefaultCompanyContextWithProviderRegistry returns the built-in Company
// context seed using the registry's default-enabled providers.
func DefaultCompanyContextWithProviderRegistry(createdAt time.Time, registry provider.Registry) Context {
	return defaultContextSeed(MustID(companyContextID), "Company", createdAt, registry)
}

// DefaultContextForID returns the built-in context seed for a supported default
// context ID.
func DefaultContextForID(id ID, createdAt time.Time) (Context, error) {
	return DefaultContextForIDWithProviderRegistry(id, createdAt, provider.BuiltInRegistry())
}

// DefaultContextForIDWithProviderRegistry returns the built-in context seed for
// a supported default context ID using the registry's default-enabled providers.
func DefaultContextForIDWithProviderRegistry(id ID, createdAt time.Time, registry provider.Registry) (Context, error) {
	return DefaultContextForIDWithRegistries(id, createdAt, registry, codingtool.BuiltInRegistry())
}

// DefaultContextForIDWithRegistries returns a default context using the
// provider and editor defaults owned by their respective registries.
func DefaultContextForIDWithRegistries(id ID, createdAt time.Time, providerRegistry provider.Registry, editorRegistry codingtool.Registry) (Context, error) {
	switch id.String() {
	case personalContextID:
		return defaultContextSeed(MustID(personalContextID), "Personal", createdAt, providerRegistry, editorRegistry), nil
	case companyContextID:
		return defaultContextSeed(MustID(companyContextID), "Company", createdAt, providerRegistry, editorRegistry), nil
	default:
		return Context{}, fmt.Errorf("%w: unsupported default context %q", ErrContextNotFound, id.String())
	}
}

func defaultContextSeed(id ID, name string, createdAt time.Time, registry provider.Registry, editorRegistries ...codingtool.Registry) Context {
	editorRegistry := codingtool.BuiltInRegistry()
	if len(editorRegistries) > 0 && !editorRegistries[0].IsZero() {
		editorRegistry = editorRegistries[0]
	}
	return Context{
		ID:        id,
		Name:      name,
		Tool:      codingtool.DefaultConfigForRegistry(editorRegistry),
		Providers: registry.DefaultConfigs(),
		CreatedAt: createdAt.UTC(),
	}
}
