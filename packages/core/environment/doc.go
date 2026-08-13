// Package environment owns process environment composition for isolated
// development contexts.
//
// It preserves the parent process environment before applying provider-owned
// context overrides, marks the selected context, and provides a redacted view
// for diagnostics.
package environment
