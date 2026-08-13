// Package provider owns AI and developer-tool provider abstractions.
//
// It contains generic provider configuration without Claude, Codex, or future
// provider-specific branches. Provider implementations contribute environment
// variables and local status through a small interface instead of core type
// switches. Future phases add bounded provider status states here.
package provider
