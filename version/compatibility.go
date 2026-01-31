package version

import (
	"fmt"
)

// Compatibility represents version compatibility between components.
type Compatibility struct {
	Component  string
	MinVersion Version
	MaxVersion *Version // nil means no upper bound
	Deprecated bool
	Message    string
}

// Matrix holds compatibility information for multiple components.
type Matrix struct {
	entries map[string][]Compatibility
}

// NewMatrix creates a new compatibility matrix.
func NewMatrix() *Matrix {
	return &Matrix{
		entries: make(map[string][]Compatibility),
	}
}

// Add adds a compatibility entry for a component.
func (m *Matrix) Add(comp Compatibility) {
	m.entries[comp.Component] = append(m.entries[comp.Component], comp)
}

// Check checks if a version is compatible for a component.
func (m *Matrix) Check(component string, v Version) (bool, string) {
	entries, ok := m.entries[component]
	if !ok {
		return true, "" // unknown component, assume compatible
	}

	for _, entry := range entries {
		if v.Compare(entry.MinVersion) < 0 {
			return false, fmt.Sprintf("version %s is below minimum %s", v, entry.MinVersion)
		}
		if entry.MaxVersion != nil && v.Compare(*entry.MaxVersion) > 0 {
			return false, fmt.Sprintf("version %s exceeds maximum %s", v, entry.MaxVersion)
		}
		if entry.Deprecated {
			return true, entry.Message // compatible but deprecated
		}
	}

	return true, ""
}

// Negotiate finds the best compatible version from a list.
func (m *Matrix) Negotiate(component string, available []Version) (Version, error) {
	var best *Version

	for _, v := range available {
		compatible, _ := m.Check(component, v)
		if compatible {
			if best == nil || v.GreaterThan(*best) {
				vCopy := v
				best = &vCopy
			}
		}
	}

	if best == nil {
		return Version{}, fmt.Errorf("no compatible version found for %s", component)
	}

	return *best, nil
}
