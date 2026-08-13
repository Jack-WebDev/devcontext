// Package editor owns editor intent, executable detection, and command
// construction boundaries.
//
// It contains generic editor configuration and contracts without process
// creation behavior. VS Code-specific detection and launch command planning are
// added by later phases without introducing Wails dependencies.
package editor
