# Project Overview

## Purpose
This is a Bubble Tea application template repository for creating terminal user interface (TUI) applications in Go. It's a starting point template that includes a sample app with spinner animation and basic keyboard controls.

## Tech Stack
- **Language**: Go 1.23.0
- **Main Framework**: Bubble Tea (charmbracelet/bubbletea v1.3.6) - TUI framework
- **UI Components**: Bubbles (charmbracelet/bubbles v0.21.0) - pre-built components
- **Styling**: Lip Gloss (charmbracelet/lipgloss v1.1.0) - styling for terminal apps

## Project Structure
- `main.go` - Entry point with a simple spinner app demonstrating Bubble Tea patterns
- `go.mod/go.sum` - Go module files with dependencies
- `.golangci.yml` - Linting configuration
- `.goreleaser.yaml` - Release automation configuration
- `.github/workflows/` - CI/CD workflows for build, test, lint, and release