package version

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Version represents a semantic version (major.minor.patch-prerelease+build).
type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Build      string
}

var semverRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-.]+))?(?:\+([0-9A-Za-z-.]+))?$`)

// Parse parses a semantic version string.
func Parse(s string) (Version, error) {
	matches := semverRegex.FindStringSubmatch(s)
	if matches == nil {
		return Version{}, fmt.Errorf("invalid semantic version: %s", s)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	return Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: matches[4],
		Build:      matches[5],
	}, nil
}

// MustParse parses a version string and panics on error.
func MustParse(s string) Version {
	v, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return v
}

// String returns the version as a string (with v prefix).
func (v Version) String() string {
	s := fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Prerelease != "" {
		s += "-" + v.Prerelease
	}
	if v.Build != "" {
		s += "+" + v.Build
	}
	return s
}

// Compare returns -1, 0, or 1 if v < other, v == other, or v > other.
func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		return compareInt(v.Major, other.Major)
	}
	if v.Minor != other.Minor {
		return compareInt(v.Minor, other.Minor)
	}
	if v.Patch != other.Patch {
		return compareInt(v.Patch, other.Patch)
	}
	return comparePrerelease(v.Prerelease, other.Prerelease)
}

// LessThan returns true if v < other.
func (v Version) LessThan(other Version) bool {
	return v.Compare(other) < 0
}

// GreaterThan returns true if v > other.
func (v Version) GreaterThan(other Version) bool {
	return v.Compare(other) > 0
}

// Equal returns true if v == other (ignoring build metadata).
func (v Version) Equal(other Version) bool {
	return v.Compare(other) == 0
}

// Compatible returns true if v is compatible with other (same major, v >= other).
func (v Version) Compatible(other Version) bool {
	if v.Major != other.Major {
		return false
	}
	return v.Compare(other) >= 0
}

func compareInt(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func comparePrerelease(a, b string) int {
	if a == "" && b == "" {
		return 0
	}
	if a == "" {
		return 1 // no prerelease > prerelease
	}
	if b == "" {
		return -1
	}
	return strings.Compare(a, b)
}
