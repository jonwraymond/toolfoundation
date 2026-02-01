package version

import (
	"testing"
)

// FuzzParse tests the version parser with random inputs.
// Run with: go test -fuzz=FuzzParse -fuzztime=30s ./version/...
func FuzzParse(f *testing.F) {
	// Seed corpus with valid versions
	f.Add("1.0.0")
	f.Add("v1.0.0")
	f.Add("1.2.3")
	f.Add("1.0.0-alpha")
	f.Add("1.0.0-alpha.1")
	f.Add("1.0.0+build")
	f.Add("1.0.0-beta+build.123")
	f.Add("0.0.0")
	f.Add("999.999.999")

	// Seed with invalid versions
	f.Add("")
	f.Add("invalid")
	f.Add("1.0")
	f.Add("1")
	f.Add("v")
	f.Add("1.0.0.0")
	f.Add("-1.0.0")
	f.Add("1.-1.0")
	f.Add("1.0.-1")

	// Edge cases
	f.Add("v0.0.0-0+0")
	f.Add("1.0.0-alpha-beta")
	f.Add("1.0.0+build-info")

	f.Fuzz(func(t *testing.T, input string) {
		// Parse should not panic on any input
		v, err := Parse(input)
		if err != nil {
			// Invalid input is expected, just ensure no panic
			return
		}

		// If parsing succeeded, the version should be valid
		// String() should produce a parseable version
		str := v.String()
		v2, err := Parse(str)
		if err != nil {
			t.Errorf("Parse(%q) succeeded but String() produced unparseable %q: %v", input, str, err)
		}

		// Round-trip should preserve semantics (ignoring the leading 'v')
		if !v.Equal(v2) {
			t.Errorf("Round-trip failed: Parse(%q)=%v, String()=%q, re-parsed=%v", input, v, str, v2)
		}
	})
}

// FuzzParseConstraint tests the constraint parser with random inputs.
// Run with: go test -fuzz=FuzzParseConstraint -fuzztime=30s ./version/...
func FuzzParseConstraint(f *testing.F) {
	// Seed corpus with valid constraints
	f.Add("1.0.0")
	f.Add("=1.0.0")
	f.Add(">1.0.0")
	f.Add(">=1.0.0")
	f.Add("<2.0.0")
	f.Add("<=2.0.0")
	f.Add("^1.0.0")
	f.Add("~1.0.0")
	f.Add(">=1.0.0-alpha")

	// Invalid constraints
	f.Add("")
	f.Add("invalid")
	f.Add(">>1.0.0")
	f.Add("!1.0.0")
	f.Add("1.0")

	f.Fuzz(func(t *testing.T, input string) {
		// ParseConstraint should not panic on any input
		c, err := ParseConstraint(input)
		if err != nil {
			// Invalid input is expected, just ensure no panic
			return
		}

		// If parsing succeeded, String() should work
		str := c.String()
		if str == "" {
			t.Errorf("ParseConstraint(%q) succeeded but String() returned empty", input)
		}

		// Check() should not panic with any version
		testVersions := []Version{
			{0, 0, 0, "", ""},
			{1, 0, 0, "", ""},
			{1, 0, 0, "alpha", ""},
			{999, 999, 999, "", ""},
		}
		for _, v := range testVersions {
			// Just ensure no panic
			_ = c.Check(v)
		}
	})
}

// FuzzVersionCompare tests version comparison with random versions.
// Run with: go test -fuzz=FuzzVersionCompare -fuzztime=30s ./version/...
func FuzzVersionCompare(f *testing.F) {
	// Seed with version pairs
	f.Add("1.0.0", "1.0.0")
	f.Add("1.0.0", "2.0.0")
	f.Add("1.0.0-alpha", "1.0.0")
	f.Add("1.0.0", "1.0.0-alpha")

	f.Fuzz(func(t *testing.T, a, b string) {
		va, errA := Parse(a)
		vb, errB := Parse(b)

		if errA != nil || errB != nil {
			return // Skip invalid versions
		}

		// Compare should be consistent
		cmp := va.Compare(vb)
		cmpReverse := vb.Compare(va)

		// a.Compare(b) should be -b.Compare(a)
		if cmp != -cmpReverse {
			t.Errorf("Compare inconsistent: %s.Compare(%s)=%d, but reverse=%d", a, b, cmp, cmpReverse)
		}

		// Equal should be consistent with Compare
		if va.Equal(vb) != (cmp == 0) {
			t.Errorf("Equal inconsistent with Compare for %s vs %s", a, b)
		}

		// LessThan/GreaterThan should be consistent
		if va.LessThan(vb) != (cmp < 0) {
			t.Errorf("LessThan inconsistent with Compare for %s vs %s", a, b)
		}
		if va.GreaterThan(vb) != (cmp > 0) {
			t.Errorf("GreaterThan inconsistent with Compare for %s vs %s", a, b)
		}
	})
}
