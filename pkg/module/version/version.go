package version

import (
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/StackCatalyst/common-lib/pkg/module"
)

// Manager handles version control operations for modules
type Manager interface {
	// Parse parses a version string into a Version object
	Parse(version string) (*Version, error)

	// Resolve resolves a version constraint to a specific version
	Resolve(constraint string, versions []string) (string, error)

	// Lock marks a module version as immutable
	Lock(module *module.Module) error

	// Verify verifies module integrity and signature
	Verify(module *module.Module) error

	// Compare compares two versions
	Compare(v1, v2 string) (int, error)

	// IsValid checks if a version string is valid
	IsValid(version string) bool

	// IsSatisfied checks if a version satisfies a constraint
	IsSatisfied(version, constraint string) (bool, error)
}

// Version represents a semantic version
type Version struct {
	*semver.Version
}

// DefaultManager is the default implementation of Manager
type DefaultManager struct{}

// NewManager creates a new version manager
func NewManager() Manager {
	return &DefaultManager{}
}

// Parse parses a version string into a Version object
func (m *DefaultManager) Parse(version string) (*Version, error) {
	v, err := semver.NewVersion(version)
	if err != nil {
		return nil, fmt.Errorf("invalid version format: %w", err)
	}
	return &Version{Version: v}, nil
}

// Resolve resolves a version constraint to a specific version
func (m *DefaultManager) Resolve(constraint string, versions []string) (string, error) {
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return "", fmt.Errorf("invalid constraint format: %w", err)
	}

	var semvers []*semver.Version
	for _, v := range versions {
		sv, err := semver.NewVersion(v)
		if err != nil {
			continue // Skip invalid versions
		}
		semvers = append(semvers, sv)
	}

	// Sort versions in descending order
	sort.Slice(semvers, func(i, j int) bool {
		return semvers[i].GreaterThan(semvers[j])
	})

	// Find the highest version that satisfies the constraint
	for _, sv := range semvers {
		if c.Check(sv) {
			return sv.String(), nil
		}
	}

	return "", fmt.Errorf("no version satisfies constraint %s", constraint)
}

// Lock marks a module version as immutable
func (m *DefaultManager) Lock(module *module.Module) error {
	if !m.IsValid(module.Version) {
		return fmt.Errorf("invalid version format: %s", module.Version)
	}
	// The actual locking is handled by the storage layer
	return nil
}

// Verify verifies module integrity and signature
func (m *DefaultManager) Verify(module *module.Module) error {
	if !m.IsValid(module.Version) {
		return fmt.Errorf("invalid version format: %s", module.Version)
	}

	// Verify dependencies
	for _, dep := range module.Dependencies {
		if !m.IsValid(dep.Version) {
			return fmt.Errorf("invalid dependency version format: %s", dep.Version)
		}
	}

	// TODO: Implement signature verification
	return nil
}

// Compare compares two versions
func (m *DefaultManager) Compare(v1, v2 string) (int, error) {
	sv1, err := semver.NewVersion(v1)
	if err != nil {
		return 0, fmt.Errorf("invalid version format for v1: %w", err)
	}

	sv2, err := semver.NewVersion(v2)
	if err != nil {
		return 0, fmt.Errorf("invalid version format for v2: %w", err)
	}

	return sv1.Compare(sv2), nil
}

// IsValid checks if a version string is valid
func (m *DefaultManager) IsValid(version string) bool {
	_, err := semver.NewVersion(version)
	return err == nil
}

// IsSatisfied checks if a version satisfies a constraint
func (m *DefaultManager) IsSatisfied(version, constraint string) (bool, error) {
	v, err := semver.NewVersion(version)
	if err != nil {
		return false, fmt.Errorf("invalid version format: %w", err)
	}

	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return false, fmt.Errorf("invalid constraint format: %w", err)
	}

	return c.Check(v), nil
}

// String returns the string representation of a Version
func (v *Version) String() string {
	return v.Version.String()
}

// IsPrerelease checks if the version is a prerelease
func (v *Version) IsPrerelease() bool {
	return v.Version.Prerelease() != ""
}

// Major returns the major version
func (v *Version) Major() uint64 {
	return v.Version.Major()
}

// Minor returns the minor version
func (v *Version) Minor() uint64 {
	return v.Version.Minor()
}

// Patch returns the patch version
func (v *Version) Patch() uint64 {
	return v.Version.Patch()
}

// Prerelease returns the prerelease version
func (v *Version) Prerelease() string {
	return v.Version.Prerelease()
}

// Metadata returns the metadata
func (v *Version) Metadata() string {
	return v.Version.Metadata()
}

// Original returns the original version string
func (v *Version) Original() string {
	return v.Version.Original()
}
