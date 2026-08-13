// Package environment owns process environment composition for isolated
// development contexts.
//
// It preserves the parent process environment before applying provider-owned
// context overrides. Future phases add active-context markers and redaction
// here.
package environment
