// Package project owns project identity and project-to-context bindings.
//
// Repository operations canonicalize paths before using them as binding keys.
// Bind validates both the project directory and target context, while lookup
// can additionally report a binding whose target context is missing.
package project
