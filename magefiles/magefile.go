//go:build mage
package main

import (
	"fmt"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Lint runs golangci-lint with the project configuration
func Lint() error {
	fmt.Println("Running golangci-lint...")
	return sh.RunV("golangci-lint", "run")
}

// Test runs all tests in the repository
func Test() error {
	fmt.Println("Running go test ./...")
	return sh.RunV("go", "test", "-v", "./...")
}

// Check performs both linting and testing
func Check() {
	mg.Deps(Lint, Test)
}

// Release creates a new release using GoReleaser
// It requires a GITHUB_TOKEN environment variable and a git tag.
func Release() error {
	fmt.Println("Running GoReleaser...")
	return sh.RunV("goreleaser", "release", "--clean")
}

// Snapshot creates a temporary release without publishing
func Snapshot() error {
	fmt.Println("Running GoReleaser snapshot...")
	return sh.RunV("goreleaser", "release", "--snapshot", "--clean")
}
