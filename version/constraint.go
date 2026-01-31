package version

import (
	"strings"
)

// Constraint represents a version constraint (e.g., ">=1.0.0", "^2.0.0").
type Constraint struct {
	Op      string  // "", "=", ">", ">=", "<", "<=", "^", "~"
	Version Version
}

// ParseConstraint parses a version constraint string.
func ParseConstraint(s string) (Constraint, error) {
	s = strings.TrimSpace(s)

	var op string
	var versionStr string

	switch {
	case strings.HasPrefix(s, ">="):
		op = ">="
		versionStr = strings.TrimPrefix(s, ">=")
	case strings.HasPrefix(s, "<="):
		op = "<="
		versionStr = strings.TrimPrefix(s, "<=")
	case strings.HasPrefix(s, ">"):
		op = ">"
		versionStr = strings.TrimPrefix(s, ">")
	case strings.HasPrefix(s, "<"):
		op = "<"
		versionStr = strings.TrimPrefix(s, "<")
	case strings.HasPrefix(s, "^"):
		op = "^"
		versionStr = strings.TrimPrefix(s, "^")
	case strings.HasPrefix(s, "~"):
		op = "~"
		versionStr = strings.TrimPrefix(s, "~")
	case strings.HasPrefix(s, "="):
		op = "="
		versionStr = strings.TrimPrefix(s, "=")
	default:
		op = "="
		versionStr = s
	}

	v, err := Parse(strings.TrimSpace(versionStr))
	if err != nil {
		return Constraint{}, err
	}

	return Constraint{Op: op, Version: v}, nil
}

// Check returns true if the given version satisfies the constraint.
func (c Constraint) Check(v Version) bool {
	switch c.Op {
	case "", "=":
		return v.Equal(c.Version)
	case ">":
		return v.GreaterThan(c.Version)
	case ">=":
		return v.GreaterThan(c.Version) || v.Equal(c.Version)
	case "<":
		return v.LessThan(c.Version)
	case "<=":
		return v.LessThan(c.Version) || v.Equal(c.Version)
	case "^":
		// Caret: compatible with (same major, >= version)
		return v.Major == c.Version.Major && (v.GreaterThan(c.Version) || v.Equal(c.Version))
	case "~":
		// Tilde: same major.minor, >= version
		return v.Major == c.Version.Major && v.Minor == c.Version.Minor &&
			(v.GreaterThan(c.Version) || v.Equal(c.Version))
	default:
		return false
	}
}

// String returns the constraint as a string.
func (c Constraint) String() string {
	if c.Op == "" || c.Op == "=" {
		return c.Version.String()
	}
	return c.Op + c.Version.String()
}
