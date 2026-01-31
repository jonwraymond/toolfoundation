// Package version provides semantic versioning utilities for the ApertureStack ecosystem.
//
// This package handles version parsing, comparison, compatibility checking, and
// version negotiation for tool and protocol versions.
//
// # Parsing Versions
//
// Parse semantic version strings:
//
//	v, err := version.Parse("1.2.3")
//	v, err := version.Parse("v2.0.0-beta.1+build.123")
//
// # Comparing Versions
//
//	v1 := version.MustParse("1.0.0")
//	v2 := version.MustParse("2.0.0")
//
//	v1.LessThan(v2)    // true
//	v1.GreaterThan(v2) // false
//	v1.Equal(v2)       // false
//	v1.Compatible(v2)  // false (different major)
//
// # Version Constraints
//
// Parse and check version constraints:
//
//	c, _ := version.ParseConstraint(">=1.0.0")
//	c.Check(version.MustParse("1.5.0")) // true
//	c.Check(version.MustParse("0.9.0")) // false
//
// Supported constraint operators:
//   - "=" or "" - exact match
//   - ">" - greater than
//   - ">=" - greater than or equal
//   - "<" - less than
//   - "<=" - less than or equal
//   - "^" - compatible (same major)
//   - "~" - approximately (same major.minor)
//
// # Compatibility Matrix
//
// Track version compatibility across components:
//
//	matrix := version.NewMatrix()
//	matrix.Add(version.Compatibility{
//	    Component:  "toolfoundation",
//	    MinVersion: version.MustParse("0.1.0"),
//	})
//
//	ok, msg := matrix.Check("toolfoundation", version.MustParse("0.2.0"))
//
// # Version Negotiation
//
// Find the best compatible version from available options:
//
//	available := []version.Version{
//	    version.MustParse("1.0.0"),
//	    version.MustParse("1.1.0"),
//	    version.MustParse("2.0.0"),
//	}
//	best, err := matrix.Negotiate("component", available)
package version
